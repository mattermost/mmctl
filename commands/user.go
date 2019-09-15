package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Management of users",
}

var UserDeactivateCmd = &cobra.Command{
	Use:   "deactivate [emails, usernames, userIds]",
	Short: "Deactivate users",
	Long:  "Deactivate users. Deactivated users are immediately logged out of all sessions and are unable to log back in.",
	Example: `  user deactivate user@example.com
  user deactivate username`,
	RunE: userDeactivateCmdF,
}

var UserCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a user",
	Long:    "Create a user",
	Example: `  user create --email user@example.com --username userexample --password Password1`,
	RunE:    userCreateCmdF,
}

var UserInviteCmd = &cobra.Command{
	Use:   "invite [email] [teams]",
	Short: "Send user an email invite to a team.",
	Long: `Send user an email invite to a team.
You can invite a user to multiple teams by listing them.
You can specify teams by name or ID.`,
	Example: `  user invite user@example.com myteam
  user invite user@example.com myteam1 myteam2`,
	RunE: userInviteCmdF,
}

var SendPasswordResetEmailCmd = &cobra.Command{
	Use:     "reset_password [users]",
	Short:   "Send users an email to reset their password",
	Long:    "Send users an email to reset their password",
	Example: "  user reset_password user@example.com",
	RunE:    sendPasswordResetEmailCmdF,
}

var updateUserEmailCmd = &cobra.Command{
	Use:   "email [user] [new email]",
	Short: "Change email of the user",
	Long:  "Change email of the user.",
	Example: `  user email test user@example.com
  user activate username`,
	RunE: updateUserEmailCmdF,
}

var ResetUserMfaCmd = &cobra.Command{
	Use:   "resetmfa [users]",
	Short: "Turn off MFA",
	Long: `Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they login.`,
	Example: "  user resetmfa user@example.com",
	RunE:    resetUserMfaCmdF,
}

var SearchUserCmd = &cobra.Command{
	Use:     "search [users]",
	Short:   "Search for users",
	Long:    "Search for users based on username, email, or user ID.",
	Example: "  user search user1@mail.com user2@mail.com",
	RunE:    searchUserCmdF,
}

func init() {
	UserCreateCmd.Flags().String("username", "", "Required. Username for the new user account.")
	UserCreateCmd.MarkFlagRequired("username")
	UserCreateCmd.Flags().String("email", "", "Required. The email address for the new user account.")
	UserCreateCmd.MarkFlagRequired("email")
	UserCreateCmd.Flags().String("password", "", "Required. The password for the new user account.")
	UserCreateCmd.MarkFlagRequired("password")
	UserCreateCmd.Flags().String("nickname", "", "Optional. The nickname for the new user account.")
	UserCreateCmd.Flags().String("firstname", "", "Optional. The first name for the new user account.")
	UserCreateCmd.Flags().String("lastname", "", "Optional. The last name for the new user account.")
	UserCreateCmd.Flags().String("locale", "", "Optional. The locale (ex: en, fr) for the new user account.")
	UserCreateCmd.Flags().Bool("system_admin", false, "Optional. If supplied, the new user will be a system administrator. Defaults to false.")

	UserCmd.AddCommand(
		UserDeactivateCmd,
		UserCreateCmd,
		UserInviteCmd,
		SendPasswordResetEmailCmd,
		updateUserEmailCmd,
		ResetUserMfaCmd,
		SearchUserCmd,
	)

	RootCmd.AddCommand(UserCmd)
}

func deactivateUsers(c *model.Client4, userArgs []string) {
	users := getUsersFromUserArgs(c, userArgs)
	for i, user := range users {
		if user.IsSSOUser() {
			fmt.Println("You must also deactivate user " + userArgs[i] + " in the SSO provider or they will be reactivated on next login or sync.")
		}

		if _, response := c.DeleteUser(user.Id); response.Error != nil {
			CommandPrintErrorln("Unable to deactivate user " + userArgs[i] + ". Error: " + response.Error.Error())
		}
	}
}

func userDeactivateCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	deactivateUsers(c, args)

	return nil
}

func userCreateCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	username, erru := command.Flags().GetString("username")
	if erru != nil {
		return errors.Wrap(erru, "Username is required")
	}
	email, erre := command.Flags().GetString("email")
	if erre != nil {
		return errors.Wrap(erre, "Email is required")
	}
	password, errp := command.Flags().GetString("password")
	if errp != nil {
		return errors.Wrap(errp, "Password is required")
	}
	nickname, _ := command.Flags().GetString("nickname")
	firstname, _ := command.Flags().GetString("firstname")
	lastname, _ := command.Flags().GetString("lastname")
	locale, _ := command.Flags().GetString("locale")
	systemAdmin, _ := command.Flags().GetBool("system_admin")

	user := &model.User{
		Username:  username,
		Email:     email,
		Password:  password,
		Nickname:  nickname,
		FirstName: firstname,
		LastName:  lastname,
		Locale:    locale,
	}

	ruser, response := c.CreateUser(user)

	if response.Error != nil {
		return errors.New("Unable to create user. Error: " + response.Error.Error())
	}

	if systemAdmin {
		if _, response := c.UpdateUserRoles(ruser.Id, "system_user system_admin"); response.Error != nil {
			return errors.New("Unable to update user roles. Error: " + response.Error.Error())
		}
	}

	Log.PrintT("Created user {{.Username}}", ruser)

	return nil
}

func userInviteCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.New("Expected at least two arguments. See help text for details.")
	}

	email := args[0]
	if !model.IsValidEmail(email) {
		return errors.New("Invalid email")
	}

	teams := getTeamsFromTeamArgs(c, args[1:])
	for i, team := range teams {
		err := inviteUser(c, email, team, args[i+1])

		if err != nil {
			CommandPrintErrorln(err.Error())
		}
	}

	return nil
}

func inviteUser(c *model.Client4, email string, team *model.Team, teamArg string) error {
	invites := []string{email}
	if team == nil {
		return fmt.Errorf("Can't find team '%v'", teamArg)
	}

	if _, response := c.InviteUsersToTeam(team.Id, invites); response.Error != nil {
		return errors.New("Unable to invite user with email " + email + " to team " + team.Name + ". Error: " + response.Error.Error())
	}

	CommandPrettyPrintln("Invites may or may not have been sent.")

	return nil
}

func sendPasswordResetEmailCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for _, email := range args {
		if !model.IsValidEmail(email) {
			CommandPrintErrorln("Invalid email '" + email + "'")
			continue
		}
		if _, response := c.SendPasswordResetEmail(email); response.Error != nil {
			CommandPrintErrorln("Unable send reset password email to email " + email + ". Error: " + response.Error.Error())
		}
	}

	return nil
}

func updateUserEmailCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) != 2 {
		return errors.New("Expected two arguments. See help text for details.")
	}

	newEmail := args[1]

	if !model.IsValidEmail(newEmail) {
		return errors.New("Invalid email: '" + newEmail + "'")
	}

	if len(args) != 2 {
		return errors.New("Expected two arguments. See help text for details.")
	}

	user := getUserFromUserArg(c, args[0])
	if user == nil {
		return errors.New("Unable to find user '" + args[0] + "'")
	}

	user.Email = newEmail

	ruser, response := c.UpdateUser(user)
	if response.Error != nil {
		return errors.New(response.Error.Error())
	}

	Log.PrintT("User {{.Username}} updated successfully", ruser)

	return nil
}

func resetUserMfaCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	users := getUsersFromUserArgs(c, args)

	for i, user := range users {
		if user == nil {
			CommandPrintErrorln("Unable to find user '" + args[i] + "'")
			continue
		}
		if _, response := c.UpdateUserMfa(user.Id, "", false); response.Error != nil {
			CommandPrintErrorln("Unable to reset user '" + args[i] + "' MFA. Error: " + response.Error.Error())
		}
	}

	return nil
}

func searchUserCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	users := getUsersFromUserArgs(c, args)

	for i, user := range users {
		tpl := `id: {{.Id}}
username: {{.Username}}
nickname: {{.Nickname}}
position: {{.Position}}
first_name: {{.FirstName}}
last_name: {{.LastName}}
email: {{.Email}}
auth_service: {{.AuthService}}`
		if i > 0 {
			tpl = "------------------------------\n" + tpl
		}
		if user == nil {
			CommandPrintErrorln("Unable to find user '" + args[i] + "'")
			continue
		}

		Log.PrintT(tpl, user)
	}

	return nil
}

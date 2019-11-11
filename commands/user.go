package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

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
	RunE: withClient(userDeactivateCmdF),
}

var UserCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a user",
	Long:    "Create a user",
	Example: `  user create --email user@example.com --username userexample --password Password1`,
	RunE:    withClient(userCreateCmdF),
}

var UserInviteCmd = &cobra.Command{
	Use:   "invite [email] [teams]",
	Short: "Send user an email invite to a team.",
	Long: `Send user an email invite to a team.
You can invite a user to multiple teams by listing them.
You can specify teams by name or ID.`,
	Example: `  user invite user@example.com myteam
  user invite user@example.com myteam1 myteam2`,
	RunE: withClient(userInviteCmdF),
}

var SendPasswordResetEmailCmd = &cobra.Command{
	Use:     "reset_password [users]",
	Short:   "Send users an email to reset their password",
	Long:    "Send users an email to reset their password",
	Example: "  user reset_password user@example.com",
	RunE:    withClient(sendPasswordResetEmailCmdF),
}

var updateUserEmailCmd = &cobra.Command{
	Use:   "email [user] [new email]",
	Short: "Change email of the user",
	Long:  "Change email of the user.",
	Example: `  user email test user@example.com
  user activate username`,
	RunE: withClient(updateUserEmailCmdF),
}

var ResetUserMfaCmd = &cobra.Command{
	Use:   "resetmfa [users]",
	Short: "Turn off MFA",
	Long: `Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they login.`,
	Example: "  user resetmfa user@example.com",
	RunE:    withClient(resetUserMfaCmdF),
}

var SearchUserCmd = &cobra.Command{
	Use:     "search [users]",
	Short:   "Search for users",
	Long:    "Search for users based on username, email, or user ID.",
	Example: "  user search user1@mail.com user2@mail.com",
	RunE:    withClient(searchUserCmdF),
}

var ListUsersCmd = &cobra.Command{
	Use:     "list",
	Short:   "List users",
	Long:    "List all the users",
	Example: "  user list",
	RunE:    withClient(listUsersCmdF),
	Args:    cobra.NoArgs,
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

	ListUsersCmd.Flags().Int("page", 0, "Start page for list of users")
	ListUsersCmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	ListUsersCmd.Flags().Bool("all", false, "Fetch all users. Will ignore --page and --per-page")

	UserCmd.AddCommand(
		UserDeactivateCmd,
		UserCreateCmd,
		UserInviteCmd,
		SendPasswordResetEmailCmd,
		updateUserEmailCmd,
		ResetUserMfaCmd,
		SearchUserCmd,
		ListUsersCmd,
	)

	RootCmd.AddCommand(UserCmd)
}

func deactivateUsers(c client.Client, userArgs []string) {
	users := getUsersFromUserArgs(c, userArgs)
	for i, user := range users {
		if user.IsSSOUser() {
			printer.Print("You must also deactivate user " + userArgs[i] + " in the SSO provider or they will be reactivated on next login or sync.")
		}

		if _, response := c.DeleteUser(user.Id); response.Error != nil {
			printer.PrintError("Unable to deactivate user " + userArgs[i] + ". Error: " + response.Error.Error())
		}
	}
}

func userDeactivateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	deactivateUsers(c, args)

	return nil
}

func userCreateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	username, erru := cmd.Flags().GetString("username")
	if erru != nil {
		return errors.Wrap(erru, "Username is required")
	}
	email, erre := cmd.Flags().GetString("email")
	if erre != nil {
		return errors.Wrap(erre, "Email is required")
	}
	password, errp := cmd.Flags().GetString("password")
	if errp != nil {
		return errors.Wrap(errp, "Password is required")
	}
	nickname, _ := cmd.Flags().GetString("nickname")
	firstname, _ := cmd.Flags().GetString("firstname")
	lastname, _ := cmd.Flags().GetString("lastname")
	locale, _ := cmd.Flags().GetString("locale")
	systemAdmin, _ := cmd.Flags().GetBool("system_admin")

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

	printer.PrintT("Created user {{.Username}}", ruser)

	return nil
}

func userInviteCmdF(c client.Client, cmd *cobra.Command, args []string) error {
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
			printer.PrintError(err.Error())
		}
	}

	return nil
}

func inviteUser(c client.Client, email string, team *model.Team, teamArg string) error {
	invites := []string{email}
	if team == nil {
		return fmt.Errorf("Can't find team '%v'", teamArg)
	}

	if _, response := c.InviteUsersToTeam(team.Id, invites); response.Error != nil {
		return errors.New("Unable to invite user with email " + email + " to team " + team.Name + ". Error: " + response.Error.Error())
	}

	printer.Print("Invites may or may not have been sent.")

	return nil
}

func sendPasswordResetEmailCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for _, email := range args {
		if !model.IsValidEmail(email) {
			printer.PrintError("Invalid email '" + email + "'")
			continue
		}
		if _, response := c.SendPasswordResetEmail(email); response.Error != nil {
			printer.PrintError("Unable send reset password email to email " + email + ". Error: " + response.Error.Error())
		}
	}

	return nil
}

func updateUserEmailCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

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

	printer.PrintT("User {{.Username}} updated successfully", ruser)

	return nil
}

func resetUserMfaCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	users := getUsersFromUserArgs(c, args)

	for i, user := range users {
		if user == nil {
			printer.PrintError("Unable to find user '" + args[i] + "'")
			continue
		}
		if _, response := c.UpdateUserMfa(user.Id, "", false); response.Error != nil {
			printer.PrintError("Unable to reset user '" + args[i] + "' MFA. Error: " + response.Error.Error())
		}
	}

	return nil
}

func searchUserCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

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
			printer.PrintError("Unable to find user '" + args[i] + "'")
			continue
		}

		printer.PrintT(tpl, user)
	}

	return nil
}

func listUsersCmdF(c client.Client, command *cobra.Command, args []string) error {
	page, _ := command.Flags().GetInt("page")
	perPage, _ := command.Flags().GetInt("per-page")
	showAll, _ := command.Flags().GetBool("all")

	tpl := `{{.Id}}: {{.Username}} ({{.Email}})`
	for users, _ := c.GetUsers(page, perPage, ""); users != nil && len(users) > 0; users, _ = c.GetUsers(page, perPage, "") {
		for _, user := range users {
			printer.PrintT(tpl, user)
		}

		if !showAll {
			break
		}
		page++
	}

	return nil
}

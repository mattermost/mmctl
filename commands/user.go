// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Management of users",
}

var UserActivateCmd = &cobra.Command{
	Use:   "activate [emails, usernames, userIds]",
	Short: "Activate users",
	Long:  "Activate users that have been deactivated.",
	Example: `  user activate user@example.com
  user activate username`,
	RunE: withClient(userActivateCmdF),
	Args: cobra.MinimumNArgs(1),
}

var UserDeactivateCmd = &cobra.Command{
	Use:   "deactivate [emails, usernames, userIds]",
	Short: "Deactivate users",
	Long:  "Deactivate users. Deactivated users are immediately logged out of all sessions and are unable to log back in.",
	Example: `  user deactivate user@example.com
  user deactivate username`,
	RunE: withClient(userDeactivateCmdF),
	Args: cobra.MinimumNArgs(1),
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
	Use:     "email [user] [new email]",
	Short:   "Change email of the user",
	Long:    "Change email of the user.",
	Example: "  user email testuser user@example.com",
	RunE:    withClient(updateUserEmailCmdF),
}

var ResetUserMfaCmd = &cobra.Command{
	Use:   "resetmfa [users]",
	Short: "Turn off MFA",
	Long: `Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they login.`,
	Example: "  user resetmfa user@example.com",
	RunE:    withClient(resetUserMfaCmdF),
}

var DeleteAllUsersCmd = &cobra.Command{
	Use:     "deleteall",
	Short:   "Delete all users and all posts. Local command only.",
	Long:    "Permanently delete all users and all related information including posts. This command can only be run in local mode.",
	Example: "  user deleteall",
	Args:    cobra.NoArgs,
	PreRun:  localOnlyPrecheck,
	RunE:    withClient(deleteAllUsersCmdF),
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
	Long:    "List all users",
	Example: "  user list",
	RunE:    withClient(listUsersCmdF),
	Args:    cobra.NoArgs,
}

func init() {
	UserCreateCmd.Flags().String("username", "", "Required. Username for the new user account.")
	_ = UserCreateCmd.MarkFlagRequired("username")
	UserCreateCmd.Flags().String("email", "", "Required. The email address for the new user account.")
	_ = UserCreateCmd.MarkFlagRequired("email")
	UserCreateCmd.Flags().String("password", "", "Required. The password for the new user account.")
	_ = UserCreateCmd.MarkFlagRequired("password")
	UserCreateCmd.Flags().String("nickname", "", "Optional. The nickname for the new user account.")
	UserCreateCmd.Flags().String("firstname", "", "Optional. The first name for the new user account.")
	UserCreateCmd.Flags().String("lastname", "", "Optional. The last name for the new user account.")
	UserCreateCmd.Flags().String("locale", "", "Optional. The locale (ex: en, fr) for the new user account.")
	UserCreateCmd.Flags().Bool("system_admin", false, "Optional. If supplied, the new user will be a system administrator. Defaults to false.")

	DeleteAllUsersCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the user and a DB backup has been performed.")

	ListUsersCmd.Flags().Int("page", 0, "Page number to fetch for the list of users")
	ListUsersCmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	ListUsersCmd.Flags().Bool("all", false, "Fetch all users. --page flag will be ignore if provided")

	UserCmd.AddCommand(
		UserActivateCmd,
		UserDeactivateCmd,
		UserCreateCmd,
		UserInviteCmd,
		SendPasswordResetEmailCmd,
		updateUserEmailCmd,
		ResetUserMfaCmd,
		DeleteAllUsersCmd,
		SearchUserCmd,
		ListUsersCmd,
	)

	RootCmd.AddCommand(UserCmd)
}

func userActivateCmdF(c client.Client, command *cobra.Command, args []string) error {
	changeUsersActiveStatus(c, args, true)

	return nil
}

func changeUsersActiveStatus(c client.Client, userArgs []string, active bool) {
	users := getUsersFromUserArgs(c, userArgs)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("can't find user '%v'", userArgs[i]))
			continue
		}

		err := changeUserActiveStatus(c, user, userArgs[i], active)

		if err != nil {
			printer.PrintError(err.Error())
		}
	}
}

func changeUserActiveStatus(c client.Client, user *model.User, userArg string, activate bool) error {
	if !activate && user.IsSSOUser() {
		printer.Print("You must also deactivate user " + userArg + " in the SSO provider or they will be reactivated on next login or sync.")
	}
	if _, response := c.UpdateUserActive(user.Id, activate); response.Error != nil {
		return fmt.Errorf("unable to change activation status of user: %v", userArg)
	}

	return nil
}

func userDeactivateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	changeUsersActiveStatus(c, args, false)

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
		return errors.New("expected at least two arguments. See help text for details")
	}

	email := args[0]
	if !model.IsValidEmail(email) {
		return errors.New("invalid email")
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
		return fmt.Errorf("can't find team '%v'", teamArg)
	}

	if _, response := c.InviteUsersToTeam(team.Id, invites); response.Error != nil {
		return errors.New("Unable to invite user with email " + email + " to team " + team.Name + ". Error: " + response.Error.Error())
	}

	printer.Print("Invites may or may not have been sent.")

	return nil
}

func sendPasswordResetEmailCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("expected at least one argument. See help text for details")
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
		return errors.New("expected two arguments. See help text for details")
	}

	newEmail := args[1]

	if !model.IsValidEmail(newEmail) {
		return errors.New("invalid email: '" + newEmail + "'")
	}

	if len(args) != 2 {
		return errors.New("expected two arguments. See help text for details")
	}

	user := getUserFromUserArg(c, args[0])
	if user == nil {
		return errors.New("unable to find user '" + args[0] + "'")
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
		return errors.New("expected at least one argument. See help text for details")
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

func deleteAllUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Println("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("aborted: You did not answer YES exactly, in all capitals.")
		}
		fmt.Println("Are you sure you want to permanently delete all user accounts? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("aborted: You did not answer YES exactly, in all capitals.")
		}
	}

	if _, response := c.PermanentDeleteAllUsers(); response.Error != nil {
		return response.Error
	}

	printer.Print("All users successfully deleted")

	return nil
}

func searchUserCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	if len(args) < 1 {
		return errors.New("expected at least one argument. See help text for details")
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
	page, err := command.Flags().GetInt("page")
	if err != nil {
		return err
	}
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return err
	}
	showAll, err := command.Flags().GetBool("all")
	if err != nil {
		return err
	}

	if showAll {
		page = 0
	}

	tpl := `{{.Id}}: {{.Username}} ({{.Email}})`

	for {
		users, res := c.GetUsers(page, perPage, "")
		if res.Error != nil {
			return errors.Wrap(res.Error, "Failed to fetch users")
		}
		if len(users) == 0 {
			break
		}

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

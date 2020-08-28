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

var UpdateUserEmailCmd = &cobra.Command{
	Use:     "email [user] [new email]",
	Short:   "Change email of the user",
	Long:    "Change email of the user.",
	Example: "  user email testuser user@example.com",
	RunE:    withClient(updateUserEmailCmdF),
}

var ChangePasswordUserCmd = &cobra.Command{
	Use:   "change-password <user>",
	Short: "Changes a user's password",
	Long:  "Changes the password of a user by a new one provided. If the user is changing their own password, the flag --current must indicate the current password. The flag --hashed can be used to indicate that the new password has been introduced already hashed",
	Example: `  # if you have system permissions, you can change other user's passwords
  $ mmctl user change-password john_doe --password new-password

  # if you are changing your own password, you need to provide the current one
  $ mmctl user change-password my-username --current current-password --password new-password

  # you can ommit these flags to introduce them interactively
  $ mmctl user change-password my-username
  Are you changing your own password? (YES/NO): YES
  Current password:
  New password:

  # if you have system permissions, you can update the password with the already hashed new password
  $ mmctl user change-password john_doe --password HASHED_PASSWORD --hashed`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(changePasswordUserCmdF),
}

var ResetUserMfaCmd = &cobra.Command{
	Use:   "resetmfa [users]",
	Short: "Turn off MFA",
	Long: `Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they login.`,
	Example: "  user resetmfa user@example.com",
	RunE:    withClient(resetUserMfaCmdF),
}

var DeleteUsersCmd = &cobra.Command{
	Use:   "delete [users]",
	Short: "Delete users",
	Long: `Permanently delete some users.
Permanently deletes one or multiple users along with all related information including posts from the database.`,
	Example: "  user delete user@example.com",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(deleteUsersCmdF),
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

var VerifyUserEmailWithoutTokenCmd = &cobra.Command{
	Use:     "verify [users]",
	Short:   "Verify email of users",
	Long:    "Verify the emails of some users.",
	Example: "  user verify user1",
	RunE:    withClient(verifyUserEmailWithoutTokenCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var UserConvertCmd = &cobra.Command{
	Use:   "convert (--bot [emails] [usernames] [userIds] | --user <username> --password PASSWORD [--email EMAIL])",
	Short: "Convert users to bots, or a bot to a user",
	Long:  "Convert users to bots, or a bot to a user",
	Example: `  # you can convert a user to a bot providing its email, id or username
  $ mmctl user convert user@example.com --bot

  # or multiple users in one go
  $ mmctl user convert user@example.com anotherUser --bot

  # you can convert a bot to a user specifying the email and password that the user will have after conversion
  $ mmctl user convert botusername --email new.email@email.com --password password --user`,
	RunE: withClient(userConvertCmdF),
	Args: cobra.MinimumNArgs(1),
}

func init() {
	UserCreateCmd.Flags().String("username", "", "Required. Username for the new user account")
	_ = UserCreateCmd.MarkFlagRequired("username")
	UserCreateCmd.Flags().String("email", "", "Required. The email address for the new user account")
	_ = UserCreateCmd.MarkFlagRequired("email")
	UserCreateCmd.Flags().String("password", "", "Required. The password for the new user account")
	_ = UserCreateCmd.MarkFlagRequired("password")
	UserCreateCmd.Flags().String("nickname", "", "Optional. The nickname for the new user account")
	UserCreateCmd.Flags().String("firstname", "", "Optional. The first name for the new user account")
	UserCreateCmd.Flags().String("lastname", "", "Optional. The last name for the new user account")
	UserCreateCmd.Flags().String("locale", "", "Optional. The locale (ex: en, fr) for the new user account")
	UserCreateCmd.Flags().Bool("system_admin", false, "Optional. If supplied, the new user will be a system administrator. Defaults to false")

	DeleteUsersCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the user and a DB backup has been performed")
	DeleteAllUsersCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the user and a DB backup has been performed")

	ListUsersCmd.Flags().Int("page", 0, "Page number to fetch for the list of users")
	ListUsersCmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	ListUsersCmd.Flags().Bool("all", false, "Fetch all users. --page flag will be ignore if provided")

	UserConvertCmd.Flags().Bool("bot", false, "If supplied, convert users to bots")
	UserConvertCmd.Flags().Bool("user", false, "If supplied, convert a bot to a user")
	UserConvertCmd.Flags().String("password", "", "The password for converted new user account. Required when \"user\" flag is set")
	UserConvertCmd.Flags().String("username", "", "Username for the converted user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().String("email", "", "The email address for the converted user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().String("nickname", "", "The nickname for the converted user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().String("firstname", "", "The first name for the converted user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().String("lastname", "", "The last name for the converted user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().String("locale", "", "The locale (ex: en, fr) for converted new user account. Required when the \"bot\" flag is set")
	UserConvertCmd.Flags().Bool("system_admin", false, "If supplied, the converted user will be a system administrator. Defaults to false. Required when the \"bot\" flag is set")

	ChangePasswordUserCmd.Flags().StringP("current", "c", "", "The current password of the user. Use only if changing your own password")
	ChangePasswordUserCmd.Flags().StringP("password", "p", "", "The new password for the user")
	ChangePasswordUserCmd.Flags().Bool("hashed", false, "The supplied password is already hashed")

	UserCmd.AddCommand(
		UserActivateCmd,
		UserDeactivateCmd,
		UserCreateCmd,
		UserInviteCmd,
		SendPasswordResetEmailCmd,
		UpdateUserEmailCmd,
		ChangePasswordUserCmd,
		ResetUserMfaCmd,
		DeleteUsersCmd,
		DeleteAllUsersCmd,
		SearchUserCmd,
		ListUsersCmd,
		VerifyUserEmailWithoutTokenCmd,
		UserConvertCmd,
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

func changePasswordUserCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	password, _ := cmd.Flags().GetString("password")
	current, _ := cmd.Flags().GetString("current")
	hashed, _ := cmd.Flags().GetBool("hashed")

	if password == "" {
		var confirm string
		fmt.Printf("Are you changing your own password? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm == "YES" {
			fmt.Printf("Current password: ")
			var err error
			current, err = getPasswordFromStdin()
			if err != nil {
				return errors.New("couldn't read password: " + err.Error())
			}
		}

		fmt.Printf("New password: ")
		var err error
		password, err = getPasswordFromStdin()
		if err != nil {
			return errors.New("couldn't read password: " + err.Error())
		}
	}

	user := getUserFromUserArg(c, args[0])
	if user == nil {
		return errors.New("couldn't find user '" + args[0] + "'")
	}

	if hashed {
		if _, resp := c.UpdateUserHashedPassword(user.Id, password); resp.Error != nil {
			return errors.New("changing user password failed: " + resp.Error.Error())
		}
	} else {
		if _, resp := c.UpdateUserPassword(user.Id, current, password); resp.Error != nil {
			return errors.New("changing user password failed: " + resp.Error.Error())
		}
	}

	printer.PrintT("Password for user {{.Username}} successfully changed", user)
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

func deleteUser(c client.Client, user *model.User) (bool, *model.Response) {
	return c.PermanentDeleteUser(user.Id)
}

func getUserDeleteConfirmation() error {
	var confirm string
	fmt.Println("Have you performed a database backup? (YES/NO): ")
	fmt.Scanln(&confirm)

	if confirm != "YES" {
		return errors.New("aborted: You did not answer YES exactly, in all capitals")
	}
	fmt.Println("Are you sure you want to delete the users specified? All data will be permanently deleted? (YES/NO): ")
	fmt.Scanln(&confirm)
	if confirm != "YES" {
		return errors.New("aborted: You did not answer YES exactly, in all capitals")
	}
	return nil
}

func deleteUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		if err := getUserDeleteConfirmation(); err != nil {
			return err
		}
	}

	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError("Unable to find user '" + args[i] + "'")
			continue
		}
		if _, response := deleteUser(c, user); response.Error != nil {
			printer.PrintError("Unable to delete user '" + user.Username + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Deleted user '{{.Username}}'", user)
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
			return errors.New("aborted: You did not answer YES exactly, in all capitals")
		}
		fmt.Println("Are you sure you want to permanently delete all user accounts? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("aborted: You did not answer YES exactly, in all capitals")
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

func verifyUserEmailWithoutTokenCmdF(c client.Client, cmd *cobra.Command, userArgs []string) error {
	users := getUsersFromUserArgs(c, userArgs)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("can't find user '%v'", userArgs[i]))
			continue
		}

		if newUser, resp := c.VerifyUserEmailWithoutToken(user.Id); resp.Error != nil {
			printer.PrintError(fmt.Sprintf("unable to verify user %s email: %s", user.Id, resp.Error))
		} else {
			printer.PrintT("User {{.Username}} verified", newUser)
		}
	}
	return nil
}

func userConvertCmdF(c client.Client, cmd *cobra.Command, userArgs []string) error {
	toBot, _ := cmd.Flags().GetBool("bot")
	toUser, _ := cmd.Flags().GetBool("user")

	if !(toUser || toBot) {
		return fmt.Errorf("either %q flag or %q flag should be provided", "user", "bot")
	}

	if toBot {
		return convertUserToBot(c, cmd, userArgs)
	}

	return convertBotToUser(c, cmd, userArgs)
}

func convertUserToBot(c client.Client, _ *cobra.Command, userArgs []string) error {
	users := getUsersFromUserArgs(c, userArgs)
	for _, user := range users {
		if user == nil {
			continue
		}
		bot, resp := c.ConvertUserToBot(user.Id)
		if resp.Error != nil {
			printer.PrintError(resp.Error.Error())
			continue
		}

		printer.PrintT("{{.Username}} converted to bot.", bot)
	}
	return nil
}

func convertBotToUser(c client.Client, cmd *cobra.Command, userArgs []string) error {
	user := getUserFromUserArg(c, userArgs[0])
	if user == nil {
		return fmt.Errorf("could not find user by %q", userArgs[0])
	}

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		return errors.New("password is required")
	}

	up := &model.UserPatch{Password: &password}

	username, _ := cmd.Flags().GetString("username")
	if username == "" {
		if user.Username == "" {
			return errors.New("username is empty")
		}
	} else {
		up.Username = model.NewString(username)
	}

	email, _ := cmd.Flags().GetString("email")
	if email == "" {
		if user.Email == "" {
			return errors.New("email is empty")
		}
	} else {
		up.Email = model.NewString(email)
	}

	nickname, _ := cmd.Flags().GetString("nickname")
	if nickname != "" {
		up.Nickname = model.NewString(nickname)
	}

	firstname, _ := cmd.Flags().GetString("firstname")
	if firstname != "" {
		up.FirstName = model.NewString(firstname)
	}

	lastname, _ := cmd.Flags().GetString("lastname")
	if lastname != "" {
		up.LastName = model.NewString(lastname)
	}

	locale, _ := cmd.Flags().GetString("locale")
	if locale != "" {
		up.Locale = model.NewString(locale)
	}

	systemAdmin, _ := cmd.Flags().GetBool("system_admin")

	user, resp := c.ConvertBotToUser(user.Id, up, systemAdmin)
	if resp.Error != nil {
		return resp.Error
	}

	printer.PrintT("{{.Username}} converted to user.", user)

	return nil
}

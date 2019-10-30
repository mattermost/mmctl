package commands

import (
	"errors"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

var RolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Management of user roles",
}

var MakeSystemAdminCmd = &cobra.Command{
	Use:     "system_admin [users]",
	Short:   "Set a user as system admin",
	Long:    "Make some users system admins",
	Example: "  roles system_admin user1",
	RunE:    withClient(makeSystemAdminCmdF),
}

var MakeMemberCmd = &cobra.Command{
	Use:     "member [users]",
	Short:   "Remove system admin privileges",
	Long:    "Remove system admin privileges from some users.",
	Example: "  roles member user1",
	RunE:    withClient(makeMemberCmdF),
}

func init() {
	RolesCmd.AddCommand(
		MakeSystemAdminCmd,
		MakeMemberCmd,
	)
	RootCmd.AddCommand(RolesCmd)
}

func makeSystemAdminCmdF(c client.Client, command *cobra.Command, args []string) error {
	printer.SetSingle(true)

	if len(args) < 1 {
		return errors.New("enter at least one user")
	}

	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			return errors.New("Unable to find user '" + args[i] + "'")
		}

		systemAdmin := false
		var newRoles []string

		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			switch role {
			case model.SYSTEM_ADMIN_ROLE_ID:
				systemAdmin = true
			default:
				newRoles = append(newRoles, role)
			}
		}

		if !systemAdmin {
			newRoles = append(newRoles, model.SYSTEM_ADMIN_ROLE_ID)
			_, response := c.UpdateUserRoles(user.Id, strings.Join(newRoles, " "))
			if response.Error != nil {
				return errors.New("Unable to update user roles. Error: " + response.Error.Error())
			}
			printer.PrintT("System admin role assigned to user {{.Username}}", user)
		}

	}

	return nil
}

func makeMemberCmdF(c client.Client, command *cobra.Command, args []string) error {
	printer.SetSingle(true)

	if len(args) < 1 {
		return errors.New("enter at least one user")
	}

	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			return errors.New("Unable to find user '" + args[i] + "'")
		}

		systemAdmin := true
		var newRoles []string

		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			switch role {
			case model.SYSTEM_ADMIN_ROLE_ID:
				systemAdmin = true
			default:
				newRoles = append(newRoles, role)
			}
		}

		if systemAdmin {
			newRoles = append(newRoles, model.SYSTEM_USER_ROLE_ID)
			_, response := c.UpdateUserRoles(user.Id, strings.Join(newRoles, " "))
			if response.Error != nil {
				return errors.New("Unable to update user roles. Error: " + response.Error.Error())
			}
			printer.PrintT("System admin role revoked for user {{.Username}}", user)
		}

	}

	return nil
}

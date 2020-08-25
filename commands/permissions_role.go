// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var RoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Management of roles",
}

var ShowCmd = &cobra.Command{
	Use:     "show [role_name]",
	Short:   "Show the role information",
	Long:    "Show all the information about a role.",
	Example: `  permissions show system_user`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(showRoleCmdF),
}

var AssignCmd = &cobra.Command{
	Use:   "assign [role_name] [username...]",
	Short: "Assign users to role (EE Only)",
	Long:  "Assign users to a role by username (Only works in Enterprise Edition).",
	Example: `  # Assign users with usernames 'john.doe' and 'jane.doe' to the role named 'system_admin'.
  permissions assign system_admin john.doe jane.doe
  
  # Examples using other system roles
  permissions assign system_manager john.doe jane.doe
  permissions assign system_user_manager john.doe jane.doe
  permissions assign system_read_only_admin john.doe jane.doe`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(assignUsersCmdF),
}

var UnassignCmd = &cobra.Command{
	Use:   "unassign [role_name] [username...]",
	Short: "Unassign users from role (EE Only)",
	Long:  "Unassign users from a role by username (Only works in Enterprise Edition).",
	Example: `  # Unassign users with usernames 'john.doe' and 'jane.doe' from the role named 'system_admin'.
  permissions unassign system_admin john.doe jane.doe

  # Examples using other system roles
  permissions unassign system_manager john.doe jane.doe
  permissions unassign system_user_manager john.doe jane.doe
  permissions unassign system_read_only_admin john.doe jane.doe`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(unassignUsersCmdF),
}

func init() {
	RoleCmd.AddCommand(
		AssignCmd,
		UnassignCmd,
		ShowCmd,
	)

	PermissionsCmd.AddCommand(
		RoleCmd,
	)
}

func showRoleCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	tpl := `Name: {{.Name}}
Display Name: {{.DisplayName}}
Description: {{.Description}}
Permissions: {{.Permissions}}
{{range .Permissions}}
  - {{.}}
{{end}}
{{if .BuiltIn}}
Built in: yes
{{else}}
Built in: no
{{end}}
{{if .SchemeManaged}}
Scheme Managed: yes
{{else}}
Scheme Managed: no
{{end}}
`

	printer.PrintT(tpl, role)

	return nil
}

func assignUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	users := getUsersFromUserArgs(c, args[1:])

	for i, user := range users {
		if user == nil {
			printer.PrintError("Couldn't find user '" + args[i+1] + "'.")
			continue
		}

		var userHasRequestedRole bool
		startingRoles := strings.Fields(user.Roles)
		for _, roleName := range startingRoles {
			if roleName == role.Name {
				userHasRequestedRole = true
			}
		}

		if userHasRequestedRole {
			continue
		}

		userRoles := append(startingRoles, role.Name)
		_, response = c.UpdateUserRoles(user.Id, strings.Join(userRoles, " "))
		if response.Error != nil {
			return response.Error
		}
	}

	return nil
}

func unassignUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args[1:])

	for i, user := range users {
		if user == nil {
			printer.PrintError("Couldn't find user '" + args[i+1] + "'.")
			continue
		}

		userRoles := strings.Fields(user.Roles)
		originalCount := len(userRoles)

		for i := 0; i < len(userRoles); i++ {
			if userRoles[i] == args[0] {
				userRoles = append(userRoles[:i], userRoles[i+1:]...)
				i--
			}
		}

		if originalCount > len(userRoles) {
			_, response := c.UpdateUserRoles(user.Id, strings.Join(userRoles, " "))
			if response.Error != nil {
				return response.Error
			}
		}
	}

	return nil
}

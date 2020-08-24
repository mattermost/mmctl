// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var PermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Management of permissions",
}

var RoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Management of roles",
}

var AddPermissionsCmd = &cobra.Command{
	Use:     "add [role] [permission...]",
	Short:   "Add permissions to a role (EE Only)",
	Long:    `Add one or more permissions to an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions add system_user list_open_teams`,
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(addPermissionsCmdF),
}

var RemovePermissionsCmd = &cobra.Command{
	Use:     "remove [role] [permission...]",
	Short:   "Remove permissions from a role (EE Only)",
	Long:    `Remove one or more permissions from an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions remove system_user list_open_teams`,
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(removePermissionsCmdF),
}

var ShowRoleCmd = &cobra.Command{
	Use:        "show [role_name]",
	Deprecated: "please use \"role show\" instead",
	Short:      "Show the role information",
	Long:       "Show all the information about a role.",
	Example:    `  permissions show system_user`,
	Args:       cobra.ExactArgs(1),
	RunE:       withClient(showRoleCmdF),
}

var AssignUsersCmd = &cobra.Command{
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

var UnassignUsersCmd = &cobra.Command{
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
		AssignUsersCmd,
		UnassignUsersCmd,
		ShowRoleCmd,
	)

	PermissionsCmd.AddCommand(
		AddPermissionsCmd,
		RemovePermissionsCmd,
		ShowRoleCmd,
		RoleCmd,
	)

	RootCmd.AddCommand(PermissionsCmd)
}

func addPermissionsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}
	newPermissions := append(role.Permissions, args[1:]...)
	patchRole := model.RolePatch{
		Permissions: &newPermissions,
	}

	if _, response = c.PatchRole(role.Id, &patchRole); response.Error != nil {
		return response.Error
	}

	return nil
}

func removePermission(permissions []string, permission string) []string {
	newPermissions := []string{}
	for _, p := range permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}
	return newPermissions
}

func removePermissionsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	newPermissions := role.Permissions
	for _, arg := range args[1:] {
		newPermissions = removePermission(newPermissions, arg)
	}

	patchRole := model.RolePatch{
		Permissions: &newPermissions,
	}

	if _, response = c.PatchRole(role.Id, &patchRole); response.Error != nil {
		return response.Error
	}

	return nil
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

	for _, user := range users {
		startingRoles := strings.Fields(user.Roles)
		for _, roleName := range startingRoles {
			if roleName == role.Name {
				return nil
			}
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

	for _, user := range users {
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

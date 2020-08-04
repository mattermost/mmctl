// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"

	"github.com/spf13/cobra"
)

var PermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Management of permissions and roles",
}

var AddPermissionsCmd = &cobra.Command{
	Use:   "add [role] [permission...]",
	Short: "Add permissions to a role (EE Only)",
	Long:  `Add one or more permissions to an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions add system_user list_open_teams
  permissions add system_manager sysconsole_read_user_management_channels --ancillary`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(addPermissionsCmdF),
}

var RemovePermissionsCmd = &cobra.Command{
	Use:   "remove [role] [permission...]",
	Short: "Remove permissions from a role (EE Only)",
	Long:  `Remove one or more permissions from an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions remove system_user list_open_teams
  permissions remove system_manager sysconsole_read_user_management_channels --ancillary`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(removePermissionsCmdF),
}

var ShowRoleCmd = &cobra.Command{
	Use:     "show [role_name]",
	Short:   "Show the role information",
	Long:    "Show all the information about a role.",
	Example: `  permissions show system_user`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(showRoleCmdF),
}

var AssignUsersCmd = &cobra.Command{
	Use:   "assign [role_name] [username...]",
	Short: "Assign users to role (EE Only)",
	Long:  "Assign users to a role by username (Only works in Enterprise Edition).",
	Example: `  permissions assign system_admin john.doe jane.doe
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
	Example: `  permissions unassign system_admin john.doe jane.doe
  permissions unassign system_manager john.doe jane.doe
  permissions unassign system_user_manager john.doe jane.doe
  permissions unassign system_read_only_admin john.doe jane.doe`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(unassignUsersCmdF),
}

func init() {
	AddPermissionsCmd.Flags().Bool("include-ancillary", false, "Optional. Add ancillary permissions used by each sysconsole_* permission being added.")
	RemovePermissionsCmd.Flags().Bool("prune-ancillary", false, "Optional. Remove ancillary permissions no longer used by each sysconsole_* permission being removed.")

	PermissionsCmd.AddCommand(
		AddPermissionsCmd,
		RemovePermissionsCmd,
		ShowRoleCmd,
		AssignUsersCmd,
		UnassignUsersCmd,
	)

	RootCmd.AddCommand(PermissionsCmd)
}

func addPermissionsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	addAncillary, _ := cmd.Flags().GetBool("include-ancillary")
	newPermissions := role.Permissions

	for _, permissionID := range args[1:] {
		newPermissions = append(newPermissions, permissionID)

		if !addAncillary {
			continue
		}

		if ancillaryPermissions, ok := model.SysconsoleAncillaryPermissions[permissionID]; ok {
			for _, ancillaryPermission := range ancillaryPermissions {
				newPermissions = append(newPermissions, ancillaryPermission.Id)
			}
		}
	}

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

	// remove the permissions explicitly requested
	newPermissionSet := role.Permissions
	for _, permissionID := range args[1:] {
		newPermissionSet = removePermission(newPermissionSet, permissionID)
	}

	includes := func(haystack []*model.Permission, needle *model.Permission) bool {
		for _, item := range haystack {
			if item.Id == needle.Id {
				return true
			}
		}
		return false
	}

	// optionally remove ancillary permissions
	if ok, _ := cmd.Flags().GetBool("prune-ancillary"); ok {
		// get the set of all ancillary permissions used by the role, after the requested removals
		var ancillaryPermissionsNeededNow []*model.Permission
		for _, permissionID := range newPermissionSet {
			if ancillaryPermissions, ok := model.SysconsoleAncillaryPermissions[permissionID]; ok {
				ancillaryPermissionsNeededNow = append(ancillaryPermissionsNeededNow, ancillaryPermissions...)
			}
		}

		for _, permissionID := range args[1:] {
			if ancillaryPermissions, ok := model.SysconsoleAncillaryPermissions[permissionID]; ok {
				for _, permission := range ancillaryPermissions {
					if !includes(ancillaryPermissionsNeededNow, permission) {
						newPermissionSet = removePermission(newPermissionSet, permission.Id)
					}
				}
			}
		}
	}

	patchRole := model.RolePatch{
		Permissions: &newPermissionSet,
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

	sort.Strings(role.Permissions)

	consolePermissionMap := map[string]bool{}
	for _, perm := range role.Permissions {
		if strings.HasPrefix(perm, "sysconsole_") {
			consolePermissionMap[perm] = true
		}
	}

	getUsedBy := func(permissionID string) []string {
		var usedByIDs []string
		if !strings.HasPrefix(permissionID, "sysconsole_") {
			usedBy := map[string]bool{} // map to make a unique set
			for key, vals := range model.SysconsoleAncillaryPermissions {
				for _, val := range vals {
					if val.Id == permissionID {
						if _, ok := consolePermissionMap[key]; ok {
							usedBy[key] = true
						}
					}
				}
			}
			for key := range usedBy {
				usedByIDs = append(usedByIDs, key)
			}
		}
		return usedByIDs
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Only show the 3-column view if the role has sysconsole permissions
	// sysadmin has every permission, so no point in showing the "Used by"
	// column.
	if len(consolePermissionMap) > 0 && role.Name != "system_admin" {
		fmt.Fprintf(w, "Property\tValue\tUsed by\n")
		fmt.Fprintf(w, "--------\t-----\t-------\n")
		fmt.Fprintf(w, "Name\t%s\t\n", role.Name)
		fmt.Fprintf(w, "DisplayName\t%s\t\n", role.DisplayName)
		fmt.Fprintf(w, "BuiltIn\t%v\t\n", role.BuiltIn)
		fmt.Fprintf(w, "SchemeManaged\t%v\t\n", role.SchemeManaged)
		fmt.Fprintf(w, "Permissions\t%s\t%v\n", role.Permissions[0], strings.Join(getUsedBy(role.Permissions[0]), ", "))
		for _, perm := range role.Permissions[1:] {
			fmt.Fprintf(w, "\t%s\t%v\n", perm, strings.Join(getUsedBy(perm), ", "))
		}
	} else {
		fmt.Fprintf(w, "Property\tValue\n")
		fmt.Fprintf(w, "--------\t-----\n")
		fmt.Fprintf(w, "Name\t%s\n", role.Name)
		fmt.Fprintf(w, "DisplayName\t%s\n", role.DisplayName)
		fmt.Fprintf(w, "BuiltIn\t%v\n", role.BuiltIn)
		fmt.Fprintf(w, "SchemeManaged\t%v\n", role.SchemeManaged)
		fmt.Fprintf(w, "Permissions\t%s\n", role.Permissions[0])
		for _, perm := range role.Permissions[1:] {
			fmt.Fprintf(w, "\t%s\n", perm)
		}
	}

	w.Flush()

	return nil
}

func assignUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	for _, username := range args[1:] {
		user, response := c.GetUserByUsername(username, "")
		if response.Error != nil {
			return response.Error
		}

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
	for _, username := range args[1:] {
		user, response := c.GetUserByUsername(username, "")
		if response.Error != nil {
			return response.Error
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
			_, response = c.UpdateUserRoles(user.Id, strings.Join(userRoles, " "))
			if response.Error != nil {
				return response.Error
			}
		}
	}

	return nil
}

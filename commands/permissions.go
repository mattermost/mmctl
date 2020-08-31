// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"

	"github.com/spf13/cobra"
)

const flagAncillaryPermissions = "ancillary"

var PermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Management of permissions",
}

var AddPermissionsCmd = &cobra.Command{
	Use:   "add [role] [permission...]",
	Short: "Add permissions to a role (EE Only)",
	Long:  `Add one or more permissions to an existing role (Only works in Enterprise Edition).`,
	Example: fmt.Sprintf(`  permissions add system_user list_open_teams
  permissions add system_manager sysconsole_read_user_management_channels --%s`, flagAncillaryPermissions),
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(addPermissionsCmdF),
}

var RemovePermissionsCmd = &cobra.Command{
	Use:   "remove [role] [permission...]",
	Short: "Remove permissions from a role (EE Only)",
	Long:  `Remove one or more permissions from an existing role (Only works in Enterprise Edition).`,
	Example: fmt.Sprintf(`  permissions remove system_user list_open_teams
  permissions remove system_manager sysconsole_read_user_management_channels --%s`, flagAncillaryPermissions),
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(removePermissionsCmdF),
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
	AddPermissionsCmd.Flags().Bool(flagAncillaryPermissions, false, "Optional. Add all of the ancillary permissions used by each sysconsole_* permission being added.")
	RemovePermissionsCmd.Flags().Bool(flagAncillaryPermissions, false, "Optional. Remove each of the ancillary permissions that no longer used by each sysconsole_* permission being removed, if no other sysconsole_* is using it.")

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

	addAncillary, _ := cmd.Flags().GetBool(flagAncillaryPermissions)
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

func removePermissionsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	newPermissionSet := role.Permissions
	for _, permissionID := range args[1:] {
		newPermissionSet = removeFromStringSlice(newPermissionSet, permissionID)
	}

	if ok, _ := cmd.Flags().GetBool(flagAncillaryPermissions); ok {
		var ancillaryPermissionsStillUsed []*model.Permission
		for _, permissionID := range newPermissionSet {
			if ancillaryPermissions, ok := model.SysconsoleAncillaryPermissions[permissionID]; ok {
				ancillaryPermissionsStillUsed = append(ancillaryPermissionsStillUsed, ancillaryPermissions...)
			}
		}

		for _, permissionID := range args[1:] {
			if ancillaryPermissions, ok := model.SysconsoleAncillaryPermissions[permissionID]; ok {
				for _, permission := range ancillaryPermissions {
					if !permissionsSliceIncludes(ancillaryPermissionsStillUsed, permission) {
						newPermissionSet = removeFromStringSlice(newPermissionSet, permission.Id)
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

func removeFromStringSlice(items []string, item string) []string {
	newPermissions := []string{}
	for _, x := range items {
		if x != item {
			newPermissions = append(newPermissions, x)
		}
	}
	return newPermissions
}

func permissionsSliceIncludes(haystack []*model.Permission, needle *model.Permission) bool {
	for _, item := range haystack {
		if item.Id == needle.Id {
			return true
		}
	}
	return false
}

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/spf13/cobra"
)

var PermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Management of permissions and roles",
}

var AddPermissionsCmd = &cobra.Command{
	Use:     "add [role] [permission...]",
	Short:   "Add permissions to a role (EE Only)",
	Long:    `Add one or more permissions to an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions add system_user list_open_teams`,
	Args:    cobra.MinimumNArgs(2),
	RunE:    addPermissionsCmdF,
}

var RemovePermissionsCmd = &cobra.Command{
	Use:     "remove [role] [permission...]",
	Short:   "Remove permissions from a role (EE Only)",
	Long:    `Remove one or more permissions from an existing role (Only works in Enterprise Edition).`,
	Example: `  permissions remove system_user list_open_teams`,
	Args:    cobra.MinimumNArgs(2),
	RunE:    removePermissionsCmdF,
}

var ShowRoleCmd = &cobra.Command{
	Use:     "show [role_name]",
	Short:   "Show the role information",
	Long:    "Show all the information about a role.",
	Example: `  permissions show system_user`,
	Args:    cobra.ExactArgs(1),
	RunE:    showRoleCmdF,
}

func init() {
	PermissionsCmd.AddCommand(
		AddPermissionsCmd,
		RemovePermissionsCmd,
		ShowRoleCmd,
	)

	RootCmd.AddCommand(PermissionsCmd)
}

func addPermissionsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

func removePermissionsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

func showRoleCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	role, response := c.GetRoleByName(args[0])
	if response.Error != nil {
		return response.Error
	}

	fmt.Printf("Name: %s\n", role.Name)
	fmt.Printf("Display Name: %s\n", role.DisplayName)
	fmt.Printf("Description: %s\n", role.Description)
	fmt.Printf("Permissions:\n")
	for _, permission := range role.Permissions {
		fmt.Printf("  - %s\n", permission)
	}
	if role.BuiltIn {
		fmt.Printf("Built in: yes\n")
	} else {
		fmt.Printf("Built in: no\n")
	}
	if role.SchemeManaged {
		fmt.Printf("Scheme Managed: yes\n")
	} else {
		fmt.Printf("Scheme Managed: no\n")
	}

	return nil
}

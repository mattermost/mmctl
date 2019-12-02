package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

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
	Use:     "show [role_name]",
	Short:   "Show the role information",
	Long:    "Show all the information about a role.",
	Example: `  permissions show system_user`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(showRoleCmdF),
}

func init() {
	PermissionsCmd.AddCommand(
		AddPermissionsCmd,
		RemovePermissionsCmd,
		ShowRoleCmd,
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
{{if .BuildIn}}
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

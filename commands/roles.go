// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/spf13/cobra"
)

var RolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage user roles",
}

var RolesSystemAdminCmd = &cobra.Command{
	Use:     "system-admin [users]",
	Aliases: []string{"system_admin"},
	Short:   "Set a user as system admin",
	Long:    "Make some users system admins.",
	Example: `  # You can make one user a sysadmin
  $ mmctl roles system-admin john_doe

  # Or promote multiple users at the same time
  $ mmctl roles system-admin john_doe jane_doe`,
	RunE: withClient(rolesSystemAdminCmdF),
	Args: cobra.MinimumNArgs(1),
}

var RolesMemberCmd = &cobra.Command{
	Use:   "member [users]",
	Short: "Remove system admin privileges",
	Long:  "Remove system admin privileges from some users.",
	Example: `  # You can remove admin privileges from one user
  $ mmctl roles member john_doe

  # Or demote multiple users at the same time
  $ mmctl roles member john_doe jane_doe`,
	RunE: withClient(rolesMemberCmdF),
	Args: cobra.MinimumNArgs(1),
}

func init() {
	RolesCmd.AddCommand(
		RolesSystemAdminCmd,
		RolesMemberCmd,
	)

	RootCmd.AddCommand(RolesCmd)
}

func rolesSystemAdminCmdF(c client.Client, _ *cobra.Command, args []string) error {
	var errs *multierror.Error
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			err := fmt.Errorf("unable to find user %q", args[i])
			errs = multierror.Append(errs, err)
			printer.PrintError(err.Error())
			continue
		}

		systemAdmin := false
		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			if role == model.SystemAdminRoleId {
				systemAdmin = true
			}
		}

		if !systemAdmin {
			roles = append(roles, model.SystemAdminRoleId)
			if _, err := c.UpdateUserRoles(user.Id, strings.Join(roles, " ")); err != nil {
				err := fmt.Errorf("can't update roles for user %q: %s", args[i], err)
				errs = multierror.Append(errs, err)
				printer.PrintError(err.Error())
				continue
			}

			printer.Print(fmt.Sprintf("System admin role assigned to user %q. Current roles are: %s", args[i], strings.Join(roles, ", ")))
		}
	}

	return errs.ErrorOrNil()
}

func rolesMemberCmdF(c client.Client, _ *cobra.Command, args []string) error {
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			printer.PrintError(fmt.Sprintf("unable to find user %q", args[i]))
			continue
		}

		shouldRemoveSysadmin := false
		var newRoles []string

		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			switch role {
			case model.SystemAdminRoleId:
				shouldRemoveSysadmin = true
			default:
				newRoles = append(newRoles, role)
			}
		}

		if shouldRemoveSysadmin {
			if _, err := c.UpdateUserRoles(user.Id, strings.Join(newRoles, " ")); err != nil {
				printer.PrintError(fmt.Sprintf("can't update roles for user %q: %s", args[i], err))
				continue
			}

			printer.Print(fmt.Sprintf("System admin role revoked for user %q. Current roles are: %s", args[i], strings.Join(newRoles, ", ")))
		}
	}

	return nil
}

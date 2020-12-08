package commands

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *MmctlE2ETestSuite) TestResetPermissionsCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	s.Run("Shouldn't let a non-system-admin reset a role's permissions", func() {
		printer.Clean()

		// update the role to have some non-default permissions
		role, err := s.th.App.GetRoleByName(model.SYSTEM_USER_MANAGER_ROLE_ID)
		s.Require().Nil(err)

		defaultPermissions := role.Permissions
		expectedPermissions := []string{model.PERMISSION_USE_GROUP_MENTIONS.Id, model.PERMISSION_USE_CHANNEL_MENTIONS.Id}
		role.Permissions = expectedPermissions
		role, err = s.th.App.UpdateRole(role)
		s.Require().Nil(err)

		// reset to defaults when we're done
		defer func() {
			role.Permissions = defaultPermissions
			_, err = s.th.App.UpdateRole(role)
			s.Require().Nil(err)
		}()

		// try to reset the permissions
		err2 := resetPermissionsCmdF(s.th.Client, &cobra.Command{}, []string{model.SYSTEM_USER_MANAGER_ROLE_ID})
		s.Require().Error(err2)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		// ensure reset didn't happen
		roleAfterResetAttempt, err := s.th.App.GetRoleByName(model.SYSTEM_USER_MANAGER_ROLE_ID)
		s.Require().Nil(err)
		s.Require().ElementsMatch(expectedPermissions, roleAfterResetAttempt.Permissions)
	})

	s.RunForSystemAdminAndLocal("Reset a role's permissions", func(c client.Client) {
		printer.Clean()

		// update the role to have some non-default permissions
		role, err := s.th.App.GetRoleByName(model.SYSTEM_USER_MANAGER_ROLE_ID)
		s.Require().Nil(err)

		defaultPermissions := role.Permissions
		expectedPermissions := []string{model.PERMISSION_USE_GROUP_MENTIONS.Id, model.PERMISSION_USE_CHANNEL_MENTIONS.Id}
		role.Permissions = expectedPermissions
		role, err = s.th.App.UpdateRole(role)

		// reset to defaults when we're done
		defer func() {
			role.Permissions = defaultPermissions
			_, err = s.th.App.UpdateRole(role)
			s.Require().Nil(err)
		}()

		// try to reset the permissions
		err2 := resetPermissionsCmdF(c, &cobra.Command{}, []string{model.SYSTEM_USER_MANAGER_ROLE_ID})
		s.Require().Nil(err2)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		// ensure reset was successful
		roleAfterResetAttempt, err := s.th.App.GetRoleByName(model.SYSTEM_USER_MANAGER_ROLE_ID)
		s.Require().Nil(err)
		s.Require().ElementsMatch(defaultPermissions, roleAfterResetAttempt.Permissions)
	})
}

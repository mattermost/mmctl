// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestShowRoleCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	s.RunForAllClients("Should allow all users to see a role", func(c client.Client) {
		printer.Clean()

		err := showRoleCmdF(c, &cobra.Command{}, []string{model.SYSTEM_ADMIN_ROLE_ID})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("Should return error to all users for a none exitent role", func(c client.Client) {
		printer.Clean()

		err := showRoleCmdF(c, &cobra.Command{}, []string{"none_existent_role"})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestAddPermissionsCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	role, appErr := s.th.App.GetRoleByName(model.SYSTEM_USER_ROLE_ID)
	s.Require().Nil(appErr)
	s.Require().NotContains(role.Permissions, model.PERMISSION_CREATE_BOT.Id)

	s.Run("Should not allow normal user to add a permission to a role", func() {
		printer.Clean()

		err := addPermissionsCmdF(s.th.Client, &cobra.Command{}, []string{model.SYSTEM_USER_ROLE_ID, model.PERMISSION_CREATE_BOT.Id})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should be able to add a permission to a role", func(c client.Client) {
		printer.Clean()

		err := addPermissionsCmdF(c, &cobra.Command{}, []string{model.SYSTEM_USER_ROLE_ID, model.PERMISSION_CREATE_BOT.Id})
		s.Require().NoError(err)
		defer func() {
			permissions := role.Permissions
			newRole, appErr := s.th.App.PatchRole(role, &model.RolePatch{Permissions: &permissions})
			s.Require().Nil(appErr)
			s.Require().NotContains(newRole.Permissions, model.PERMISSION_CREATE_BOT.Id)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		updatedRole, appErr := s.th.App.GetRoleByName(model.SYSTEM_USER_ROLE_ID)
		s.Require().Nil(appErr)
		s.Require().Contains(updatedRole.Permissions, model.PERMISSION_CREATE_BOT.Id)
	})
}

func (s *MmctlE2ETestSuite) TestRemovePermissionsCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	role, appErr := s.th.App.GetRoleByName(model.SYSTEM_USER_ROLE_ID)
	s.Require().Nil(appErr)
	s.Require().Contains(role.Permissions, model.PERMISSION_CREATE_DIRECT_CHANNEL.Id)

	s.Run("Should not allow normal user to remove a permission from a role", func() {
		printer.Clean()

		err := removePermissionsCmdF(s.th.Client, &cobra.Command{}, []string{model.SYSTEM_USER_ROLE_ID, model.PERMISSION_CREATE_DIRECT_CHANNEL.Id})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should be able to remove a permission from a role", func(c client.Client) {
		printer.Clean()

		err := removePermissionsCmdF(c, &cobra.Command{}, []string{model.SYSTEM_USER_ROLE_ID, model.PERMISSION_CREATE_DIRECT_CHANNEL.Id})
		s.Require().NoError(err)
		defer func() {
			permissions := []string{model.PERMISSION_CREATE_DIRECT_CHANNEL.Id}
			newRole, appErr := s.th.App.PatchRole(role, &model.RolePatch{Permissions: &permissions})
			s.Require().Nil(appErr)
			s.Require().Contains(newRole.Permissions, model.PERMISSION_CREATE_DIRECT_CHANNEL.Id)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		updatedRole, appErr := s.th.App.GetRoleByName(model.SYSTEM_USER_ROLE_ID)
		s.Require().Nil(appErr)
		s.Require().NotContains(updatedRole.Permissions, model.PERMISSION_CREATE_DIRECT_CHANNEL.Id)
	})
}

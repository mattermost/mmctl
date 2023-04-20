// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestAssignUsersCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(s.th.Context, &model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.Run("MM-T3721 Should not allow normal user to assign a role", func() {
		printer.Clean()

		err := assignUsersCmdF(s.th.Client, &cobra.Command{}, []string{model.SystemAdminRoleId, user.Email})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3722 Assigning a user to a non-existent role", func(c client.Client) {
		printer.Clean()

		err := assignUsersCmdF(c, &cobra.Command{}, []string{"not_a_role", user.Email})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Assigning non existen user to a role", func(c client.Client) {
		printer.Clean()

		err := assignUsersCmdF(c, &cobra.Command{}, []string{model.SystemManagerRoleId, "non_existent_user"})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3648 Assigning a user to a role", func(c client.Client) {
		printer.Clean()

		err := assignUsersCmdF(c, &cobra.Command{}, []string{model.SystemManagerRoleId, user.Email})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		roles := user.Roles

		u, err2 := s.th.App.GetUser(user.Id)
		s.Require().Nil(err2)
		s.Require().True(u.IsInRole(model.SystemManagerRoleId))

		_, err2 = s.th.App.UpdateUserRoles(s.th.Context, user.Id, roles, false)
		s.Require().Nil(err2)
	})
}

func (s *MmctlE2ETestSuite) TestUnassignUsersCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(s.th.Context, &model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.Run("MM-T3965 Should not allow normal user to unassign a user from a role", func() {
		printer.Clean()

		err := unassignUsersCmdF(s.th.Client, &cobra.Command{}, []string{model.SystemAdminRoleId, s.th.SystemAdminUser.Email})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("MM-T3964 Unassign a user from a role", func(c client.Client) {
		printer.Clean()

		user.Roles = user.Roles + "," + model.SystemManagerRoleId
		_, appErr = s.th.App.UpdateUser(s.th.Context, user, false)
		s.Require().Nil(appErr)
		defer func() {
			user.Roles = model.SystemUserRoleId
			_, appErr := s.th.App.UpdateUser(s.th.Context, user, false)
			s.Require().Nil(appErr)
		}()

		err := unassignUsersCmdF(c, &cobra.Command{}, []string{model.SystemManagerRoleId, user.Email})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		u, err2 := s.th.App.GetUser(user.Id)
		s.Require().Nil(err2)
		s.Require().False(u.IsInRole(model.SystemManagerRoleId))
	})
}

func (s *MmctlE2ETestSuite) TestListRolesCmdF() {
	s.SetupTestHelper().InitBasic()
	mockRoleName := "mockrole" + model.NewId()
	_, appErr := s.th.App.CreateRole(&model.Role{Name: mockRoleName, DisplayName: mockRoleName})
	s.Require().Nil(appErr)
	mockRoleNameWithPermissions := "mockrole" + model.NewId()
	mockRolePermissions := []string{"sysconsole_write_site", "edit_brand"}
	_, appErr = s.th.App.CreateRole(&model.Role{Name: mockRoleNameWithPermissions, DisplayName: mockRoleNameWithPermissions, Permissions: mockRolePermissions})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Should list all roles for syasdmin and local clients", func(c client.Client) {
		printer.Clean()

		err := listRoleCmdF(c, &cobra.Command{}, []string{})
		s.Require().NoError(err)

		size := 0
		for _, line := range printer.GetLines() {
			size += len(line.(string))
		}

		filled, buf := 0, make([]byte, size)
		for _, line := range printer.GetLines() {
			l, _ := line.(string)
			copy(buf[filled:filled+len(l)], l)
			filled += len(l)
		}

		data := string(buf)

		s.Contains(data, mockRoleName)
		s.Contains(data, mockRoleNameWithPermissions)
	})

	s.Run("Should not list teams for Client", func() {
		printer.Clean()

		err := listRoleCmdF(s.th.Client, &cobra.Command{}, []string{})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
	})
}

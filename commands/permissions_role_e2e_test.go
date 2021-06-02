// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestAssignUsersCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(s.th.Context, &model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.Run("Should not allow normal user to assign a role", func() {
		printer.Clean()

		err := assignUsersCmdF(s.th.Client, &cobra.Command{}, []string{model.SYSTEM_ADMIN_ROLE_ID, user.Email})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Assigning a user to a non-existent role", func(c client.Client) {
		printer.Clean()

		err := assignUsersCmdF(c, &cobra.Command{}, []string{"not_a_role", user.Email})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Assigning a user to a role", func(c client.Client) {
		printer.Clean()

		err := assignUsersCmdF(c, &cobra.Command{}, []string{model.SYSTEM_MANAGER_ROLE_ID, user.Email})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		roles := user.Roles

		u, err2 := s.th.App.GetUser(user.Id)
		s.Require().Nil(err2)
		s.Require().True(u.IsInRole(model.SYSTEM_MANAGER_ROLE_ID))

		_, err2 = s.th.App.UpdateUserRoles(user.Id, roles, false)
		s.Require().Nil(err2)
	})
}

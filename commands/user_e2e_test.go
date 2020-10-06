// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestUserActivateCmd() {
	s.SetupTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Activate user", func(c client.Client) {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, false)
		s.Require().Nil(appErr)

		err := userActivateCmdF(c, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().Zero(ruser.DeleteAt)
	})

	s.Run("Activate user without permissions", func() {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, false)
		s.Require().Nil(appErr)

		err := userActivateCmdF(s.th.Client, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "unable to change activation status of user: "+user.Email)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().NotZero(ruser.DeleteAt)
	})

	s.RunForAllClients("Activate nonexistent user", func(c client.Client) {
		printer.Clean()

		err := userActivateCmdF(c, &cobra.Command{}, []string{"nonexistent@email"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "can't find user 'nonexistent@email'")
	})
}

func (s *MmctlE2ETestSuite) TestUserDeactivateCmd() {
	s.SetupTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Deactivate user", func(c client.Client) {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, true)
		s.Require().Nil(appErr)

		err := userDeactivateCmdF(c, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().NotZero(ruser.DeleteAt)
	})

	s.Run("Deactivate user without permissions", func() {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, true)
		s.Require().Nil(appErr)

		err := userDeactivateCmdF(s.th.Client, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "unable to change activation status of user: "+user.Email)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().Zero(ruser.DeleteAt)
	})

	s.RunForAllClients("Deactivate nonexistent user", func(c client.Client) {
		printer.Clean()

		err := userDeactivateCmdF(c, &cobra.Command{}, []string{"nonexistent@email"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "can't find user 'nonexistent@email'")
	})
}

func (s *MmctlE2ETestSuite) TestSearchUserCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Search for an existing user", func(c client.Client) {
		printer.Clean()

		err := searchUserCmdF(c, &cobra.Command{}, []string{s.th.BasicUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		user := printer.GetLines()[0].(*model.User)
		s.Equal(s.th.BasicUser.Username, user.Username)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("Search for a nonexistent user", func(c client.Client) {
		printer.Clean()
		emailArg := "nonexistentUser@example.com"

		err := searchUserCmdF(c, &cobra.Command{}, []string{emailArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find user '"+emailArg+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlE2ETestSuite) TestUserInviteCmdf() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Invite user", func(c client.Client) {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableEmailInvitations
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = true })
		defer s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = *previousVal })

		err := userInviteCmdF(c, &cobra.Command{}, []string{s.th.BasicUser.Email, s.th.BasicTeam.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Invites may or may not have been sent.")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("Inviting when email invitation disabled", func(c client.Client) {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableEmailInvitations
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = false })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = *previousVal })
		}()

		err := userInviteCmdF(c, &cobra.Command{}, []string{s.th.BasicUser.Email, s.th.BasicTeam.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(
			printer.GetErrorLines()[0],
			fmt.Sprintf("Unable to invite user with email %s to team %s. Error: : Email invitations are disabled., ",
				s.th.BasicUser.Email,
				s.th.BasicTeam.Name,
			),
		)
	})

	s.RunForAllClients("Invite user outside of accepted domain", func(c client.Client) {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableEmailInvitations
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = true })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = *previousVal })
		}()

		team := s.th.CreateTeam()
		team.AllowedDomains = "@example.com"
		team, appErr := s.th.App.UpdateTeam(team)
		s.Require().Nil(appErr)

		user := s.th.CreateUser()
		err := userInviteCmdF(c, &cobra.Command{}, []string{user.Email, team.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0],
			fmt.Sprintf(`Unable to invite user with email %s to team %s. Error: : The following email addresses do not belong to an accepted domain: %s. Please contact your System Administrator for details., `,
				user.Email,
				team.Name,
				user.Email,
			))
	})
}

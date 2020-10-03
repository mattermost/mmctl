// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
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

func (s *MmctlE2ETestSuite) TestUserConvertCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Error when no flag provided", func(c client.Client) {
		printer.Clean()

		emailArg := "example@example.com"
		cmd := &cobra.Command{}

		err := userConvertCmdF(c, cmd, []string{emailArg})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Equal("either \"user\" flag or \"bot\" flag should be provided", err.Error())
	})

	s.RunForAllClients("Error for invalid user", func(c client.Client) {
		printer.Clean()

		emailArg := "something@something.com"
		cmd := &cobra.Command{}
		cmd.Flags().Bool("bot", true, "")

		_ = userConvertCmdF(c, cmd, []string{emailArg})
		s.Require().Len(printer.GetLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Valid user to bot convert", func(c client.Client) {
		printer.Clean()

		user, _ := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})

		email := user.Email
		cmd := &cobra.Command{}
		cmd.Flags().Bool("bot", true, "")

		err := userConvertCmdF(c, cmd, []string{email})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		bot := printer.GetLines()[0].(*model.Bot)
		s.Equal(user.Username, bot.Username)
		s.Equal(user.Id, bot.UserId)
		s.Equal(user.Id, bot.OwnerId)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Permission error for valid user to bot convert", func() {
		printer.Clean()

		email := s.th.BasicUser2.Email
		cmd := &cobra.Command{}
		cmd.Flags().Bool("bot", true, "")

		_ = userConvertCmdF(s.th.Client, cmd, []string{email})
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Equal(": You do not have the appropriate permissions., ", printer.GetErrorLines()[0])
	})

	s.RunForSystemAdminAndLocal("Valid bot to user convert", func(c client.Client) {
		printer.Clean()

		username := "fakeuser" + model.NewRandomString(10)
		bot, _ := s.th.App.CreateBot(&model.Bot{Username: username, DisplayName: username, OwnerId: username})

		cmd := &cobra.Command{}
		cmd.Flags().Bool("user", true, "")
		cmd.Flags().String("password", "password", "")

		err := userConvertCmdF(c, cmd, []string{bot.Username})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		user := printer.GetLines()[0].(*model.User)
		s.Equal(user.Username, bot.Username)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Permission error for valid bot to user convert", func() {
		printer.Clean()

		username := "fakeuser" + model.NewRandomString(10)
		bot, _ := s.th.App.CreateBot(&model.Bot{Username: username, DisplayName: username, OwnerId: username})

		cmd := &cobra.Command{}
		cmd.Flags().Bool("user", true, "")
		cmd.Flags().String("password", "password", "")

		err := userConvertCmdF(s.th.Client, cmd, []string{bot.Username})
		s.Require().Error(err)
		s.Equal(": You do not have the appropriate permissions., ", err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

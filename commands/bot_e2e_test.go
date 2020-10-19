// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestBotEnableCmd() {
	s.SetupTestHelper().InitBasic()
	s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableBotAccountCreation = true })

	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("enable a bot", func(c client.Client) {
		printer.Clean()

		newBot, appErr := s.th.App.CreateBot(&model.Bot{Username: model.NewId(), OwnerId: user.Id})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.UpdateBotActive(newBot.UserId, false)
		s.Require().Nil(appErr)

		err := botEnableCmdF(c, &cobra.Command{}, []string{newBot.UserId})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		printedBot := printer.GetLines()[0].(*model.Bot)
		s.Require().Equal(newBot.UserId, printedBot.UserId)
		s.Require().Equal(newBot.Username, printedBot.Username)
		s.Require().Equal(newBot.OwnerId, printedBot.OwnerId)

		bot, appErr := s.th.App.GetBot(newBot.UserId, false)
		s.Require().Nil(appErr)
		s.Require().Equal(newBot.UserId, bot.UserId)
		s.Require().Equal(newBot.Username, bot.Username)
		s.Require().Equal(newBot.OwnerId, bot.OwnerId)
	})

	s.Run("enable a bot without permissions", func() {
		printer.Clean()

		bot, appErr := s.th.App.CreateBot(&model.Bot{Username: model.NewId(), OwnerId: user.Id})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.UpdateBotActive(bot.UserId, false)
		s.Require().Nil(appErr)

		err := botEnableCmdF(s.th.Client, &cobra.Command{}, []string{bot.UserId})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		s.Require().Contains(printer.GetErrorLines()[0], "could not enable bot")
	})

	s.RunForSystemAdminAndLocal("enable a nonexistent bot", func(c client.Client) {
		printer.Clean()

		err := botEnableCmdF(c, &cobra.Command{}, []string{"nonexistent-bot-userid"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		s.Require().Contains(printer.GetErrorLines()[0], "can't find user 'nonexistent-bot-userid'")
	})

	s.RunForSystemAdminAndLocal("enable an already enabled bot", func(c client.Client) {
		printer.Clean()

		newBot, appErr := s.th.App.CreateBot(&model.Bot{Username: model.NewId(), OwnerId: user.Id})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.UpdateBotActive(newBot.UserId, true)
		s.Require().Nil(appErr)

		err := botEnableCmdF(c, &cobra.Command{}, []string{newBot.UserId})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		printedBot := printer.GetLines()[0].(*model.Bot)
		s.Require().Equal(newBot.UserId, printedBot.UserId)
		s.Require().Equal(newBot.Username, printedBot.Username)
		s.Require().Equal(newBot.OwnerId, printedBot.OwnerId)

		bot, appErr := s.th.App.GetBot(newBot.UserId, false)
		s.Require().Nil(appErr)
		s.Require().Equal(newBot.UserId, bot.UserId)
		s.Require().Equal(newBot.Username, bot.Username)
		s.Require().Equal(newBot.OwnerId, bot.OwnerId)
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestBotCreateCmd() {
	s.Run("Should create a bot", func() {
		botArg := "a-bot"

		cmd := &cobra.Command{}
		cmd.Flags().String("display-name", "some-name", "")
		cmd.Flags().String("description", "some-text", "")
		mockBot := model.Bot{Username: botArg, DisplayName: "some-name", Description: "some-text"}

		s.client.
			EXPECT().
			CreateBot(&mockBot).
			Return(&mockBot, &model.Response{Error: nil}).
			Times(1)

		err := botCreateCmdF(s.client, cmd, []string{botArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], fmt.Sprintf("Created bot %s", mockBot.UserId))
	})

	s.Run("Should error if no arg provided", func() {
		err := botCreateCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "expected at least one argument. See help text for details")
	})

	s.Run("Should error when creating a bot", func() {
		printer.Clean()

		botArg := "a-bot"
		mockBot := model.Bot{Username: botArg, DisplayName: "", Description: ""}

		s.client.
			EXPECT().
			CreateBot(&mockBot).
			Return(nil, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := botCreateCmdF(s.client, &cobra.Command{}, []string{botArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Contains(err.Error(), "could not create bot")
	})
}

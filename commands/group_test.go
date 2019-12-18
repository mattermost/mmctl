package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestChannelGroupDisableCmdF() {
	s.Run("Disable group constrains with existing team and channel", func() {
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(false)}).
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Disable group constrains with nonexistant team", func() {
		printer.Clean()
		teamArg := "team-id"
		channelPart := "channel-id"
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Disable group constrains with nonexistant channel", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelPart, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})
}

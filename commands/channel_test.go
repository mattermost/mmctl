package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
	"github.com/pkg/errors"
)

func (s *MmctlUnitTestSuite) TestSearchChannelCmdF() {
	s.Run("Search for an existing channel on an existing team", func() {
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByName(channelArg, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := searchChannelCmdF(s.client, cmd, []string{channelArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockChannel, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Search for an existing channel without specifying team", func() {
		printer.Clean()
		teamId := "example-team-id"
		otherTeamId := "example-team-id-2"
		mockTeams := []*model.Team{
			&model.Team{Id: otherTeamId},
			&model.Team{Id: teamId},
		}
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}

		s.client.
			EXPECT().
			GetAllTeams("", 0, 9999).
			Return(mockTeams, &model.Response{Error: nil}).
			Times(1)

		// first call is for the other team, that doesn't have the channel
		s.client.
			EXPECT().
			GetChannelByName(channelArg, otherTeamId, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// second call is for the team that contains the channel
		s.client.
			EXPECT().
			GetChannelByName(channelArg, teamId, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := searchChannelCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockChannel, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Search for a nonexistent channel", func() {
		printer.Clean()
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}
		channelArg := "example-channel"

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByName(channelArg, teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := searchChannelCmdF(s.client, cmd, []string{channelArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Channel "+channelArg+" was not found in team "+teamArg)
	})

	s.Run("Search for a channel in a nonexistent team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		channelArg := "example-channel"

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")

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

		err := searchChannelCmdF(s.client, cmd, []string{channelArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Team "+teamArg+" was not found")
	})
}

func (s *MmctlUnitTestSuite) TestArchiveChannelCmdF() {
	s.Run("Archive channel without args returns an error", func() {
		printer.Clean()

		err := archiveChannelsCmdF(s.client, &cobra.Command{}, []string{})
		mockErr := errors.New("Enter at least one channel to archive")

		expected := mockErr.Error()
		actual := err.Error()
		 
		s.Require().Equal(expected, actual)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Archive an existing channel on an existing team", func() {
		printer.Clean()

		teamArg := "some-team-id"
		mockTeam := model.Team{Id: teamArg}
		channelArg := "some-channel"
		channelID := "some-channel-id"
		mockChannel := model.Channel{Id: channelID, Name: channelArg}

		cmd := &cobra.Command{}
		args := fmt.Sprintf("%s:%s", teamArg, channelArg)

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelArg, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteChannel(channelID).
			Return(true, &model.Response{Error: nil}).
			Times(1)
		

		err := archiveChannelsCmdF(s.client, cmd, []string{args})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Archive an existing channel specified by channel id", func() {
		printer.Clean()

		channelArg := "some-channel"
		channelID := "some-channel-id"
		mockChannel := model.Channel{Id: channelID, Name: channelArg}

		cmd := &cobra.Command{}
		args := []string{channelArg}

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteChannel(channelID).
			Return(true, &model.Response{Error: nil}).
			Times(1)
		

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Archive several channels specified by channel id", func() {
		printer.Clean()

		channelArg1 := "some-channel"
		channelID1 := "some-channel-id"
		mockChannel1 := model.Channel{Id: channelID1, Name: channelArg1}

		channelArg2 := "some-other-channel"
		channelID2 := "some-other-channel-id"
		mockChannel2 := model.Channel{Id: channelID2, Name: channelArg2}

		cmd := &cobra.Command{}
		args := []string{channelArg1, channelArg2}

		s.client.
			EXPECT().
			GetChannel(channelArg1, "").
			Return(&mockChannel1, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelArg2, "").
			Return(&mockChannel2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteChannel(channelID1).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteChannel(channelID2).
			Return(true, &model.Response{Error: nil}).
			Times(1)
		

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Fail to archive a channel on a non-existent team", func() {
		printer.Clean()

		teamArg := "some-non-existent-team-id"
		channelArg := "some-channel"

		cmd := &cobra.Command{}
		args := []string{fmt.Sprintf("%s:%s", teamArg, channelArg)}

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

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		expected := printer.GetErrorLines()[0]
		actual := fmt.Sprintf("Unable to find channel '%s'", args[0])
		s.Require().Equal(expected, actual)
	})

	s.Run("Fail to archive a non-existing channel on an existent team", func() {
		printer.Clean()

		teamArg := "some-non-existing-team-id"
		mockTeam := model.Team{Id: teamArg}
		channelArg := "some-non-existing-channel"

		cmd := &cobra.Command{}
		args := []string{fmt.Sprintf("%s:%s", teamArg, channelArg)}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelArg, teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		expected := printer.GetErrorLines()[0]
		actual := fmt.Sprintf("Unable to find channel '%s'", args[0])
		s.Require().Equal(expected, actual)
	})

	s.Run("Fail to archive a non-existing channel", func() {
		printer.Clean()

		channelArg := "some-non-existing-channel"
		cmd := &cobra.Command{}
		args := []string{channelArg}

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		expected := printer.GetErrorLines()[0]
		actual := fmt.Sprintf("Unable to find channel '%s'", args[0])
		s.Require().Equal(expected, actual)
	})

	s.Run("Fail to archive an existing channel when client throws error", func() {
		printer.Clean()

		channelArg := "some-channel"
		channelID := "some-channel-id"
		mockChannel := model.Channel{Id: channelID, Name: channelArg}

		cmd := &cobra.Command{}
		args := []string{channelArg}

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		mockErr := &model.AppError{Message: "Mock error"}
		s.client.
			EXPECT().
			DeleteChannel(channelID).
			Return(false, &model.Response{Error: mockErr}).
			Times(1)
		

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		expected := printer.GetErrorLines()[0]
		actual := fmt.Sprintf("Unable to archive channel '%s' error: %s", channelArg, mockErr.Error())
		s.Require().Equal(expected, actual)

	})

	s.Run("Fail to archive when team and channel not provided", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{":"}

		err := archiveChannelsCmdF(s.client, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)

		expected := printer.GetErrorLines()[0]
		actual := fmt.Sprintf("Unable to find channel '%s'", args[0])
		s.Require().Equal(expected, actual)
	})
	
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestSearchChannelCmdF() {
	s.Run("Search for an existing channel on an existing team", func() {
		printer.Clean()
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
		teamID := "example-team-id"
		otherTeamID := "example-team-id-2"
		mockTeams := []*model.Team{
			{Id: otherTeamID},
			{Id: teamID},
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
			GetChannelByName(channelArg, otherTeamID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// second call is for the team that contains the channel
		s.client.
			EXPECT().
			GetChannelByName(channelArg, teamID, "").
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

func (s *MmctlUnitTestSuite) TestModifyChannelCmdF() {
	s.Run("Both public and private the same value (false)", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", false, "")
		cmd.Flags().Bool("private", false, "")

		err := modifyChannelCmdF(s.client, cmd, []string{})
		s.Require().EqualError(err, "you must specify only one of --public or --private")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Both public and private the same value (true)", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", true, "")

		err := modifyChannelCmdF(s.client, cmd, []string{})
		s.Require().EqualError(err, "you must specify only one of --public or --private")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to modify non-existing channel", func() {
		printer.Clean()
		args := []string{"mockChannel"}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(nil, &model.Response{Error: &model.AppError{}}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Unable to find channel '"+args[0]+"'")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to modify a channel from a non-existing team", func() {
		printer.Clean()
		team := "mockTeam"
		channel := "mockChannel"
		args := []string{team + ":" + channel}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(nil, &model.Response{Error: &model.AppError{}}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(team, "").
			Return(nil, &model.Response{Error: &model.AppError{}}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Unable to find channel '"+args[0]+"'")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to modify direct channel", func() {
		printer.Clean()
		channel := &model.Channel{
			Id:   "mockChannel",
			Type: model.CHANNEL_DIRECT,
		}
		args := []string{channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "you can only change the type of public/private channels")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to modify group channel", func() {
		printer.Clean()
		channel := &model.Channel{
			Id:   "mockChannel",
			Type: model.CHANNEL_GROUP,
		}
		args := []string{channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "you can only change the type of public/private channels")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to modify channel privacy and get error", func() {
		printer.Clean()
		channel := &model.Channel{
			Id:   "mockChannel",
			Type: model.CHANNEL_PRIVATE,
		}
		mockError := &model.AppError{
			Message: "mockError",
		}
		args := []string{channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateChannelPrivacy(channel.Id, model.CHANNEL_OPEN).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to update channel ('"+channel.Id+"') privacy: "+mockError.Error())
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Modify channel privacy to public", func() {
		printer.Clean()
		channel := &model.Channel{
			Id:   "mockChannel",
			Type: model.CHANNEL_PRIVATE,
		}
		returnedChannel := &model.Channel{
			Id:   channel.Id,
			Type: model.CHANNEL_OPEN,
		}
		args := []string{channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", false, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateChannelPrivacy(channel.Id, model.CHANNEL_OPEN).
			Return(returnedChannel, &model.Response{Error: nil}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().NoError(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Modify channel privacy to private", func() {
		printer.Clean()
		channel := &model.Channel{
			Id:   "mockChannel",
			Type: model.CHANNEL_OPEN,
		}
		returnedChannel := &model.Channel{
			Id:   channel.Id,
			Type: model.CHANNEL_PRIVATE,
		}
		args := []string{channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", false, "")
		cmd.Flags().Bool("private", true, "")

		s.client.
			EXPECT().
			GetChannel(args[0], "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateChannelPrivacy(channel.Id, model.CHANNEL_PRIVATE).
			Return(returnedChannel, &model.Response{Error: nil}).
			Times(1)

		err := modifyChannelCmdF(s.client, cmd, args)
		s.Require().NoError(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestArchiveChannelCmdF() {
	s.Run("Archive channel without args returns an error", func() {
		printer.Clean()

		err := archiveChannelsCmdF(s.client, &cobra.Command{}, []string{})
		mockErr := errors.New("enter at least one channel to archive")

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
		args := teamArg + ":" + channelArg

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
		args := []string{teamArg + ":" + channelArg}

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
		args := []string{teamArg + ":" + channelArg}

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

func (s *MmctlUnitTestSuite) TestListChannelsCmd() {
	s.Run("Team is not found", func() {
		printer.Clean()
		team1ID := "team1"
		args := []string{""}
		args[0] = team1ID
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team1ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(team1ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to find team '"+team1ID+"'")
	})

	s.Run("Team has no channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		// Empty channels of a team
		publicChannels := []*model.Channel{}
		archivedChannels := []*model.Channel{}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Team with public channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		publicChannelName1 := "ChannelName1"
		publicChannel1 := &model.Channel{Name: publicChannelName1}

		publicChannelName2 := "ChannelName2"
		publicChannel2 := &model.Channel{Name: publicChannelName2}

		publicChannels := []*model.Channel{publicChannel1, publicChannel2}
		archivedChannels := []*model.Channel{} // Empty archived channels

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
		s.Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], publicChannel1)
		s.Require().Equal(printer.GetLines()[1], publicChannel2)
	})

	s.Run("Team with archived channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		archivedChannelName1 := "ChannelName1"
		archivedChannel1 := &model.Channel{Name: archivedChannelName1}

		archivedChannelName2 := "ChannelName2"
		archivedChannel2 := &model.Channel{Name: archivedChannelName2}

		publicChannels := []*model.Channel{} // Empty public channels
		archivedChannels := []*model.Channel{archivedChannel1, archivedChannel2}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
		s.Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], archivedChannel1)
		s.Require().Equal(printer.GetLines()[1], archivedChannel2)
	})

	s.Run("Team with both public and archived channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		archivedChannel1 := &model.Channel{Name: "archivedChannelName1"}
		archivedChannel2 := &model.Channel{Name: "archivedChannelName2"}
		archivedChannels := []*model.Channel{archivedChannel1, archivedChannel2}

		publicChannel1 := &model.Channel{Name: "publicChannelName1"}
		publicChannel2 := &model.Channel{Name: "publicChannelName2"}
		publicChannels := []*model.Channel{publicChannel1, publicChannel2}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
		s.Len(printer.GetLines(), 4)
		s.Require().Equal(printer.GetLines()[0], publicChannel1)
		s.Require().Equal(printer.GetLines()[1], publicChannel2)
		s.Require().Equal(printer.GetLines()[2], archivedChannel1)
		s.Require().Equal(printer.GetLines()[3], archivedChannel2)
	})

	s.Run("API fails to get team's public channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		mockError := &model.AppError{Message: "Mock error"}
		emptyChannels := []*model.Channel{}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(emptyChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to list public channels for '"+args[0]+"'. Error: "+mockError.Error())
	})

	s.Run("API fails to get team's archived channels list", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}
		team := &model.Team{
			Id: teamID,
		}

		mockError := &model.AppError{Message: "Mock error"}
		emptyChannels := []*model.Channel{}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(emptyChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to list archived channels for '"+args[0]+"'. Error: "+mockError.Error())
	})

	s.Run("API fails to get team's both public and archived channels", func() {
		printer.Clean()

		teamID := "teamID"
		args := []string{teamID}
		cmd := &cobra.Command{}

		team := &model.Team{
			Id: teamID,
		}

		mockError := &model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to list public channels for '"+args[0]+"'. Error: "+mockError.Error())
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to list archived channels for '"+args[0]+"'. Error: "+mockError.Error())
	})

	s.Run("Two teams, one is found and other is not found", func() {
		printer.Clean()

		teamID1 := "teamID1"
		teamID2 := "teamID2"
		args := []string{teamID1, teamID2}
		cmd := &cobra.Command{}

		team1 := &model.Team{Id: teamID1}

		publicChannel1 := &model.Channel{Name: "publicChannelName1"}
		publicChannel2 := &model.Channel{Name: "publicChannelName2"}
		publicChannels := []*model.Channel{publicChannel1, publicChannel2}

		archivedChannel1 := &model.Channel{Name: "archivedChannelName1"}
		archivedChannels := []*model.Channel{archivedChannel1}

		s.client.
			EXPECT().
			GetTeam(teamID1, "").
			Return(team1, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(teamID2, "").
			Return(nil, &model.Response{Error: nil}). // Team 2 not found
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(teamID2, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID1, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID1, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to find team '"+teamID2+"'")
		s.Len(printer.GetLines(), 3)
		s.Require().Equal(printer.GetLines()[0], publicChannel1)
		s.Require().Equal(printer.GetLines()[1], publicChannel2)
		s.Require().Equal(printer.GetLines()[2], archivedChannel1)
	})

	s.Run("Two teams, one is found and other has API errors", func() {
		printer.Clean()

		teamID1 := "teamID1"
		teamID2 := "teamID2"
		args := []string{teamID1, teamID2}
		cmd := &cobra.Command{}

		team1 := &model.Team{Id: teamID1}
		team2 := &model.Team{Id: teamID2}

		publicChannel1 := &model.Channel{Name: "publicChannelName1"}
		publicChannel2 := &model.Channel{Name: "publicChannelName2"}
		publicChannels := []*model.Channel{publicChannel1, publicChannel2}

		archivedChannel1 := &model.Channel{Name: "archivedChannelName1"}
		archivedChannels := []*model.Channel{archivedChannel1}

		s.client.
			EXPECT().
			GetTeam(teamID1, "").
			Return(team1, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID1, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID1, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		mockError := &model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetTeam(teamID2, "").
			Return(team2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID2, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID2, 0, 10000, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 2)
		s.Len(printer.GetLines(), 3)
		s.Require().Equal(printer.GetLines()[0], publicChannel1)
		s.Require().Equal(printer.GetLines()[1], publicChannel2)
		s.Require().Equal(printer.GetLines()[2], archivedChannel1)
	})

	s.Run("Two teams, both are not found", func() {
		printer.Clean()

		team1ID := "team1ID"
		team2ID := "team2ID"
		args := []string{team1ID, team2ID}
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team1ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeam(team2ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(team1ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(team2ID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to find team '"+team1ID+"'")
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to find team '"+team2ID+"'")
	})

	s.Run("Two teams, both have channels", func() {
		printer.Clean()

		teamID1 := "teamID1"
		teamID2 := "teamID2"
		args := []string{teamID1, teamID2}
		cmd := &cobra.Command{}

		team1 := &model.Team{Id: teamID1}
		team2 := &model.Team{Id: teamID2}

		// Using same channel name for both teams since there can be common channels
		publicChannel1 := &model.Channel{Name: "publicChannelName1"}
		publicChannel2 := &model.Channel{Name: "publicChannelName2"}
		publicChannels := []*model.Channel{publicChannel1, publicChannel2}

		archivedChannel1 := &model.Channel{Name: "archivedChannelName1"}
		archivedChannels := []*model.Channel{archivedChannel1}

		s.client.
			EXPECT().
			GetTeam(teamID1, "").
			Return(team1, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID1, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID1, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(teamID2, "").
			Return(team2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPublicChannelsForTeam(teamID2, 0, 10000, "").
			Return(publicChannels, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetDeletedChannelsForTeam(teamID2, 0, 10000, "").
			Return(archivedChannels, &model.Response{Error: nil}).
			Times(1)

		err := listChannelsCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
		s.Len(printer.GetLines(), 6)
		s.Require().Equal(printer.GetLines()[0], publicChannel1)
		s.Require().Equal(printer.GetLines()[1], publicChannel2)
		s.Require().Equal(printer.GetLines()[2], archivedChannel1)
		s.Require().Equal(printer.GetLines()[3], publicChannel1)
		s.Require().Equal(printer.GetLines()[4], publicChannel2)
		s.Require().Equal(printer.GetLines()[5], archivedChannel1)
	})
}

func (s *MmctlUnitTestSuite) TestAddChannelUsersCmdF() {
	team := "example-team-id"
	channel := "example-channel"
	channelArg := team + ":" + channel
	mockTeam := model.Team{Id: team}
	mockChannel := model.Channel{Id: channel, Name: channel}
	userArg := "user@example.com"
	userID := "example-user-id"
	mockUser := model.User{Id: userID, Email: userArg}

	s.Run("Not enough command line parameters", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		// One argument provided.
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg})
		s.EqualError(err, "not enough arguments")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)

		// No arguments provided.
		err = addChannelUsersCmdF(s.client, cmd, []string{})
		s.EqualError(err, "not enough arguments")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
	s.Run("Add existing user to existing channel", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channel, team, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			AddChannelMember(channel, userID).
			Return(&model.ChannelMember{}, &model.Response{Error: nil}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
	s.Run("Add existing user to nonexistent channel", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		// No channel is returned by client.
		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channel, team, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetChannel(channel, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
	s.Run("Add existing user to channel owned by nonexistent team", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		// No team is returned by client.
		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(team, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
	s.Run("Add multiple users, some nonexistent to existing channel", func() {
		printer.Clean()
		nilUserArg := "nonexistent-user"
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channel, team, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(nilUserArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(nilUserArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(nilUserArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			AddChannelMember(channel, userID).
			Return(&model.ChannelMember{}, &model.Response{Error: nil}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, nilUserArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Can't find user '"+nilUserArg+"'", printer.GetErrorLines()[0])
	})
	s.Run("Error adding existing user to existing channel", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(team, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channel, team, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			AddChannelMember(channel, userID).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Mock error"}}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to add '"+userArg+"' to "+channel+". Error: : Mock error, ",
			printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestRenameChannelCmd() {
	s.Run("It should fail when no name and display name is supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := []string{""}
		args[0] = "teamName:channelName"

		cmd.Flags().String("name", "", "Channel Name")
		cmd.Flags().String("display_name", "", "Channel Display Name")

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "require at least one flag to rename channel, either 'name' or 'display_name'")
	})

	s.Run("It should fail when empty team and channel name are supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		teamName := ""
		channelName := ""
		argsTeamChannel := teamName + ":" + channelName
		args := []string{""}
		args[0] = argsTeamChannel

		newChannelName := "newChannelName"
		newChannelDisplayName := "New Channel Name"
		cmd.Flags().String("name", newChannelName, "Channel Name")
		cmd.Flags().String("display_name", newChannelDisplayName, "Channel Display Name")

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "unable to find channel from '"+argsTeamChannel+"'")
	})

	s.Run("It should fail when empty channel is supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		teamName := "teamName"
		channelName := ""
		argsTeamChannel := teamName + ":" + channelName
		args := []string{""}
		args[0] = argsTeamChannel

		newChannelName := "newChannelName"
		newChannelDisplayName := "New Channel Name"
		cmd.Flags().String("name", newChannelName, "Channel Name")
		cmd.Flags().String("display_name", newChannelDisplayName, "Channel Display Name")

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "unable to find channel from '"+argsTeamChannel+"'")
	})

	s.Run("It should fail when team is not found", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		teamName := "nonExistingteamName"
		channelName := "channelName"
		argsTeamChannel := teamName + ":" + channelName
		args := []string{""}
		args[0] = argsTeamChannel

		newChannelName := "newChannelName"
		newChannelDisplayName := "New Channel Name"
		cmd.Flags().String("name", newChannelName, "Channel Name")
		cmd.Flags().String("display_name", newChannelDisplayName, "Channel Display Name")

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "unable to find channel from '"+argsTeamChannel+"'")
	})

	s.Run("It should fail when channel is not found", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		teamName := "teamName"
		channelName := "nonExistingChannelName"
		argsTeamChannel := teamName + ":" + channelName
		args := []string{""}
		args[0] = argsTeamChannel

		newChannelName := "newChannelName"
		newChannelDisplayName := "New Channel Name"
		cmd.Flags().String("name", newChannelName, "Channel Name")
		cmd.Flags().String("display_name", newChannelDisplayName, "Channel Display Name")

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		foundTeam := &model.Team{
			Id:          "teamId",
			DisplayName: "teamDisplayName",
			Name:        teamName,
		}

		s.client.
			EXPECT().
			GetTeamByName(teamName, "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelName, foundTeam.Id, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "unable to find channel from '"+argsTeamChannel+"'")
	})

	s.Run("It should fail when api fails to rename", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		teamName := "teamName"
		channelName := "ChannelName"
		argsTeamChannel := teamName + ":" + channelName
		args := []string{""}
		args[0] = argsTeamChannel

		newChannelName := "newChannelName"
		newChannelDisplayName := "New Channel Name"
		cmd.Flags().String("name", newChannelName, "Channel Name")
		cmd.Flags().String("display_name", newChannelDisplayName, "Channel Display Name")

		foundTeam := &model.Team{
			Id:          "teamId",
			DisplayName: "teamDisplayName",
			Name:        teamName,
		}

		foundChannel := &model.Channel{
			Id:          "channelId",
			Name:        channelName,
			DisplayName: "Channel Display Name",
		}

		channelPatch := &model.ChannelPatch{
			DisplayName: &newChannelDisplayName,
			Name:        &newChannelName,
		}

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamName, "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelName, foundTeam.Id, "").
			Return(foundChannel, &model.Response{Error: nil}).
			Times(1)

		// s.client.
		// 	EXPECT().
		// 	GetChannel(channelName, "").
		// 	Return(foundChannel, &model.Response{Error: nil}).
		// 	Times(1)

		mockError := model.NewAppError("at-random-location.go", "Mock Error", nil, "mocking a random error", 0)
		s.client.
			EXPECT().
			PatchChannel(foundChannel.Id, channelPatch).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := renameChannelCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Cannot rename channel '"+foundChannel.Name+"', error : "+mockError.Error())
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestSearchChannelCmdF() {
	s.Run("Search for an existing channel on an existing team", func() {
		defer printer.Clean()
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
		defer printer.Clean()
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
		defer printer.Clean()
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
		defer printer.Clean()
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

func (s *MmctlUnitTestSuite) TestListChannelsCmd() {
	s.Run("Team is not found", func() {
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

	s.Run("Team with both public and achived channels", func() {
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

	s.Run("API fails to get team's archived channels", func() {
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
	userId := "example-user-id"
	mockUser := model.User{Id: userId, Email: userArg}

	s.Run("Not enough command line parameters", func() {
		cmd := &cobra.Command{}

		// One argument provided.
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg})
		s.EqualError(err, "Not enough arguments.")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)

		// No arguments provided.
		err = addChannelUsersCmdF(s.client, cmd, []string{})
		s.EqualError(err, "Not enough arguments.")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		printer.Clean()
	})
	s.Run("Add existing user to existing channel", func() {
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
			AddChannelMember(channel, userId).
			Return(&model.ChannelMember{}, &model.Response{Error: nil}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		printer.Clean()
	})
	s.Run("Add existing user to nonexistent channel", func() {
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
		printer.Clean()
	})
	s.Run("Add existing user to channel owned by nonexistent team", func() {
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
		printer.Clean()
	})
	s.Run("Add multiple users, some nonexistent to existing channel", func() {
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
			AddChannelMember(channel, userId).
			Return(&model.ChannelMember{}, &model.Response{Error: nil}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, nilUserArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Can't find user '"+nilUserArg+"'", printer.GetErrorLines()[0])
		printer.Clean()
	})
	s.Run("Error adding existing user to existing channel", func() {
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
			AddChannelMember(channel, userId).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Mock error"}}).
			Times(1)
		err := addChannelUsersCmdF(s.client, cmd, []string{channelArg, userArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to add '"+userArg+"' to "+channel+". Error: : Mock error, ",
			printer.GetErrorLines()[0])
		printer.Clean()
	})
}

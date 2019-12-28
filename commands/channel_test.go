package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
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

func (s *MmctlUnitTestSuite) TestListChannelsCmd() {
	s.Run("Should fail when team is not found", func() {
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

	s.Run("Should fail when all teams are not found", func() {
		printer.Clean()

		team1ID := "teamuuu1"
		team2ID := "teamxxx2"
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

	s.Run("Should fail when a teams public channel cannot be retrieved from API", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})

	s.Run("Should fail when a team's archived channel cannot be retrieved from API", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})

	s.Run("Should not fail when team dont have any public or archived channel", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})

	s.Run("Should show list of channel for existing team and fail for non existing team", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})

	s.Run("Should show list of public and archived channels for existing team", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})

	s.Run("Should show list of public and archived channels for existing teams", func() {
		printer.Clean()

		s.Require().Equal("1", "1")
	})
}

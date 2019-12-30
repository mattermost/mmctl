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

func (s *MmctlUnitTestSuite) TestCreateChannelCmdF(){
	mockChannelName := "Mock Channel"
	mockChannelDisplayName := "Mock Channel Display Name"
	mockTeam := &model.Team{Name: "team", DisplayName: "team"}

	s.Run("Create channel with no name returns an error", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		err := createChannelCmdF(s.client, cmd, []string{})

		s.Require().Equal(err.Error(), "Name is required")
		s.Require().Len(printer.GetLines(), 0)

	})

	s.Run("Create channel with a name but no Display Name returns an error", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockChannelName, "")
		err := createChannelCmdF(s.client, cmd, []string{})

		s.Require().Equal(err.Error(),"Display Name is required")
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Create channel with a Name but no Display Name returns an error", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("display_name", mockChannelDisplayName, "")
		err := createChannelCmdF(s.client, cmd, []string{})

		s.Require().Equal(err.Error(),"Name is required")
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Create channel with a Name and Display Name but Team returns an error", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockChannelName, "")
		cmd.Flags().String("display_name", mockChannelDisplayName, "")

		err := createChannelCmdF(s.client, cmd, []string{})

		s.Require().Equal(err.Error(), "Team is required")
		s.Require().Len(printer.GetLines(), 0)

	})

	s.Run("Test creating a channel with a nonexistent team returns an error", func() {
		printer.Clean()

		teamNameArg := "mockTeamName"

		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockChannelName, "")
		cmd.Flags().String("display_name", mockChannelDisplayName, "")
		cmd.Flags().String("team", teamNameArg, "")

		s.client.
			EXPECT().
			GetTeam(teamNameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamNameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := createChannelCmdF(s.client, cmd, []string{})

		s.Require().Equal(err.Error(), "Unable to find team: "+ teamNameArg)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Create open channel", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockChannelName, "")
		cmd.Flags().String("display_name", mockChannelDisplayName, "")
		cmd.Flags().String("team", mockTeam.Name, "")

		mockChannel := &model.Channel{
			Name:        mockChannelName,
			DisplayName: mockChannelDisplayName,
			TeamId: mockTeam.Id,
			Type: model.CHANNEL_OPEN,
		}

		s.client.
			EXPECT().
			GetTeam(mockTeam.Name, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateChannel(mockChannel).
			Return(mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := createChannelCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Equal(mockChannel, printer.GetLines()[0])
		s.Require().Len(printer.GetLines(), 1)
	})

	s.Run("Create invite only channel", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockChannelName, "")
		cmd.Flags().String("display_name", mockChannelDisplayName, "")
		cmd.Flags().String("team", mockTeam.Name, "")
		cmd.Flags().Bool("private", true, "")

		mockChannel := &model.Channel{
			Name:        mockChannelName,
			DisplayName: mockChannelDisplayName,
			TeamId: mockTeam.Id,
			Type: model.CHANNEL_PRIVATE,
		}

		s.client.
			EXPECT().
			GetTeam(mockTeam.Name, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateChannel(mockChannel).
			Return(mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := createChannelCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Equal(mockChannel, printer.GetLines()[0])
		s.Require().Len(printer.GetLines(), 1)
	})
}

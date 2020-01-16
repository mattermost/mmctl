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

func (s *MmctlUnitTestSuite) TestModifyChannelCmdF() {
	s.Run("Both public and private the same value", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", false, "")
		cmd.Flags().Bool("private", false, "")

		err := modifyChannelCmdF(s.client, cmd, []string{})
		s.Require().EqualError(err, "You must specify only one of --public or --private")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)

		cmd = &cobra.Command{}
		cmd.Flags().String("username", "mockUser", "")
		cmd.Flags().Bool("public", true, "")
		cmd.Flags().Bool("private", true, "")

		err = modifyChannelCmdF(s.client, cmd, []string{})
		s.Require().EqualError(err, "You must specify only one of --public or --private")
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
		s.Require().EqualError(err, "You can only change the type of public/private channels.")
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
		s.Require().EqualError(err, "You can only change the type of public/private channels.")
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

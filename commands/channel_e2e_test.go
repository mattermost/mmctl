// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)


func (s *MmctlE2ETestSuite) TestUnarchiveChannelsCmdF() {
	s.SetupTestHelper().InitBasic()

	s.Run("Unarchive channel", func() {
		printer.Clean()

		err := unarchiveChannelsCmdF(s.th.SystemAdminClient, &cobra.Command{}, []string{fmt.Sprintf("%s:%s", s.th.BasicTeam.Id, s.th.BasicDeletedChannel.Name)})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		channel, appErr := s.th.App.GetChannel(s.th.BasicDeletedChannel.Id)
		s.Require().Nil(appErr)
		s.Require().True(channel.IsOpen())
	})

	s.Run("Unarchive channel without permissions", func() {
		printer.Clean()

		err := unarchiveChannelsCmdF(s.th.Client, &cobra.Command{}, []string{fmt.Sprintf("%s:%s", s.th.BasicTeam.Id, s.th.BasicDeletedChannel.Name)})
		s.Require().Nil(err)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to unarchive channel '%s:%s'", s.th.BasicTeam.Id, s.th.BasicDeletedChannel.Name))
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions.")
	})

	s.RunForAllClients("Unarchive nonexistent channel", func(c client.Client) {
		printer.Clean()

		err := unarchiveChannelsCmdF(c, &cobra.Command{}, []string{fmt.Sprintf("%s:%s", s.th.BasicTeam.Id, "nonexistent-channel")})
		s.Require().Nil(err)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to find channel '%s:%s'", s.th.BasicTeam.Id, "nonexistent-channel"))
	})

	s.Run("Unarchive open channel", func() {
		printer.Clean()

		err := unarchiveChannelsCmdF(s.th.SystemAdminClient, &cobra.Command{}, []string{fmt.Sprintf("%s:%s", s.th.BasicTeam.Id, s.th.BasicChannel.Name)})
		s.Require().Nil(err)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to unarchive channel '%s:%s'", s.th.BasicTeam.Id, s.th.BasicChannel.Name))
		s.Require().Contains(printer.GetErrorLines()[0], "Unable to unarchive channel. The channel is not archived.")
	})
}

func (s *MmctlE2ETestSuite) TestMoveChannelCmd() {
	s.SetupTestHelper().InitBasic()
	initChannelName := api4.GenerateTestChannelName()
	var appErr *model.AppError
	var channel *model.Channel
	var team *model.Team
	channel, appErr = s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        initChannelName,
		DisplayName: "dName_" + initChannelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)

	s.RunForAllClients("Move nonexistent team", func(c client.Client) {
		printer.Clean()

		err := moveChannelCmdF(c, &cobra.Command{}, []string{"test"})
		s.Require().NotNil(err)
		s.Require().Equal(`unable to find destination team "test"`, err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Move existing channel to specified team", func(c client.Client) {
		printer.Clean()

		testTeamName := api4.GenerateTestTeamName()
		team, appErr = s.th.App.CreateTeam(&model.Team{
			Name:        testTeamName,
			DisplayName: "dName_" + testTeamName,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		args := []string{team.Id, channel.Id}
		cmd := &cobra.Command{}
		cmd.Flags().String("team", team.Id, "")
		cmd.Flags().String("channel", channel.Id, "")

		err := moveChannelCmdF(c, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		actualChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok)
		s.Require().Equal(channel.Name, actualChannel.Name)
	})

	s.RunForSystemAdminAndLocal("Moving team to non existing channel", func(c client.Client) {
		printer.Clean()

		args := []string{s.th.BasicTeam.Id, "no-channel"}
		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicTeam.Id, "")
		cmd.Flags().String("channel", "no-channel", "")

		err := moveChannelCmdF(c, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("Unable to find channel %q", "no-channel"), printer.GetErrorLines()[0])
	})

	s.RunForSystemAdminAndLocal("Moving channel which is already moved to particular team", func(c client.Client) {
		printer.Clean()

		s.SetupTestHelper().InitBasic()
		initChannelName := api4.GenerateTestChannelName()
		channel, appErr = s.th.App.CreateChannel(&model.Channel{
			TeamId:      s.th.BasicTeam.Id,
			Name:        initChannelName,
			DisplayName: "dName_" + initChannelName,
			Type:        model.CHANNEL_OPEN,
		}, false)
		s.Require().Nil(appErr)

		args := []string{channel.TeamId, channel.Id}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", channel.TeamId, "")
		cmd.Flags().String("channel", channel.Name, "")

		err := moveChannelCmdF(c, cmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Move existing channel to specified team should fail for client", func() {
		printer.Clean()

		testTeamName := api4.GenerateTestTeamName()
		team, appErr = s.th.App.CreateTeam(&model.Team{
			Name:        testTeamName,
			DisplayName: "dName_" + testTeamName,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		args := []string{team.Id, channel.Id}
		cmd := &cobra.Command{}
		cmd.Flags().String("team", team.Id, "")
		cmd.Flags().String("channel", channel.Id, "")

		err := moveChannelCmdF(s.th.Client, cmd, args)

		s.Require().Equal(fmt.Sprintf("unable to find destination team %q", team.Id), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestListChannelsCmdF() {
	s.SetupTestHelper().InitBasic()

	var assertChannelNames = func(want []string, lines []interface{}) {
		var got []string
		for i := 0; i < len(lines); i++ {
			got = append(got, lines[i].(*model.Channel).Name)
		}

		sort.Strings(want)
		sort.Strings(got)

		s.Equal(want, got)
	}

	s.Run("List channels/Client", func() {
		printer.Clean()
		wantNames := append(
			s.th.App.DefaultChannelNames(),
			[]string{
				s.th.BasicChannel.Name,
				s.th.BasicChannel2.Name,
				s.th.BasicDeletedChannel.Name,
				s.th.BasicPrivateChannel.Name,
			}...,
		)

		err := listChannelsCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Equal(6, len(printer.GetLines()))
		assertChannelNames(wantNames, printer.GetLines())
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("List channels", func(c client.Client) {
		printer.Clean()
		wantNames := append(
			s.th.App.DefaultChannelNames(),
			[]string{
				s.th.BasicChannel.Name,
				s.th.BasicChannel2.Name,
				s.th.BasicDeletedChannel.Name,
				s.th.BasicPrivateChannel.Name,
				s.th.BasicPrivateChannel2.Name,
			}...,
		)

		err := listChannelsCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Equal(7, len(printer.GetLines()))
		assertChannelNames(wantNames, printer.GetLines())
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("List channels for non existent team", func(c client.Client) {
		printer.Clean()
		team := "non-existent-team"

		err := listChannelsCmdF(c, &cobra.Command{}, []string{team})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+team+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlE2ETestSuite) TestSearchChannelCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Search nonexistent channel", func(c client.Client) {
		printer.Clean()

		err := searchChannelCmdF(c, &cobra.Command{}, []string{"test"})
		s.Require().NotNil(err)
		s.Require().Equal(`channel "test" was not found in any team`, err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Search existing channel", func(c client.Client) {
		printer.Clean()

		err := searchChannelCmdF(c, &cobra.Command{}, []string{s.th.BasicChannel.Name})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		actualChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok)
		s.Require().Equal(s.th.BasicChannel.Name, actualChannel.Name)
	})

	s.RunForSystemAdminAndLocal("Search existing channel of a team", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().String("team", s.th.BasicChannel.TeamId, "")

		err := searchChannelCmdF(c, cmd, []string{s.th.BasicChannel.Name})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		actualChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok)
		s.Require().Equal(s.th.BasicChannel.Name, actualChannel.Name)
	})

	s.RunForSystemAdminAndLocal("Search existing channel that does not belong to a team", func(c client.Client) {
		printer.Clean()

		testTeamName := api4.GenerateTestTeamName()

		team, appErr := s.th.App.CreateTeam(&model.Team{
			Name:        testTeamName,
			DisplayName: "dn_" + testTeamName,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		cmd := &cobra.Command{}
		cmd.Flags().String("team", team.Id, "")

		err := searchChannelCmdF(c, cmd, []string{s.th.BasicChannel.Name})
		s.Require().NotNil(err)
		s.Require().Equal(`: Channel does not exist., `, err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Search existing channel should fail for Client", func() {
		printer.Clean()

		err := searchChannelCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicChannel.Name})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("channel \"%s\" was not found in any team", s.th.BasicChannel.Name), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlE2ETestSuite) TestCreateChannelCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("create channel successfully", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		channelName := model.NewRandomString(10)
		teamName := s.th.BasicTeam.Name
		channelDisplayName := "channelDisplayName"
		cmd.Flags().String("name", channelName, "channel name")
		cmd.Flags().String("team", teamName, "team name")
		cmd.Flags().String("display_name", channelDisplayName, "display name")

		err := createChannelCmdF(c, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)

		printerChannel := printer.GetLines()[0].(*model.Channel)
		s.Require().Equal(channelName, printerChannel.Name)
		s.Require().Equal(s.th.BasicTeam.Id, printerChannel.TeamId)

		newChannel, err := s.th.App.GetChannelByName(channelName, s.th.BasicTeam.Id, false)
		s.Require().Nil(err)
		s.Require().Equal(channelName, newChannel.Name)
		s.Require().Equal(channelDisplayName, newChannel.DisplayName)
		s.Require().Equal(s.th.BasicTeam.Id, newChannel.TeamId)
	})

	s.RunForAllClients("create channel with nonexistent team", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		channelName := model.NewRandomString(10)
		teamName := "nonexistent team"
		channelDisplayName := "channelDisplayName"
		cmd.Flags().String("name", channelName, "channel name")
		cmd.Flags().String("team", teamName, "team name")
		cmd.Flags().String("display_name", channelDisplayName, "display name")

		err := createChannelCmdF(c, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Equal("unable to find team: "+teamName, err.Error())
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)

		_, err = s.th.App.GetChannelByName(channelName, s.th.BasicTeam.Id, false)
		s.Require().NotNil(err)
	})

	s.RunForAllClients("create channel with invalid name", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		channelName := "invalid name"
		teamName := s.th.BasicTeam.Name
		channelDisplayName := "channelDisplayName"
		cmd.Flags().String("name", channelName, "channel name")
		cmd.Flags().String("team", teamName, "team name")
		cmd.Flags().String("display_name", channelDisplayName, "display name")

		err := createChannelCmdF(c, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Equal(": Name must be 2 or more lowercase alphanumeric characters., ", err.Error())
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)

		_, err = s.th.App.GetChannelByName(channelName, s.th.BasicTeam.Id, false)
		s.Require().NotNil(err)
	})
}

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
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        initChannelName,
		DisplayName: "dName_" + initChannelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)

	s.RunForAllClients("Move nonexistent team", func(c client.Client) {
		printer.Clean()

		err := moveChannelCmdF(c, &cobra.Command{}, []string{"test"})
		s.Require().Error(err)
		s.Require().Equal(`unable to find destination team "test"`, err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Move existing channel to specified team", func(c client.Client) {
		printer.Clean()

		testTeamName := api4.GenerateTestTeamName()
		var team *model.Team
		team, appErr = s.th.App.CreateTeam(&model.Team{
			Name:        testTeamName,
			DisplayName: "dName_" + testTeamName,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		args := []string{team.Id, channel.Id}
		cmd := &cobra.Command{}

		err := moveChannelCmdF(c, cmd, args)

		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		actualChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok)
		s.Require().Equal(channel.Name, actualChannel.Name)
		s.Require().Equal(team.Id, actualChannel.TeamId)
	})

	s.RunForSystemAdminAndLocal("Moving team to non existing channel", func(c client.Client) {
		printer.Clean()

		args := []string{s.th.BasicTeam.Id, "no-channel"}
		cmd := &cobra.Command{}

		err := moveChannelCmdF(c, cmd, args)
		s.Require().NoError(err)
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

		err := moveChannelCmdF(c, cmd, args)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Move existing channel to specified team should fail for client", func() {
		printer.Clean()

		testTeamName := api4.GenerateTestTeamName()
		var team *model.Team
		team, appErr = s.th.App.CreateTeam(&model.Team{
			Name:        testTeamName,
			DisplayName: "dName_" + testTeamName,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		args := []string{team.Id, channel.Id}
		cmd := &cobra.Command{}

		err := moveChannelCmdF(s.th.Client, cmd, args)
		s.Require().Error(err)
		s.Require().Equal(fmt.Sprintf("unable to find destination team %q", team.Id), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

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

func (s *MmctlE2ETestSuite) TestChannelRenameCmd() {
	s.SetupTestHelper().InitBasic()

	initChannelName := api4.GenerateTestChannelName()
	initChannelDisplayName := "dn_" + initChannelName

	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        initChannelName,
		DisplayName: initChannelDisplayName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)

	s.RunForAllClients("Rename nonexistent channel", func(c client.Client) {
		printer.Clean()

		nonexistentChannelName := api4.GenerateTestChannelName()

		cmd := &cobra.Command{}
		cmd.Flags().String("name", "name", "")
		cmd.Flags().String("display_name", "name", "")

		err := renameChannelCmdF(c, cmd, []string{s.th.BasicTeam.Id + ":" + nonexistentChannelName})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("unable to find channel from \"%s:%s\"", s.th.BasicTeam.Id, nonexistentChannelName), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Rename channel", func(c client.Client) {
		printer.Clean()

		newChannelName := api4.GenerateTestChannelName()
		newChannelDisplayName := "dn_" + newChannelName

		cmd := &cobra.Command{}
		cmd.Flags().String("name", newChannelName, "")
		cmd.Flags().String("display_name", newChannelDisplayName, "")

		err := renameChannelCmdF(c, cmd, []string{s.th.BasicTeam.Id + ":" + channel.Id})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		printedChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok, "unexpected printer output type")

		s.Require().Equal(newChannelName, printedChannel.Name)
		s.Require().Equal(newChannelDisplayName, printedChannel.DisplayName)

		rchannel, err := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(err)
		s.Require().Equal(newChannelName, rchannel.Name)
		s.Require().Equal(newChannelDisplayName, rchannel.DisplayName)
	})

	s.Run("Rename channel without permission", func() {
		printer.Clean()

		channelInit, appErr := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(appErr)

		newChannelName := api4.GenerateTestChannelName()
		newChannelDisplayName := "dn_" + newChannelName

		cmd := &cobra.Command{}
		cmd.Flags().String("name", newChannelName, "")
		cmd.Flags().String("display_name", newChannelDisplayName, "")

		err := renameChannelCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Id + ":" + channel.Id})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(fmt.Sprintf("cannot rename channel \"%s\", error: : You do not have the appropriate permissions., ", channelInit.Name), err.Error())

		rchannel, err := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(err)
		s.Require().Equal(channelInit.Name, rchannel.Name)
		s.Require().Equal(channelInit.DisplayName, rchannel.DisplayName)
	})

	s.Run("Rename channel with permission", func() {
		printer.Clean()

		_, appErr := s.th.App.AddChannelMember(s.th.BasicUser.Id, channel, "", "")
		s.Require().Nil(appErr)

		newChannelName := api4.GenerateTestChannelName()
		newChannelDisplayName := "dn_" + newChannelName

		cmd := &cobra.Command{}
		cmd.Flags().String("name", newChannelName, "")
		cmd.Flags().String("display_name", newChannelDisplayName, "")

		err := renameChannelCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Id + ":" + channel.Id})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		printedChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok, "unexpected printer output type")

		s.Require().Equal(newChannelName, printedChannel.Name)
		s.Require().Equal(newChannelDisplayName, printedChannel.DisplayName)

		rchannel, err := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(err)
		s.Require().Equal(newChannelName, rchannel.Name)
		s.Require().Equal(newChannelDisplayName, rchannel.DisplayName)
	})
}

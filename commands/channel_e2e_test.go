// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

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

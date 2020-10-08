// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestChannelRenameCmd() {
	s.SetupTestHelper().InitBasic()

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

	s.Run("Rename channel without permission", func() {
		printer.Clean()

		initChannelName := api4.GenerateTestChannelName()
		initChannelDisplayName := "dn_" + initChannelName

		channel, appErr := s.th.App.CreateChannel(&model.Channel{
			TeamId:      s.th.BasicTeam.Id,
			Name:        initChannelName,
			DisplayName: initChannelDisplayName,
			Type:        model.CHANNEL_OPEN,
		}, false)
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
		s.Require().Equal(fmt.Sprintf("cannot rename channel \"%s\", error: : You do not have the appropriate permissions., ", channel.Name), err.Error())

		rchannel, err := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(err)
		s.Require().Equal(initChannelName, rchannel.Name)
		s.Require().Equal(initChannelDisplayName, rchannel.DisplayName)
	})

	s.RunForAllClients("Rename channel", func(c client.Client) {
		printer.Clean()

		newChannelName := api4.GenerateTestChannelName()
		newChannelDisplayName := "dn_" + newChannelName

		cmd := &cobra.Command{}
		cmd.Flags().String("name", newChannelName, "")
		cmd.Flags().String("display_name", newChannelDisplayName, "")

		err := renameChannelCmdF(c, cmd, []string{s.th.BasicTeam.Id + ":" + s.th.BasicChannel.Id})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		printedChannel, ok := printer.GetLines()[0].(*model.Channel)
		s.Require().True(ok, "unexpected printer output type")

		s.Require().Equal(newChannelName, printedChannel.Name)
		s.Require().Equal(newChannelDisplayName, printedChannel.DisplayName)

		rchannel, err := s.th.App.GetChannel(s.th.BasicChannel.Id)
		s.Require().Nil(err)
		s.Require().Equal(newChannelName, rchannel.Name)
		s.Require().Equal(newChannelDisplayName, rchannel.DisplayName)
	})
}

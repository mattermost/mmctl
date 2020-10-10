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

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestRenameTeamCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Error renaming team which does not exist", func(c client.Client) {
		printer.Clean()
		nonExistentTeamName := "existingName"
		cmd := &cobra.Command{}
		args := []string{nonExistentTeamName}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(c, cmd, args)
		s.Require().EqualError(err, "Unable to find team 'existingName', to see the all teams try 'team list' command")
	})

	s.RunForSystemAdminAndLocal("Rename an existing team", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{s.th.BasicTeam.Name}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(c, cmd, args)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal("'"+s.th.BasicTeam.Name+"' team renamed", printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Permission error renaming an existing team", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{s.th.BasicTeam.Name}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(s.th.Client, cmd, args)
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Equal("Cannot rename team '"+s.th.BasicTeam.Name+"', error : : You do not have the appropriate permissions., ", err.Error())
	})
}

func (s *MmctlE2ETestSuite) TestSearchTeamCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("Search for existing team", func(c client.Client) {
		printer.Clean()

		err := searchTeamCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		team := printer.GetLines()[0].(*model.Team)
		s.Equal(s.th.BasicTeam.Name, team.Name)
	})

	s.Run("Search for existing team with Client", func() {
		printer.Clean()

		err := searchTeamCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+s.th.BasicTeam.Name+"'", printer.GetErrorLines()[0])
	})

	s.RunForAllClients("Search of nonexistent team", func(c client.Client) {
		printer.Clean()

		teamnameArg := "nonexistentteam"
		err := searchTeamCmdF(c, &cobra.Command{}, []string{teamnameArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+teamnameArg+"'", printer.GetErrorLines()[0])
	})
}

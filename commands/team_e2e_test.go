// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-server/v5/model"
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

func (s *MmctlE2ETestSuite) TestDeleteTeamsCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Error deleting team which does not exist", func(c client.Client) {
		printer.Clean()
		nonExistentName := "existingName"
		cmd := &cobra.Command{}
		args := []string{nonExistentName}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")
		cmd.Flags().Bool("confirm", true, "")

		_ = deleteTeamsCmdF(c, cmd, args)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to find team '"+nonExistentName+"'", printer.GetErrorLines()[0])
	})

	s.Run("Permission error while deleting a valid team", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := []string{s.th.BasicTeam.Name}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")
		cmd.Flags().Bool("confirm", true, "")

		_ = deleteTeamsCmdF(s.th.Client, cmd, args)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to delete team '"+s.th.BasicTeam.Name+"' error: : You do not have the appropriate permissions., ", printer.GetErrorLines()[0])
		team, _ := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Equal(team.Name, s.th.BasicTeam.Name)
	})

	s.RunForSystemAdminAndLocal("Delete a valid team", func(c client.Client) {
		printer.Clean()

		teamName := "teamname" + model.NewRandomString(10)
		teamDisplayname := "Mock Display Name"
		cmd := &cobra.Command{}
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", teamDisplayname, "")
		err := createTeamCmdF(s.th.LocalClient, cmd, []string{})
		s.Require().Nil(err)

		cmd = &cobra.Command{}
		args := []string{teamName}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")
		cmd.Flags().Bool("confirm", true, "")

		// Set EnableAPITeamDeletion
		enableConfig := true
		config, _ := c.GetConfig()
		config.ServiceSettings.EnableAPITeamDeletion = &enableConfig
		_, _ = c.UpdateConfig(config)

		// Deletion should succeed for both local and SystemAdmin client now
		err = deleteTeamsCmdF(c, cmd, args)
		s.Require().Nil(err)
		team := printer.GetLines()[0].(*model.Team)
		s.Equal(teamName, team.Name)
		s.Len(printer.GetErrorLines(), 0)

		// Reset config
		enableConfig = false
		config, _ = c.GetConfig()
		config.ServiceSettings.EnableAPITeamDeletion = &enableConfig
		_, _ = c.UpdateConfig(config)
	})

	s.Run("Permission denied error for system admin when deleting a valid team", func() {
		printer.Clean()

		args := []string{s.th.BasicTeam.Name}
		cmd := &cobra.Command{}
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")
		cmd.Flags().Bool("confirm", true, "")

		// Delete should fail for SystemAdmin client
		err := deleteTeamsCmdF(s.th.SystemAdminClient, cmd, args)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to delete team '"+s.th.BasicTeam.Name+"' error: : Permanent team deletion feature is not enabled. Please contact your System Administrator., ", printer.GetErrorLines()[0])

		// verify team still exists
		team, _ := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Equal(team.Name, s.th.BasicTeam.Name)

		// Delete should succeed for local client
		printer.Clean()
		err = deleteTeamsCmdF(s.th.LocalClient, cmd, args)
		s.Require().Nil(err)
		team = printer.GetLines()[0].(*model.Team)
		s.Equal(team.Name, s.th.BasicTeam.Name)
		s.Len(printer.GetErrorLines(), 0)
	})
}

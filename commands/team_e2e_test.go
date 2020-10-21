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

func (s *MmctlE2ETestSuite) TestTeamCreateCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Should not create a team w/o name", func(c client.Client) {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("display_name", "somedisplayname", "")

		err := createTeamCmdF(c, cmd, []string{})
		s.EqualError(err, "name is required")
		s.Require().Empty(printer.GetLines())
	})

	s.RunForAllClients("Should not create a team w/o display_name", func(c client.Client) {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", model.NewId(), "")

		err := createTeamCmdF(c, cmd, []string{})
		s.EqualError(err, "display Name is required")
		s.Require().Empty(printer.GetLines())
	})

	s.Run("Should create a new team w/ email using LocalClient", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		teamName := model.NewId()
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", "somedisplayname", "")
		email := "someemail@example.com"
		cmd.Flags().String("email", email, "")

		err := createTeamCmdF(s.th.LocalClient, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		newTeam, err := s.th.App.GetTeamByName(teamName)
		s.Require().Nil(err)
		s.Equal(email, newTeam.Email)
	})

	s.Run("Should create a new team w/ assigned email using SystemAdminClient", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		teamName := model.NewId()
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", "somedisplayname", "")
		email := "someemail@example.com"
		cmd.Flags().String("email", email, "")

		err := createTeamCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		newTeam, err := s.th.App.GetTeamByName(teamName)
		s.Require().Nil(err)
		s.NotEqual(email, newTeam.Email)
	})

	s.Run("Should create a new team w/ assigned email using Client", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		teamName := model.NewId()
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", "somedisplayname", "")
		email := "someemail@example.com"
		cmd.Flags().String("email", email, "")

		err := createTeamCmdF(s.th.Client, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		newTeam, err := s.th.App.GetTeamByName(teamName)
		s.Require().Nil(err)
		s.NotEqual(email, newTeam.Email)
	})

	s.RunForAllClients("Should create a new open team", func(c client.Client) {
		printer.Clean()
		cmd := &cobra.Command{}
		teamName := model.NewId()
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", "somedisplayname", "")

		err := createTeamCmdF(c, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		newTeam, err := s.th.App.GetTeamByName(teamName)
		s.Require().Nil(err)
		s.Equal(newTeam.Type, model.TEAM_OPEN)
	})

	s.RunForAllClients("Should create a new private team", func(c client.Client) {
		printer.Clean()
		cmd := &cobra.Command{}
		teamName := model.NewId()
		cmd.Flags().String("name", teamName, "")
		cmd.Flags().String("display_name", "somedisplayname", "")
		cmd.Flags().Bool("private", true, "")

		err := createTeamCmdF(c, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		newTeam, err := s.th.App.GetTeamByName(teamName)
		s.Require().Nil(err)
		s.Equal(newTeam.Type, model.TEAM_INVITE)
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

func (s *MmctlE2ETestSuite) TestArchiveTeamsCmd() {
	s.SetupTestHelper().InitBasic()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("confirm", true, "Confirm you really want to archive the team and a DB backup has been performed.")

	s.RunForAllClients("Archive nonexistent team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{"unknown-team"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to find team 'unknown-team'", printer.GetErrorLines()[0])
	})

	s.RunForSystemAdminAndLocal("Archive basic team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		team := printer.GetLines()[0].(*model.Team)
		s.Require().Equal(s.th.BasicTeam.Name, team.Name)
		s.Require().Len(printer.GetErrorLines(), 0)

		basicTeam, err := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
		s.Require().NotZero(basicTeam.DeleteAt)

		err = s.th.App.RestoreTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
	})

	s.Run("Archive team without permissions", func() {
		printer.Clean()

		err := archiveTeamsCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions.")

		basicTeam, err := s.th.App.GetTeam(s.th.BasicTeam.Id)
		s.Require().Nil(err)
		s.Require().Zero(basicTeam.DeleteAt)
	})
}

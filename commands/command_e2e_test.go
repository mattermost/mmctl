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

func (s *MmctlE2ETestSuite) TestListCommandCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("List commands for a non existing team", func(c client.Client) {
		printer.Clean()

		nonexistentTeamID := "nonexistent-team-id"

		err := listCommandCmdF(c, &cobra.Command{}, []string{nonexistentTeamID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+nonexistentTeamID+"'", printer.GetErrorLines()[0])
	})

	s.RunForAllClients("List commands for a specific team", func(c client.Client) {
		printer.Clean()

		team, appErr := s.th.App.CreateTeam(&model.Team{
			DisplayName: "dn_" + model.NewId(),
			Name:        api4.GenerateTestTeamName(),
			Email:       s.th.BasicUser.Email,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.AddUserToTeam(team.Id, s.th.BasicUser.Id, "")
		s.Require().Nil(appErr)

		command, appErr := s.th.App.CreateCommand(&model.Command{
			DisplayName: "command",
			CreatorId:   s.th.BasicUser.Id,
			TeamId:      team.Id,
			URL:         "http://localhost:8000/example",
			Method:      model.COMMAND_METHOD_GET,
			Trigger:     "trigger",
		})
		s.Require().Nil(appErr)
		defer func() {
			appErr = s.th.App.DeleteCommand(command.Id)
			s.Require().Nil(appErr)
		}()

		err := listCommandCmdF(c, &cobra.Command{}, []string{team.Id})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(command, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("List all commands from all teams", func() {
		// add team1
		team1, appErr := s.th.App.CreateTeam(&model.Team{
			DisplayName: "dn_" + model.NewId(),
			Name:        api4.GenerateTestTeamName(),
			Email:       s.th.BasicUser.Email,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.AddUserToTeam(team1.Id, s.th.BasicUser.Id, "")
		s.Require().Nil(appErr)

		command1, appErr := s.th.App.CreateCommand(&model.Command{
			DisplayName: "command1",
			CreatorId:   s.th.BasicUser.Id,
			TeamId:      team1.Id,
			URL:         "http://localhost:8000/example",
			Method:      model.COMMAND_METHOD_GET,
			Trigger:     "trigger",
		})
		s.Require().Nil(appErr)
		defer func() {
			appErr = s.th.App.DeleteCommand(command1.Id)
			s.Require().Nil(appErr)
		}()

		// add team 2
		team2, appErr := s.th.App.CreateTeam(&model.Team{
			DisplayName: "dn_" + model.NewId(),
			Name:        api4.GenerateTestTeamName(),
			Email:       s.th.BasicUser.Email,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		_, appErr = s.th.App.AddUserToTeam(team2.Id, s.th.BasicUser.Id, "")
		s.Require().Nil(appErr)

		command2, appErr := s.th.App.CreateCommand(&model.Command{
			DisplayName: "command2",
			CreatorId:   s.th.BasicUser.Id,
			TeamId:      team2.Id,
			URL:         "http://localhost:8000/example",
			Method:      model.COMMAND_METHOD_GET,
			Trigger:     "trigger",
		})
		s.Require().Nil(appErr)
		defer func() {
			appErr = s.th.App.DeleteCommand(command2.Id)
			s.Require().Nil(appErr)
		}()

		s.RunForSystemAdminAndLocal("Run list command", func(c client.Client) {
			printer.Clean()

			err := listCommandCmdF(c, &cobra.Command{}, []string{})
			s.Require().Nil(err)
			s.Len(printer.GetLines(), 2)
			s.ElementsMatch([]*model.Command{command1, command2}, printer.GetLines())
			s.Len(printer.GetErrorLines(), 0)
		})
	})

	s.Run("List commands for a specific team without permission", func() {
		printer.Clean()

		team, appErr := s.th.App.CreateTeam(&model.Team{
			DisplayName: "dn_" + model.NewId(),
			Name:        api4.GenerateTestTeamName(),
			Email:       s.th.BasicUser.Email,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		command, appErr := s.th.App.CreateCommand(&model.Command{
			DisplayName: "command",
			CreatorId:   s.th.BasicUser.Id,
			TeamId:      team.Id,
			URL:         "http://localhost:8000/example",
			Method:      model.COMMAND_METHOD_GET,
			Trigger:     "trigger",
		})
		s.Require().Nil(appErr)
		defer func() {
			appErr = s.th.App.DeleteCommand(command.Id)
			s.Require().Nil(appErr)
		}()

		err := listCommandCmdF(s.th.Client, &cobra.Command{}, []string{team.Id})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+team.Id+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlE2ETestSuite) TestArchiveCommandCmdF() {
	s.SetupTestHelper().InitBasic()

	teamOfBasicUser, appErr := s.th.App.CreateTeam(&model.Team{
		DisplayName: "dn_" + model.NewId(),
		Name:        api4.GenerateTestTeamName(),
		Email:       s.th.BasicUser.Email,
		Type:        model.TEAM_OPEN,
	})
	s.Require().Nil(appErr)

	_, appErr = s.th.App.AddUserToTeam(teamOfBasicUser.Id, s.th.BasicUser.Id, "")
	s.Require().Nil(appErr)

	s.RunForAllClients("Archive nonexistent command", func(c client.Client) {
		printer.Clean()

		nonexistentCommandID := "nonexistent-command-id"

		err := archiveCommandCmdF(c, &cobra.Command{}, []string{nonexistentCommandID})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("Unable to archive command '%s' error: : Sorry, we could not find the page., There doesn't appear to be an api call for the url='/api/v4/commands/nonexistent-command-id'.  Typo? are you missing a team_id or user_id as part of the url?", nonexistentCommandID), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("Archive command", func(c client.Client) {
		printer.Clean()

		command, appErr := s.th.App.CreateCommand(&model.Command{
			TeamId:      teamOfBasicUser.Id,
			DisplayName: "command",
			Description: "command",
			Trigger:     api4.GenerateTestId(),
			URL:         "http://localhost:8000/example",
			CreatorId:   s.th.BasicUser.Id,
			Username:    s.th.BasicUser.Username,
			IconURL:     "http://localhost:8000/icon.ico",
			Method:      model.COMMAND_METHOD_GET,
		})
		s.Require().Nil(appErr)

		err := archiveCommandCmdF(c, &cobra.Command{}, []string{command.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(map[string]interface{}{"status": "ok"}, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)

		rcommand, err := s.th.App.GetCommand(command.Id)
		s.Require().NotNil(err)
		s.Require().Nil(rcommand)
		s.Require().Equal("SqlCommandStore.Get: Command does not exist., ", err.Error())
	})

	s.Run("Archive command without permission", func() {
		printer.Clean()

		teamOfAdminUser, appErr := s.th.App.CreateTeam(&model.Team{
			DisplayName: "dn_" + model.NewId(),
			Name:        api4.GenerateTestTeamName(),
			Email:       s.th.SystemAdminUser.Email,
			Type:        model.TEAM_OPEN,
		})
		s.Require().Nil(appErr)

		command, appErr := s.th.App.CreateCommand(&model.Command{
			TeamId:      teamOfAdminUser.Id,
			DisplayName: "command",
			Description: "command",
			Trigger:     api4.GenerateTestId(),
			URL:         "http://localhost:8000/example",
			CreatorId:   s.th.SystemAdminUser.Id,
			Username:    s.th.SystemAdminUser.Username,
			IconURL:     "http://localhost:8000/icon.ico",
			Method:      model.COMMAND_METHOD_GET,
		})
		s.Require().Nil(appErr)

		err := archiveCommandCmdF(s.th.Client, &cobra.Command{}, []string{command.Id})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("Unable to archive command '%s' error: : Unable to get the command., ", command.Id), err.Error())

		rcommand, err := s.th.App.GetCommand(command.Id)
		s.Require().Nil(err)
		s.Require().NotNil(rcommand)
		s.Require().Equal(int64(0), rcommand.DeleteAt)
	})
}

func (s *MmctlE2ETestSuite) TestModifyCommandCmdF() {
	s.SetupTestHelper().InitBasic()

	// create new command
	newCmd := &model.Command{
		CreatorId: s.th.BasicUser.Id,
		TeamId:    s.th.BasicTeam.Id,
		URL:       "http://nowhere.com",
		Method:    model.COMMAND_METHOD_POST,
		Trigger:   "trigger",
	}

	command, _ := s.th.SystemAdminClient.CreateCommand(newCmd)
	copy := command
	originalURL := newCmd.URL

	s.RunForAllClients("modifyCommandCmdF", func(c client.Client) {
		printer.Clean()

		command = copy
		s.Require().Equal(command.URL, originalURL)

		// Reset the cmd and parse to force Flag.Changed to be true.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addCommandFieldsFlags(cmd)
		url := command.URL + "_modified"
		err := cmd.ParseFlags([]string{
			command.Id,
			"--url=" + url,
		})
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.th.SystemAdminClient, cmd, []string{command.Id})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)

		check := getCommandFromCommandArg(s.th.SystemAdminClient, command.Id)
		s.Require().Equal(check.URL, url)
	})

	s.RunForAllClients("modifyCommandCmdF for command that does not exist", func(c client.Client) {
		printer.Clean()
		cmd := &cobra.Command{}

		err := modifyCommandCmdF(s.th.SystemAdminClient, cmd, []string{"nothing"})
		s.Require().NotNil(err)
		s.Require().Equal("unable to find command 'nothing'", err.Error())
	})

	s.RunForAllClients("modifyCommandCmdF with a space in trigger word", func(c client.Client) {
		printer.Clean()
		// Reset the cmd and parse to force Flag.Changed to be true.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addCommandFieldsFlags(cmd)
		err := cmd.ParseFlags([]string{
			command.Id,
			"--trigger-word=modified with space",
		})
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.th.SystemAdminClient, cmd, []string{command.Id})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "a trigger word must not contain spaces")
	})
}

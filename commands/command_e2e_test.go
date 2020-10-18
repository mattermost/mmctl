// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestListCommandCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("List commands for a non existing team", func(c client.Client) {
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
}

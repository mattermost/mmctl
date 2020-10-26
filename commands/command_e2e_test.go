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

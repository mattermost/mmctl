// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestCommandCreateCmd() {
	s.Run("Create a new custom slash command for a specified team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorIdArg := "example-user-id"
		creatorUsernameArg := "example-user"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg, Username: creatorUsernameArg}
		mockCommand := model.Command{
			TeamId:           teamArg,
			DisplayName:      titleArg,
			Description:      descriptionArg,
			Trigger:          triggerWordArg,
			URL:              urlArg,
			CreatorId:        creatorIdArg,
			Username:         responseUsernameArg,
			IconURL:          iconArg,
			Method:           method,
			AutoComplete:     autocomplete,
			AutoCompleteDesc: autocompleteDesc,
			AutoCompleteHint: autocompleteHint,
		}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")
		cmd.Flags().String("title", titleArg, "")
		cmd.Flags().String("description", descriptionArg, "")
		cmd.Flags().String("trigger-word", triggerWordArg, "")
		cmd.Flags().String("url", urlArg, "")
		cmd.Flags().String("creator", creatorIdArg, "")
		cmd.Flags().String("response-username", responseUsernameArg, "")
		cmd.Flags().String("icon", iconArg, "")
		cmd.Flags().String("method", method, "")
		cmd.Flags().Bool("autocomplete", autocomplete, "")
		cmd.Flags().String("autocompleteDesc", autocompleteDesc, "")
		cmd.Flags().String("autocompleteHint", autocompleteHint, "")

		// createCommandCmdF will call getTeamFromTeamArg,  getUserFromUserArg which then calls GetUserByEmail
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorIdArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			CreateCommand(&mockCommand).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockCommand, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a slash command only providing team, trigger word, url, creator", func() {
		printer.Clean()
		teamArg := "example-team-id"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorIdArg := "example-user-id"
		creatorUsernameArg := "example-user"
		method := "G"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg, Username: creatorUsernameArg}
		mockCommand := model.Command{
			TeamId:    teamArg,
			Trigger:   triggerWordArg,
			URL:       urlArg,
			CreatorId: creatorIdArg,
			Method:    method,
		}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")
		cmd.Flags().String("trigger-word", triggerWordArg, "")
		cmd.Flags().String("url", urlArg, "")
		cmd.Flags().String("creator", creatorIdArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorIdArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			CreateCommand(&mockCommand).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockCommand, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create slash command for a nonexistent team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "unable to find team '"+teamArg+"'")
	})

	s.Run("Create slash command with a space in trigger word", func() {
		printer.Clean()
		teamArg := "example-team-id"
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example    trigger    word"
		urlArg := "http://localhost:8000/example"
		creatorIdArg := "example-user-id"
		creatorUsernameArg := "example-user"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg, Username: creatorUsernameArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")
		cmd.Flags().String("title", titleArg, "")
		cmd.Flags().String("description", descriptionArg, "")
		cmd.Flags().String("trigger-word", triggerWordArg, "")
		cmd.Flags().String("url", urlArg, "")
		cmd.Flags().String("creator", creatorIdArg, "")
		cmd.Flags().String("response-username", responseUsernameArg, "")
		cmd.Flags().String("icon", iconArg, "")
		cmd.Flags().String("method", method, "")
		cmd.Flags().Bool("autocomplete", autocomplete, "")
		cmd.Flags().String("autocompleteDesc", autocompleteDesc, "")
		cmd.Flags().String("autocompleteHint", autocompleteHint, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorIdArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "a trigger word must not contain spaces")
	})

	s.Run("Create slash command with trigger word prefixed with /", func() {
		printer.Clean()
		teamArg := "example-team-id"
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "/example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorIdArg := "example-user-id"
		creatorUsernameArg := "example-user"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg, Username: creatorUsernameArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")
		cmd.Flags().String("title", titleArg, "")
		cmd.Flags().String("description", descriptionArg, "")
		cmd.Flags().String("trigger-word", triggerWordArg, "")
		cmd.Flags().String("url", urlArg, "")
		cmd.Flags().String("creator", creatorIdArg, "")
		cmd.Flags().String("response-username", responseUsernameArg, "")
		cmd.Flags().String("icon", iconArg, "")
		cmd.Flags().String("method", method, "")
		cmd.Flags().Bool("autocomplete", autocomplete, "")
		cmd.Flags().String("autocompleteDesc", autocompleteDesc, "")
		cmd.Flags().String("autocompleteHint", autocompleteHint, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorIdArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "a trigger word cannot begin with a /")
	})

	s.Run("Create slash command fail", func() {
		printer.Clean()
		teamArg := "example-team-id"
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorIdArg := "example-user-id"
		creatorUsernameArg := "example-user"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg, Username: creatorUsernameArg}
		mockCommand := model.Command{
			TeamId:           teamArg,
			DisplayName:      titleArg,
			Description:      descriptionArg,
			Trigger:          triggerWordArg,
			URL:              urlArg,
			CreatorId:        creatorIdArg,
			Username:         responseUsernameArg,
			IconURL:          iconArg,
			Method:           method,
			AutoComplete:     autocomplete,
			AutoCompleteDesc: autocompleteDesc,
			AutoCompleteHint: autocompleteHint,
		}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")
		cmd.Flags().String("title", titleArg, "")
		cmd.Flags().String("description", descriptionArg, "")
		cmd.Flags().String("trigger-word", triggerWordArg, "")
		cmd.Flags().String("url", urlArg, "")
		cmd.Flags().String("creator", creatorIdArg, "")
		cmd.Flags().String("response-username", responseUsernameArg, "")
		cmd.Flags().String("icon", iconArg, "")
		cmd.Flags().String("method", method, "")
		cmd.Flags().Bool("autocomplete", autocomplete, "")
		cmd.Flags().String("autocompleteDesc", autocompleteDesc, "")
		cmd.Flags().String("autocompleteHint", autocompleteHint, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorIdArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)
		mockError := &model.AppError{Message: "Mock Error, simulated error for CreateCommand"}
		s.client.
			EXPECT().
			CreateCommand(&mockCommand).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "unable to create command '"+mockCommand.DisplayName+"'. "+mockError.Error())
	})
}

func (s *MmctlUnitTestSuite) TestDeleteCommandCmd() {
	s.Run("Delete without errors", func() {
		printer.Clean()
		arg := "cmd1"
		outputMessage := map[string]interface{}{"status": "ok"}

		s.client.
			EXPECT().
			DeleteCommand(arg).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := deleteCommandCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Not able to delete", func() {
		printer.Clean()
		arg := "cmd1"
		outputMessage := map[string]interface{}{"status": "error"}

		s.client.
			EXPECT().
			DeleteCommand(arg).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := deleteCommandCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete with response error", func() {
		printer.Clean()
		arg := "cmd1"
		mockError := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			DeleteCommand(arg).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := deleteCommandCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().NotNil(err)
		s.Require().Equal(err, errors.New("Unable to delete command '"+arg+"' error: "+mockError.Error()))
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestCommandListCmdF() {
	s.Run("List all commands from all teams", func() {
		printer.Clean()
		team1ID := "team-id-1"
		team2Id := "team-id-2"

		commandTeam1ID := "command-team1-id"
		commandTeam2Id := "command-team2-id"
		teams := []*model.Team{
			&model.Team{Id: team1ID},
			&model.Team{Id: team2Id},
		}

		team1Commands := []*model.Command{
			&model.Command{
				Id: commandTeam1ID,
			},
		}
		team2Commands := []*model.Command{
			&model.Command{
				Id: commandTeam2Id,
			},
		}

		cmd := &cobra.Command{}
		s.client.EXPECT().GetAllTeams("", 0, 10000).Return(teams, &model.Response{Error: nil}).Times(1)
		s.client.EXPECT().ListCommands(team1ID, true).Return(team1Commands, &model.Response{Error: nil}).Times(1)
		s.client.EXPECT().ListCommands(team2Id, true).Return(team2Commands, &model.Response{Error: nil}).Times(1)
		err := listCommandCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 2)
		s.Equal(team1Commands[0], printer.GetLines()[0])
		s.Equal(team2Commands[0], printer.GetLines()[1])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("List commands for a specific team", func() {
		printer.Clean()
		teamID := "team-id"
		commandID := "command-id"
		team := &model.Team{Id: teamID}
		teamCommand := []*model.Command{
			&model.Command{
				Id: commandID,
			},
		}

		cmd := &cobra.Command{}
		s.client.EXPECT().GetTeam(teamID, "").Return(team, &model.Response{Error: nil}).Times(1)
		s.client.EXPECT().ListCommands(teamID, true).Return(teamCommand, &model.Response{Error: nil}).Times(1)
		err := listCommandCmdF(s.client, cmd, []string{teamID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(teamCommand[0], printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("List commands for a non existing team", func() {
		teamID := "non-existing-team"
		printer.Clean()
		cmd := &cobra.Command{}
		// first try to get team by id
		s.client.EXPECT().GetTeam(teamID, "").Return(nil, &model.Response{Error: nil}).Times(1)
		// second try to search the team by name
		s.client.EXPECT().GetTeamByName(teamID, "").Return(nil, &model.Response{Error: nil}).Times(1)
		err := listCommandCmdF(s.client, cmd, []string{teamID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+teamID+"'", printer.GetErrorLines()[0])
	})

	s.Run("Failling to list commands for an existing team", func() {
		teamID := "team-id"
		printer.Clean()
		cmd := &cobra.Command{}
		team := &model.Team{Id: teamID}
		s.client.EXPECT().GetTeam(teamID, "").Return(team, &model.Response{Error: nil}).Times(1)
		s.client.EXPECT().ListCommands(teamID, true).Return(nil, &model.Response{Error: &model.AppError{}}).Times(1)
		err := listCommandCmdF(s.client, cmd, []string{teamID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to list commands for '"+teamID+"'", printer.GetErrorLines()[0])
	})
}

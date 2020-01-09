package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func (s *MmctlUnitTestSuite) helperCreateCommand(*model.Command) {
	s.T().Helper()

}

func (s *MmctlUnitTestSuite) TestCommandModifyCmd() {
	arg := "cmd1"
	teamId := "example-team-id"
	titleArg := "example-command-name"
	descriptionArg := "example-description-text"
	triggerWordArg := "example-trigger-word"
	urlArg := "http://localhost:8000/example"
	creatorIdArg := "example-user-id"
	responseUsernameArg := "example-username2"
	iconArg := "icon-url"
	method := "G"
	autocomplete := false
	autocompleteDesc := "autocompleteDesc"
	autocompleteHint := "autocompleteHint"

	mockCommand := model.Command{
		TeamId:           teamId,
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

	s.Run("Modify a custom slash command by id", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.DisplayName = titleArg + "_modified"
		mockCommandModified.Description = descriptionArg + "_modified"
		mockCommandModified.Trigger = triggerWordArg + "_modified"
		mockCommandModified.URL = urlArg + "_modified"
		mockCommandModified.CreatorId = creatorIdArg + "_modified"
		mockCommandModified.Username = responseUsernameArg + "_modified"
		mockCommandModified.IconURL = iconArg + "_modified"
		mockCommandModified.Method = method
		mockCommandModified.AutoComplete = !autocomplete
		mockCommandModified.AutoCompleteDesc = autocompleteDesc + "_modified"
		mockCommandModified.AutoCompleteHint = autocompleteHint + "_modified"

		cli := []string{
			arg,
			"--title=" + mockCommandModified.DisplayName,
			"--description=" + mockCommandModified.Description,
			"--trigger-word=" + mockCommandModified.Trigger,
			"--url=" + mockCommandModified.URL,
			"--creator=" + mockCommandModified.CreatorId,
			"--response-username=" + mockCommandModified.Username,
			"--icon=" + mockCommandModified.IconURL,
			"--autocomplete=" + strconv.FormatBool(mockCommandModified.AutoComplete),
			"--autocompleteDesc=" + mockCommandModified.AutoCompleteDesc,
			"--autocompleteHint=" + mockCommandModified.AutoCompleteHint,
			"--post=" + strconv.FormatBool(method2Bool(mockCommandModified.Method)),
		}

		// modifyCommandCmdF will call getCommandById, GetUserByEmail and UpdateCommand
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(mockCommandModified.CreatorId, "").
			Return(&model.User{Id: mockCommandModified.CreatorId}, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateCommand(&mockCommand).
			Return(mockCommandModified, &model.Response{Error: nil}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(mockCommandModified, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Modify slash command using a nonexistent commandID", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.DisplayName = titleArg + "_modified"

		cli := []string{
			arg,
			"--title=" + mockCommandModified.DisplayName,
		}

		// modifyCommandCmdF will call getCommandById
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true for all flags on the CLI.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "unable to find command '"+arg+"'")
	})

	s.Run("Modify slash command with invalid user name", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.CreatorId = creatorIdArg + "_modified"

		bogusUsername := "bogus"
		cli := []string{
			arg,
			"--creator=" + bogusUsername,
		}

		// modifyCommandCmdF will call getCommandById, then try looking up user
		// via email, username, and id.
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(bogusUsername, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(bogusUsername, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(bogusUsername, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true for all flags on the CLI.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "unable to find user '"+bogusUsername+"'")
	})

	s.Run("Modify slash command with a space in trigger word", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.Trigger = creatorIdArg + " modified with space"

		cli := []string{
			arg,
			"--trigger-word=" + mockCommandModified.Trigger,
		}

		// modifyCommandCmdF will call getCommandById
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true for all flags on the CLI.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "a trigger word must not contain spaces")
	})

	s.Run("Create slash command with trigger word prefixed with /", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.Trigger = "/modified_with_slash"

		cli := []string{
			arg,
			"--trigger-word=" + mockCommandModified.Trigger,
		}

		// modifyCommandCmdF will call getCommandById
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true for all flags on the CLI.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "a trigger word cannot begin with a /")
	})

	s.Run("Create slash command fail", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.Trigger = creatorIdArg + "_modified"

		cli := []string{
			arg,
			"--trigger-word=" + mockCommandModified.Trigger,
		}

		// modifyCommandCmdF will call getCommandById then UpdateCommand
		s.client.
			EXPECT().
			GetCommandById(arg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		mockError := &model.AppError{Message: "Mock Error, simulated error for CreateCommand"}
		s.client.
			EXPECT().
			UpdateCommand(&mockCommand).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		// Reset the cmd and parse to force Flag.Changed to be true for all flags on the CLI.
		cmd := CommandModifyCmd
		cmd.ResetFlags()
		addFlags(cmd)
		err := cmd.ParseFlags(cli)
		s.Require().Nil(err)

		err = modifyCommandCmdF(s.client, cmd, []string{arg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "unable to modify command '"+mockCommand.DisplayName+"'. "+mockError.Error())
	})

}

func method2Bool(method string) bool {
	switch strings.ToUpper(method) {
	case "P":
		return true
	case "G":
		return false
	default:
		panic(fmt.Errorf("invalid method '%s'", method))
	}
}

func copyCommand(cmd *model.Command) *model.Command {
	json := cmd.ToJson()
	r := strings.NewReader(json)
	return model.CommandFromJson(r)
}

func (s *MmctlUnitTestSuite) TestCommandMoveCmd() {
	commandArg := "cmd1"
	commandArgBogus := "bogus-command-id"
	teamArg := "example-team-id"
	teamArgBogus := "bogus-team-id"

	mockTeam := model.Team{Id: "orig-team-id"}

	mockCommand := model.Command{
		Id:          commandArg,
		TeamId:      mockTeam.Id,
		DisplayName: "example-title",
		Trigger:     "example-trigger",
	}

	mockError := &model.AppError{Message: "Mock Error"}
	outputMessageOK := map[string]interface{}{"status": "ok"}
	outputMessageError := map[string]interface{}{"status": "error"}

	s.Run("Move custom slash command to another team by id", func() {
		printer.Clean()
		mockCommandModified := copyCommand(&mockCommand)
		mockCommandModified.TeamId = teamArg

		// moveCommandCmdF will look up team by id then name, call getCommandById and UpdateCommand
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetCommandById(commandArg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateCommand(&mockCommand).
			Return(mockCommandModified, &model.Response{Error: nil}).
			Times(1)

		err := moveCommandCmdF(s.client, &cobra.Command{}, []string{teamArg, commandArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessageOK)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Move custom slash command to invalid team by id", func() {
		printer.Clean()
		// moveCommandCmdF will look up team by id then name
		s.client.
			EXPECT().
			GetTeam(teamArgBogus, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(teamArgBogus, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := moveCommandCmdF(s.client, &cobra.Command{}, []string{teamArgBogus, commandArg})
		s.Require().NotNil(err)
		s.EqualError(err, "unable to find team '"+teamArgBogus+"'")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Move custom slash command to different team by invalid id", func() {
		printer.Clean()
		// moveCommandCmdF will look up team by id, then call GetCommandById
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetCommandById(commandArgBogus).
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := moveCommandCmdF(s.client, &cobra.Command{}, []string{teamArg, commandArgBogus})
		s.Require().NotNil(err)
		s.EqualError(err, "unable to find command '"+commandArgBogus+"'")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Unable to move custom slash command", func() {
		printer.Clean()
		// moveCommandCmdF will look up team by id, then call GetCommandById
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetCommandById(commandArgBogus).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateCommand(&mockCommand).
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := moveCommandCmdF(s.client, &cobra.Command{}, []string{teamArg, commandArgBogus})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessageError)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Move custom slash command with response error", func() {
		printer.Clean()
		// moveCommandCmdF will look up team by id, then call GetCommandById
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetCommandById(commandArg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateCommand(&mockCommand).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := moveCommandCmdF(s.client, &cobra.Command{}, []string{teamArg, commandArg})
		s.Require().NotNil(err)
		s.Require().EqualError(err, "unable to move command '"+commandArg+"'. "+mockError.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestCommandShowCmd() {
	commandArg := "example-command-id"
	commandArgBogus := "bogus-command-id"

	mockCommand := model.Command{
		Id:               commandArg,
		TeamId:           "example-team-id",
		DisplayName:      "example-command-name",
		Description:      "example-description-text",
		Trigger:          "example-trigger-word",
		URL:              "http://localhost:8000/example",
		CreatorId:        "example-user-id",
		Username:         "example-username2",
		IconURL:          "http://mydomain/example-icon-url",
		Method:           "G",
		AutoComplete:     false,
		AutoCompleteDesc: "example autocomplete description",
		AutoCompleteHint: "autocompleteHint",
	}

	s.Run("Show custom slash command", func() {
		printer.Clean()

		// showCommandCmdF will look up command by id
		s.client.
			EXPECT().
			GetCommandById(commandArg).
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := showCommandCmdF(s.client, &cobra.Command{}, []string{commandArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Equal(&mockCommand, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Show custom slash command with invalid id", func() {
		printer.Clean()
		// showCommandCmdF will look up command by id
		s.client.
			EXPECT().
			GetCommandById(commandArgBogus).
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := showCommandCmdF(s.client, &cobra.Command{}, []string{commandArgBogus})
		s.Require().NotNil(err)
		s.EqualError(err, "unable to find command '"+commandArgBogus+"'")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

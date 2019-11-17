package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestCreateCommandCmdF() {
	s.Run("Create a new custom slash command for a specified team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorUsernameArg := "example-username"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockCreator := model.User{Username: creatorUsernameArg}
		mockCommand := model.Command{
			CreatorId:        creatorUsernameArg,
			TeamId:           teamArg,
			Trigger:          triggerWordArg,
			Username:         responseUsernameArg,
			IconURL:          iconArg,
			DisplayName:      titleArg,
			Description:      descriptionArg,
			URL:              urlArg,
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
		cmd.Flags().String("creator", creatorUsernameArg, "")
		cmd.Flags().String("response-username", responseUsernameArg, "")
		cmd.Flags().String("icon", iconArg, "")

		// createCommandCmdF will call getTeamFromTeamArg,  getUserFromUserArg,
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByEmail(creatorUsernameArg, "").
			Return(&mockCreator, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			CreateCommand(cmd).		// ERROR ; passing in the wrong argument???
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockCommand, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	/*
	s.Run("List commands for a existing team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")


		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorUsernameArg := "example-username"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"
		mockCommand := model.Command{
			CreatorId:        creatorUsernameArg,
			TeamId:           teamArg,
			Trigger:          triggerWordArg,
			Username:         responseUsernameArg,
			IconURL:          iconArg,
			DisplayName:      titleArg,
			Description:      descriptionArg,
			URL:              urlArg,
			Method:           method,
			AutoComplete:     autocomplete,
			AutoCompleteDesc: autocompleteDesc,
			AutoCompleteHint: autocompleteHint,
		}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			ListCommands(teamArg, "").
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := listCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete commands for a existing team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("team", teamArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		err := deleteCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
	}) */
}

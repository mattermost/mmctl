package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestCreateCommandCmdF() {
	s.Run("Create a new custom slash command for a specified team", func() {
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}
		titleArg := "example-command-name"
		descriptionArg := "example-description-text"
		triggerWordArg := "example-trigger-word"
		urlArg := "http://localhost:8000/example"
		creatorUsernameArg := "example-username"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"

		mockCommand := model.Command{
			CreatorId:        creatorUsernameArg,
			TeamId:           teamArg,
			Trigger:          triggerWordArg,
			Username:         responseUsernameArg,
			IconURL:          iconArg,
			DisplayName:      titleArg,
			Description:      descriptionArg,
			URL:              urlArg,
			Method:           method,		// it's 'P' or 'G'
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

		// get the mock team
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		
		err := createCommandCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&mockChannel, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})


	s.Run("List commands for a existing team", func() {
		teamArg := "example-team-id"

		cmd := &cobra.Command{}
		// cmd.Flags().String("team", teamArg, "")

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)
		
		err := listCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete commands for a existing team", func() {
		teamArg := "example-team-id"

		cmd := &cobra.Command{}
		// cmd.Flags().String("team", teamArg, "")

		
		err := deleteCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
	})
}

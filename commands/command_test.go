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
		creatorIdArg := "example-user-id"
		responseUsernameArg := "example-username2"
		iconArg := "icon-url"
		method := "G"
		autocomplete := false
		autocompleteDesc := "autocompleteDesc"
		autocompleteHint := "autocompleteHint"

		mockTeam := model.Team{Id: teamArg}
		mockUser := model.User{Id: creatorIdArg}
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
			CreateCommand(&mockCommand).	// still gives an error!
			Return(&mockCommand, &model.Response{Error: nil}).
			Times(1)

		err := createCommandCmdF(s.client, cmd, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
	})
}

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestRenameTeamCmdF() {

	existingDisplayName := "Existing Display Name"
	existingName := "existingteamname"

	foundTeam := &model.Team{
		Id:             "pm695ajd5pdotqs46144rcejnc",
		CreateAt:       1574191499747,
		UpdateAt:       1575551058238,
		DeleteAt:       0,
		DisplayName:    existingDisplayName,
		Name:           existingName,
		Description:    "",
		Email:          "sampleemail@emailhost.com",
		Type:           "O",
		CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
		AllowedDomains: "",
		InviteId:       "",
	}

	newDisplayName := "New Display Name"
	newName := "newteamname"

	newTeam := &model.Team{
		Id:             "pm695ajd5pdotqs46144rcejnc",
		CreateAt:       1574191499747,
		UpdateAt:       1575551058238,
		DeleteAt:       0,
		DisplayName:    newDisplayName,
		Name:           newName,
		Description:    "",
		Email:          "sampleemail@emailhost.com",
		Type:           "O",
		CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
		AllowedDomains: "",
		InviteId:       "",
	}

	mockError := model.NewAppError("at-random-location.go", "Mock Error", nil, "mocking a random error", 0)

	s.Run("Team rename without existing and new name arguments", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := make([]string, 2)
		args[0] = ""
		args[1] = ""

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Error: requires at least 2 arg(s), only received 0")
	})

	s.Run("Team rename without new team name argument", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = existingName //Existing team name
		args[1] = ""           //New team name

		s.client.
			EXPECT().
			GetTeam(args[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(args[0], "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Error: required at least 2 arg(s), only received 1, If you like to change only display name; pass '-' after existing team name")
	})

	s.Run("Team rename without display name flag", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = existingName //Existing team name
		args[1] = newName      //New team name

		s.client.
			EXPECT().
			GetTeam(args[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(args[0], "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "missing display name, append '--display_name' flag to your command")
	})

	s.Run("Team rename with invalid display name flag", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = existingName //Existing team name
		args[1] = newName      //New team name

		// Setting flag as display-name instead of display_name
		cmd.Flags().String("display-name", newDisplayName, "Team Display Name")

		s.client.
			EXPECT().
			GetTeam(args[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(args[0], "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "missing display name, append '--display_name' flag to your command")
	})

	s.Run("Team rename with unknown existing team name", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = existingName //Non Existing team name
		args[1] = newName      //New team name

		// GetTeam searches with team id, if team not found proceeds to with team name search
		s.client.
			EXPECT().
			GetTeam(existingName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// GetTeamByname is called, if GetTeam fails to return any team, as team name was passed instead of team id
		s.client.
			EXPECT().
			GetTeamByName(existingName, "").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Unable to find team '"+existingName+"', to see the all teams try 'team list' command")
	})

	s.Run("Team rename when api fails to rename", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = existingName //Existing team name
		args[1] = newName      //New team name

		cmd.Flags().String("display_name", newDisplayName, "Some Display Name")

		s.client.
			EXPECT().
			GetTeam(args[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(args[0], "").
			Return(foundTeam, &model.Response{Error: nil}).
			Times(1)

		// Some UN-foreseeable error from the api
		// Mock out UpdateTeam which calls the api to rename team
		s.client.
			EXPECT().
			UpdateTeam(newTeam).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Cannot rename team '"+existingName+"', error : at-random-location.go: Mock Error, mocking a random error")
	})

	// s.Run("Team rename should work as expected", func(){}

	// s.Run("Team rename should work as expected even if same name as existing is supplied", func(){}

	// s.Run("Team rename should work as expected even if new name is -", func(){}

}

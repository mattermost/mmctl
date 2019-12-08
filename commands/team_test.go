package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestRenameTeamCmdF() {

	s.Run("Team rename should fail without existing and new name arguments", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		args := make([]string, 2)
		args[0] = ""
		args[1] = ""

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Error: requires at least 2 arg(s), only received 0")
	})

	s.Run("Team rename should fail without new team name argument", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName" //Existing team name
		args[1] = ""             //New team name

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Error: required at least 2 arg(s), only received 1, If you like to change only display name; pass '-' after existing team name")
	})

	s.Run("Team rename should fail without display name flag", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName" //Existing team name
		args[1] = "newName"      //New team name

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Missing display name, append '--display_name' flag to your command")
	})

	s.Run("Team rename should fail with invalid display name flag", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName" //Existing team name
		args[1] = "newName"      //New team name

		// Setting flag as display-name instead of display_name
		cmd.Flags().String("display-name", "newDisplayName", "Team Display Name")

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Missing display name, append '--display_name' flag to your command")
	})

	s.Run("Team rename should fail with unknown existing team name is entered", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName"
		args[1] = "newName"
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		// GetTeam searches with team id, if team not found proceeds to with team name search
		s.client.
			EXPECT().
			GetTeam("existingName", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// GetTeamByname is called, if GetTeam fails to return any team, as team name was passed instead of team id
		s.client.
			EXPECT().
			GetTeamByName("existingName", "").
			Return(nil, &model.Response{Error: nil}). // Error is nil as not found will not return error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Unable to find team 'existingName', to see the all teams try 'team list' command")
	})

	s.Run("Team rename should fail when no new inputs are passed", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName"
		args[1] = "existingName" //Same name

		cmd.Flags().String("display_name", "existingDisplayName", "Display Name")

		sameTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "existingDisplayName",
			Name:           "existingName",
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		s.client.
			EXPECT().
			GetTeam(args[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(args[0], "").
			Return(sameTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to rename, entered display name and name are same for team")
	})

	s.Run("Team rename should fail when api fails to rename", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := make([]string, 2)
		args[0] = "existingName" //Existing team name
		args[1] = "newTeamName"  //New team name

		cmd.Flags().String("display_name", "newDisplayName", "Display Name")

		foundTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "existingDisplayName",
			Name:           "existingteamname",
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}
		renamedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "newDisplayName",
			Name:           "newTeamName",
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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
		mockError := model.NewAppError("at-random-location.go", "Mock Error", nil, "mocking a random error", 0)

		// Mock out UpdateTeam which calls the api to rename team
		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Cannot rename team '"+"existingName"+"', error : at-random-location.go: Mock Error, mocking a random error")
	})

	s.Run("Team rename should fail when hyphen as name argument is passed but api didnt update display name", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newName := "-"
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

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

		renamedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    newDisplayName,
			Name:           "-", // Since same name
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    existingDisplayName, // Display name not changed
			Name:           existingName,        // Name not changed
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}). // No error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to rename display name of '"+existingName+"'")
	})

	s.Run("Team rename should fail when same name argument is passed but api although succeded but didnt actually update display name", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newName := "existingTeamName" // Same existing name
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

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

		renamedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    newDisplayName,
			Name:           "-", // Since same name
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    existingDisplayName, // Display name not changed
			Name:           existingName,        // Name not changed
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}). // No error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to rename display name of '"+existingName+"'")
	})

	s.Run("Team rename should fail when api succeded to update name but not display name", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newName := "newTeamName"
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

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

		renamedTeam := &model.Team{
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

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    existingDisplayName, // Display name not changed
			Name:           newName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}). // No error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to rename display name of '"+existingName+"'")
	})

	s.Run("Team rename should fail when api succeded but didnt update name and display name", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newName := "newTeamName"
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

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

		renamedTeam := &model.Team{
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

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    existingDisplayName, // Display name not changed
			Name:           existingName,        // name not changed
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}). // No error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Failed to rename team '"+existingName+"'")
	})

	s.Run("Team rename should work as expected", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		newName := "newTeamName"
		newDisplayName := "newDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Some Display Name")

		foundTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "Existing Display Name",
			Name:           existingName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}
		updatedTeam := &model.Team{
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

		s.client.
			EXPECT().
			UpdateTeam(updatedTeam).
			Return(updatedTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Successfully renamed team '"+existingName+"'")
	})

	s.Run("Team rename should work as expected even if same name as existing is supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = existingName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

		foundTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "Existing Display Name",
			Name:           existingName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		renamedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    newDisplayName,
			Name:           "-", // As '-' needs to be passed to API if name is not being renamed
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    newDisplayName,
			Name:           existingName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Successfully renamed team '"+existingName+"'")
	})

	s.Run("Team rename should work as expected even if only display name is supplied along with existing name", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		newName := "-"
		newDisplayName := "NewDisplayName"

		args := make([]string, 2)
		args[0] = existingName
		args[1] = newName

		cmd.Flags().String("display_name", newDisplayName, "Display Name")

		foundTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    "Existing Display Name",
			Name:           existingName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

		renamedTeam := &model.Team{
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

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    newDisplayName,
			Name:           existingName,
			Description:    "",
			Email:          "sampleemail@emailhost.com",
			Type:           "O",
			CompanyName:    "pk1qtd1hnbyhbbk79cwshxc6se",
			AllowedDomains: "",
			InviteId:       "",
		}

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

		s.client.
			EXPECT().
			UpdateTeam(renamedTeam).
			Return(updatedTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Successfully renamed team '"+existingName+"'")
	})
}

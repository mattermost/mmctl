package commands

import (
	"errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestCreateTeamCmd() {
	mockTeamName := "Mock Team"
	mockTeamDisplayname := "Mock Display Name"
	mockTeamEmail := "mock@mattermost.com"

	s.Run("Create team with no name returns error", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		err := createTeamCmdF(s.client, cmd, []string{})

		s.Require().Equal(err, errors.New("Name is required"))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Create team with a name but no display name returns error", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockTeamName, "")

		err := createTeamCmdF(s.client, cmd, []string{})
		s.Require().Equal(err, errors.New("Display Name is required"))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Create valid open team prints the created team", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockTeamName, "")
		cmd.Flags().String("display_name", mockTeamDisplayname, "")

		mockTeam := &model.Team{
			Name:        mockTeamName,
			DisplayName: mockTeamDisplayname,
			Type:        model.TEAM_OPEN,
		}

		s.client.
			EXPECT().
			CreateTeam(mockTeam).
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		err := createTeamCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Equal(mockTeam, printer.GetLines()[0])
		s.Require().Len(printer.GetLines(), 1)
	})

	s.Run("Create valid invite team with email prints the created team", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockTeamName, "")
		cmd.Flags().String("display_name", mockTeamDisplayname, "")
		cmd.Flags().String("email", mockTeamEmail, "")
		cmd.Flags().Bool("private", true, "")

		mockTeam := &model.Team{
			Name:        mockTeamName,
			DisplayName: mockTeamDisplayname,
			Email:       mockTeamEmail,
			Type:        model.TEAM_INVITE,
		}

		s.client.
			EXPECT().
			CreateTeam(mockTeam).
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		err := createTeamCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Equal(mockTeam, printer.GetLines()[0])
		s.Require().Len(printer.GetLines(), 1)
	})

	s.Run("Create returns an error when the client returns an error", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().String("name", mockTeamName, "")
		cmd.Flags().String("display_name", mockTeamDisplayname, "")

		mockTeam := &model.Team{
			Name:        mockTeamName,
			DisplayName: mockTeamDisplayname,
			Type:        model.TEAM_OPEN,
		}
		mockError := &model.AppError{Message: "Remote error"}

		s.client.
			EXPECT().
			CreateTeam(mockTeam).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := createTeamCmdF(s.client, cmd, []string{})
		s.Require().Equal("Team creation failed: : Remote error, ", err.Error())
		s.Require().Len(printer.GetLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestRenameTeamCmdF() {
	s.Run("Team rename should fail with missing name and display name flag", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := []string{""}
		args[0] = "existingName"

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Require atleast one flag to rename team, either 'name' or 'display_name'")
	})

	s.Run("Team rename should fail with invalid flags", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		args := []string{""}
		args[0] = "existingName"

		// Setting incorrect name flag
		cmd.Flags().String("-name", "newName", "Team Name Name")
		err := renameTeamCmdF(s.client, cmd, args)

		s.Require().EqualError(err, "Require atleast one flag to rename team, either 'name' or 'display_name'")

		// Setting flag as display-name instead of display_name
		cmd.Flags().String("display-name", "newDisplayName", "Team Display Name")
		s.Require().EqualError(err, "Require atleast one flag to rename team, either 'name' or 'display_name'")
	})

	s.Run("Team rename should fail when unknown existing team name is entered", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		args := []string{""}
		args[0] = "existingName"
		cmd.Flags().String("name", "newName", "Team Name")
		cmd.Flags().String("display_name", "newDisplayName", "Team Display Name")

		// Mocking : GetTeam searches with team id, if team not found proceeds with team name search
		s.client.
			EXPECT().
			GetTeam("existingName", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// Mocking : GetTeamByname is called, if GetTeam fails to return any team, as team name was passed instead of team id
		s.client.
			EXPECT().
			GetTeamByName("existingName", "").
			Return(nil, &model.Response{Error: nil}). // Error is nil as team not found will not return error from API
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)
		s.Require().EqualError(err, "Unable to find team 'existingName', to see the all teams try 'team list' command")
	})

	s.Run("Team rename should fail when api fails to rename", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newDisplayName := "NewDisplayName"
		newName := "newTeamName"
		args := []string{""}

		args[0] = existingName
		cmd.Flags().String("name", newName, "Display Name")
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
		s.Require().EqualError(err, "Cannot rename team '"+existingName+"', error : at-random-location.go: Mock Error, mocking a random error")
	})

	s.Run("Team rename should work as expected", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newDisplayName := "NewDisplayName"
		newName := "newTeamName"
		args := []string{""}

		args[0] = existingName
		cmd.Flags().String("name", newName, "Display Name")
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
		s.Require().Equal(printer.GetLines()[0], "'"+existingName+"' team renamed")
	})

	s.Run("Team rename should work as expected even if only display name is supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newDisplayName := "NewDisplayName"
		args := []string{""}

		args[0] = existingName
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
			UpdateTeam(updatedTeam).
			Return(updatedTeam, &model.Response{Error: nil}).
			Times(1)

		err := renameTeamCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "'"+existingName+"' team renamed")
	})

	s.Run("Team rename should work as expected even if only name is supplied", func() {
		printer.Clean()

		cmd := &cobra.Command{}

		existingName := "existingTeamName"
		existingDisplayName := "existingDisplayName"
		newName := "newTeamName"
		args := []string{""}

		args[0] = existingName
		cmd.Flags().String("name", newName, "Display Name")

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

		updatedTeam := &model.Team{
			Id:             "pm695ajd5pdotqs46144rcejnc",
			CreateAt:       1574191499747,
			UpdateAt:       1575551058238,
			DeleteAt:       0,
			DisplayName:    existingDisplayName,
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
		s.Require().Equal(printer.GetLines()[0], "'"+existingName+"' team renamed")
	})
}

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

func (s *MmctlUnitTestSuite) TestAddUsersCmd() {
	s.Run("Add users with not enough arguments returns error", func() {

		cmd := &cobra.Command{}
		err := addUsersCmdF(s.client, cmd, []string{})

		s.Require().Equal(err, errors.New("Not enough arguments."))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Add users with no team in arguments returns error", func() {

		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := addUsersCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Equal(err, errors.New("Unable to find team 'team1'"))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Add users with no existed user in arguments prints error", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		mockTeam := model.Team{
			Id:          "TeamId",
			Name:        "team1",
			DisplayName: "DisplayName",
		}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := addUsersCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Can't find user 'user1'")
	})

	s.Run("Add users should print error when cannot add team member", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		mockTeam := model.Team{
			Id:          "TeamId",
			Name:        "team1",
			DisplayName: "DisplayName",
		}

		mockUser := model.User{
			Id:       "UserID",
			Username: "ExampleUser",
			Email:    "example@example.com",
		}
		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		mockError := &model.AppError{Message: "Cannot add team member"}

		s.client.
			EXPECT().
			AddTeamMember("TeamId", "UserID").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := addUsersCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to add 'user1' to team1. Error: : Cannot add team member, ")
	})

	s.Run("Add users should not print in console anything on success", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		mockTeam := model.Team{
			Id:          "TeamId",
			Name:        "team1",
			DisplayName: "DisplayName",
		}

		mockUser := model.User{
			Id:       "UserID",
			Username: "ExampleUser",
			Email:    "example@example.com",
		}
		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			AddTeamMember("TeamId", "UserID").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := addUsersCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

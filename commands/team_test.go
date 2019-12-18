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

func (s *MmctlUnitTestSuite) TestRemoveUserCmd() {
	teamArg := "example-team-id"
	userArg := "example-user-id"
	s.Run("Remove users from team, without args returns an error", func() {
		printer.Clean()

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Equal(err, errors.New("Not enough arguments."))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Remove users from team, with one arg returns an error", func() {
		printer.Clean()

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{teamArg})
		s.Require().Equal(err, errors.New("Not enough arguments."))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Remove users from team, team not found, returns an error", func() {
		printer.Clean()

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

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{teamArg, userArg})
		s.Require().Equal(err, errors.New("Unable to find team '"+teamArg+"'"))
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Remove users from team, user not found error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{teamArg, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Can't find user '"+userArg+"'")
	})

	s.Run("Remove user by email get team by name, no error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by email, no error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by username, no error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by user no error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team unable to remove error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg, Name: "example-name"}
		mockUser := &model.User{Id: userArg}
		mockError := model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := removeUsersCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to remove '"+mockUser.Id+"' from "+mockTeam.Name+". Error: "+mockError.Error())
	})

}

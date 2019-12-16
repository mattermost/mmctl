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

func (s *MmctlUnitTestSuite) TestDeleteTeamsCmd() {
	teamName := "team1"
	teamId := "teamId"

	s.Run("Delete teams with no arguments returns an error", func() {
		cmd := &cobra.Command{}
		err := deleteTeamsCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "Not enough arguments.")
	})

	s.Run("Delete teams with confirm false returns an error", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", false, "")
		err := deleteTeamsCmdF(s.client, cmd, []string{"some"})
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "ABORTED: You did not answer YES exactly, in all capitals.")
	})

	s.Run("Delete teams with team not exist in db returns an error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetTeamByName(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteTeamsCmdF(s.client, cmd, []string{"team1"})
		s.Require().Nil(err)
		s.Require().Equal("Unable to find team 'team1'", printer.GetErrorLines()[0])
	})

	s.Run("Delete teams should delete team", func() {
		printer.Clean()
		mockTeam := model.Team{
			Id:   teamId,
			Name: teamName,
		}

		s.client.
			EXPECT().
			PermanentDeleteTeam(teamId).
			Return(true, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteTeamsCmdF(s.client, cmd, []string{"team1"})
		s.Require().Nil(err)
		s.Require().Equal(&mockTeam, printer.GetLines()[0])
	})

	s.Run("Delete teams with error on PermanentDeleteTeam returns an error", func() {
		printer.Clean()
		mockTeam := model.Team{
			Id:   teamId,
			Name: teamName,
		}

		mockError := &model.AppError{Message: "Permanent Delete Team Error"}

		s.client.
			EXPECT().
			PermanentDeleteTeam(teamId).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(teamName, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		err := deleteTeamsCmdF(s.client, cmd, []string{"team1"})
		s.Require().Nil(err)
		s.Require().Equal("Unable to delete team 'team1' error: : Permanent Delete Team Error, ", printer.GetErrorLines()[0])
	})
}

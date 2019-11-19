package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

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

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestArchiveTeamsCmd() {
	s.SetupTestHelper().InitBasic()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("confirm", true, "Confirm you really want to archive the team and a DB backup has been performed.")

	s.RunForAllClients("Archive nonexistent team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{"unknown-team"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to find team 'unknown-team'", printer.GetErrorLines()[0])
	})

	s.RunForSystemAdminAndLocal("Archive basic team", func(c client.Client) {
		printer.Clean()

		err := archiveTeamsCmdF(c, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		team := printer.GetLines()[0].(*model.Team)
		s.Require().Equal(s.th.BasicTeam.Name, team.Name)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Archive team without permissions", func() {
		printer.Clean()

		err := archiveTeamsCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions.")
	})
}

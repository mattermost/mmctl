package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestMoveCommandCmdF() {
	s.SetupTestHelper().InitBasic()

	// create new command
	newCmd := &model.Command{
		CreatorId: s.th.BasicUser.Id,
		TeamId:    s.th.BasicTeam.Id,
		URL:       "http://nowhere.com",
		Method:    model.COMMAND_METHOD_POST,
		Trigger:   "trigger",
	}
	command ,_ := s.th.SystemAdminClient.CreateCommand(newCmd)

	s.RunForAllClients("move command to non existing team", func(c client.Client) {
		printer.Clean()

		team := "nonexisting team"
		err := moveCommandCmdF(s.th.SystemAdminClient,
			&cobra.Command{},
			[]string{team, command.Id})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("unable to find team '%s'", team), err.Error())
	})

	s.RunForAllClients("move non existing command", func(c client.Client) {
		printer.Clean()

		err := moveCommandCmdF(s.th.SystemAdminClient,
			&cobra.Command{},
			[]string{s.th.BasicTeam.Name, "nothing"})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("unable to find command 'nothing'"), err.Error())
	})

	s.RunForAllClients("move existing command to existing team", func(c client.Client) {
		printer.Clean()

		err := moveCommandCmdF(s.th.SystemAdminClient,
			&cobra.Command{},
			[]string{s.th.BasicTeam.Name, command.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

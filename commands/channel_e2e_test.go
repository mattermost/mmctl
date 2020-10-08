package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestListChannelsCmdF() {
	s.SetupTestHelper().InitBasic()

	s.Run("List channels/Client", func() {
		printer.Clean()

		err := listChannelsCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Equal(6, len(printer.GetLines()))
		s.assertOutputDisplaysChannels()
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("List channels", func(c client.Client) {
		printer.Clean()

		err := listChannelsCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name})
		s.Require().Nil(err)
		s.Equal(7, len(printer.GetLines()))
		s.assertOutputDisplaysChannels()
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("List channels for non existent team", func(c client.Client) {
		printer.Clean()
		team := "non-existent-team"

		err := listChannelsCmdF(c, &cobra.Command{}, []string{team})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find team '"+team+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlE2ETestSuite) assertOutputDisplaysChannels() {
	validChannelNames := append(
		[]string{
			s.th.BasicChannel.Name,
			s.th.BasicChannel2.Name,
			s.th.BasicPrivateChannel.Name,
			s.th.BasicPrivateChannel2.Name,
			s.th.BasicDeletedChannel.Name,
		},
		s.th.App.DefaultChannelNames()...,
	)

	for i := range printer.GetLines() {
		s.Contains(validChannelNames, (printer.GetLines()[i].(*model.Channel)).Name)
	}
}

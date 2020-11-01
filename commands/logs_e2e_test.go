// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/client"

	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestlogsCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Display single log line", func(c client.Client) {
		cmd := &cobra.Command{}
		cmd.Flags().Int("number", 1, "")

		data, err := testLogsCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(data, 2)
		s.Contains(data[1], "info app/plugin.go:223 Syncing plugins from the file store")
	})

	s.RunForAllClients("Display logs", func(c client.Client) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("logrus", true, "")
		cmd.Flags().Int("number", 1, "")

		data, err := testLogsCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(data, 2)
		s.Contains(data[1], "level=info msg=\"Syncing plugins from the file store\" caller=\"app/plugin.go:223\"")
	})

	s.RunForAllClients("Error when using format flag", func(c client.Client) {
		cmd := &cobra.Command{}
		cmd.Flags().String("format", "json", "")
		cmd.Flags().Lookup("format").Changed = true

		data, err := testLogsCmdF(s.th.SystemAdminClient, cmd, []string{})

		s.Require().Error(err)
		s.Require().Equal(err.Error(), "the \"--format\" flag cannot be used with this command")
		s.Require().Len(data, 0)
	})
}

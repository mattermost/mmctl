// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestPluginMarketplaceListCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("List all plugins", func(c client.Client) {
		printer.Clean()

		id := "myplugin"
		version := "2.0.0"
		pluginRequest := &model.InstallMarketplacePluginRequest{Id: id, Version: version}
		_, _ = c.InstallMarketplacePlugin(pluginRequest)

		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", 1, "")
		cmd.Flags().Bool("all", true, "")

		err := pluginMarketplaceListCmdF(c, cmd, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 9)
		s.Require().Len(printer.GetLines(), 2)
	})
}

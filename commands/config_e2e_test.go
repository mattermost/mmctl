// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
package commands

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestConfigSetCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("Set config value for a given key", func(c client.Client) {
		printer.Clean()

		args := []string{"SqlSettings.DriverName", "mysql"}
		err := configSetCmdF(c, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		config, ok := printer.GetLines()[0].(*model.Config)
		s.Require().True(ok)
		s.Require().Equal("mysql", *(config.SqlSettings.DriverName))
	})

	s.RunForSystemAdminAndLocal("Get error if the key doesn't exists", func(c client.Client) {
		printer.Clean()

		args := []string{"SqlSettings.WrongKey", "mysql"}
		err := configSetCmdF(c, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set config value for a given key without permissions", func() {
		printer.Clean()

		args := []string{"SqlSettings.DriverName", "mysql"}
		err := configSetCmdF(s.th.Client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

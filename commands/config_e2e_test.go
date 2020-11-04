// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestConfigResetCmdE2E() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("System admin and local reset", func(c client.Client) {
		printer.Clean()
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.PrivacySettings.ShowEmailAddress = false })
		resetCmd := &cobra.Command{}
		resetCmd.Flags().Bool("confirm", true, "")
		err := configResetCmdF(c, resetCmd, []string{"PrivacySettings"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		config := s.th.App.Config()
		s.Require().True(*config.PrivacySettings.ShowEmailAddress)
	})

	s.Run("Reset for user without permission", func() {
		printer.Clean()
		resetCmd := &cobra.Command{}
		args := []string{"PrivacySettings"}
		resetCmd.Flags().Bool("confirm", true, "")
		err := configResetCmdF(s.th.Client, resetCmd, args)
		s.Require().NotNil(err)
		s.Assert().Errorf(err, "You do not have the appropriate permissions.")
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

	})
}

func (s *MmctlE2ETestSuite) TestConfigGetCmdF() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("Get config value for a given key", func(c client.Client) {
		printer.Clean()

		args := []string{"SqlSettings.DriverName"}
		err := configGetCmdF(c, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("postgres", *(printer.GetLines()[0].(*string)))
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Expect error when using a nonexistent key", func(c client.Client) {
		printer.Clean()

		args := []string{"NonExistent.Key"}
		err := configGetCmdF(c, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get config value for a given key without permissions", func() {
		printer.Clean()

		args := []string{"SqlSettings.DriverName"}
		err := configGetCmdF(s.th.Client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

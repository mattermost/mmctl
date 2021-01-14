// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"io/ioutil"
	"os"
	"os/exec"

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

func (s *MmctlE2ETestSuite) TestConfigEditCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("Edit a key in config", func(c client.Client) {
		printer.Clean()

		// check the value before editing
		args := []string{"ServiceSettings.EnableSVGs"}
		err := configGetCmdF(c, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().False(*printer.GetLines()[0].(*bool))
		s.Require().Len(printer.GetErrorLines(), 0)
		printer.Clean()

		// create a shell script to edit config
		content := `#! /bin/bash
sed -i'old' 's/\"EnableSVGs\": false/\"EnableSVGs\": true/' $1
rm $1'old'`

		file, err := ioutil.TempFile(os.TempDir(), "config_edit_*.sh")
		s.Require().Nil(err)
		defer func() {
			file.Close()
			os.Remove(file.Name())
			resetCmd := &cobra.Command{}
			resetCmd.Flags().Bool("confirm", true, "")
			s.Require().Nil(configResetCmdF(c, resetCmd, []string{"ServiceSettings"}))
		}()
		_, err = file.Write([]byte(content))
		s.Require().Nil(err)
		editorCmd := exec.Command("chmod", "+x", file.Name())
		s.Require().Nil(editorCmd.Run())

		os.Setenv("EDITOR", file.Name())

		// check the value after editing
		err = configEditCmdF(c, nil, nil)
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		config, ok := printer.GetLines()[0].(*model.Config)
		s.Require().True(ok)
		s.Require().True(*config.ServiceSettings.EnableSVGs)
	})

	s.Run("Edit config value without permissions", func() {
		printer.Clean()

		err := configEditCmdF(s.th.Client, nil, nil)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

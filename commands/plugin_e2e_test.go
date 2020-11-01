// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"

	"github.com/mattermost/mattermost-server/v5/model"

	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestPluginAddCmd() {
	s.SetupTestHelper().InitBasic()

	mmPath := os.Getenv("MM_SERVER_PATH")
	pluginPath := filepath.Join(mmPath, "../mmctl/commands/test_files/testplugin.tar.gz")
	fmt.Println(pluginPath)

	s.RunForSystemAdminAndLocal("admin and local can't add plugins if the config doesn't allow it", func(c client.Client) {
		printer.Clean()
		err := pluginAddCmdF(c, &cobra.Command{}, []string{pluginPath})
		s.Require().Nil(err)
		s.Require().Equal(1, len(printer.GetErrorLines()))
		s.Require().Contains(printer.GetErrorLines()[0], "Plugins and/or plugin uploads have been disabled.,")
	})

	s.RunForSystemAdminAndLocal("admin and local can add a plugin if the config allows it", func(c client.Client) {
		printer.Clean()
		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
			*cfg.PluginSettings.EnableUploads = true
		})

		err := pluginAddCmdF(c, &cobra.Command{}, []string{pluginPath})
		s.Require().Nil(err)

		s.Require().Equal(1, len(printer.GetLines()))
		s.Require().Contains(printer.GetLines()[0], "Added plugin: ")

		res, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Equal(1, len(res.Inactive))

		// teardown
		pInfo := res.Inactive[0]
		appErr = s.th.App.RemovePlugin(pInfo.Id)
		s.Require().Nil(appErr)

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = false
			*cfg.PluginSettings.EnableUploads = false
		})
	})

	s.Run("normal user can't add plugin", func() {
		printer.Clean()
		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = true
			*cfg.PluginSettings.EnableUploads = true
		})

		err := pluginAddCmdF(s.th.Client, &cobra.Command{}, []string{pluginPath})
		s.Require().Nil(err)
		s.Require().Equal(1, len(printer.GetErrorLines()))
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions")

		// teardown
		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = false
			*cfg.PluginSettings.EnableUploads = false
		})
	})
}

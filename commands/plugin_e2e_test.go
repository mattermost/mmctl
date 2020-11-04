// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestPluginInstallURLCmd() {
	s.SetupTestHelper().InitBasic()
	s.th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PluginSettings.Enable = true
		*cfg.PluginSettings.EnableUploads = true
	})

	const (
		jiraURL        = "https://plugins-store.test.mattermost.com/release/mattermost-plugin-jira-v3.0.0.tar.gz"
		jiraPluginID   = "jira"
		githubURL      = "https://plugins-store.test.mattermost.com/release/mattermost-plugin-github-v2.0.0.tar.gz"
		githubPluginID = "github"
	)

	s.RunForSystemAdminAndLocal("install new plugins", func(c client.Client) {
		printer.Clean()
		defer removePluginIfInstalled(s, jiraPluginID)
		defer removePluginIfInstalled(s, githubPluginID)

		err := pluginInstallURLCmdF(c, &cobra.Command{}, []string{jiraURL, githubURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(jiraPluginID, printer.GetLines()[0].(*model.Manifest).Id)
		s.Require().Equal(githubPluginID, printer.GetLines()[1].(*model.Manifest).Id)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 2)
	})

	s.Run("install a plugin without permissions", func() {
		printer.Clean()
		defer removePluginIfInstalled(s, jiraPluginID)

		err := pluginInstallURLCmdF(s.th.Client, &cobra.Command{}, []string{jiraURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to install plugin from URL \"%s\".", jiraURL))
		s.Require().Contains(printer.GetErrorLines()[0], "You do not have the appropriate permissions.")

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 0)
	})

	s.RunForSystemAdminAndLocal("install a nonexistent plugin", func(c client.Client) {
		printer.Clean()

		const pluginURL = "https://plugins-store.test.mattermost.com/release/mattermost-nonexistent-plugin-v2.0.0.tar.gz"

		err := pluginInstallURLCmdF(c, &cobra.Command{}, []string{pluginURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to install plugin from URL \"%s\".", pluginURL))
		s.Require().Contains(printer.GetErrorLines()[0], "An error occurred while downloading the plugin.")

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 0)
	})

	s.RunForSystemAdminAndLocal("install an already installed plugin without force", func(c client.Client) {
		printer.Clean()
		defer removePluginIfInstalled(s, jiraPluginID)

		err := pluginInstallURLCmdF(c, &cobra.Command{}, []string{jiraURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(jiraPluginID, printer.GetLines()[0].(*model.Manifest).Id)

		err = pluginInstallURLCmdF(c, &cobra.Command{}, []string{jiraURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Contains(printer.GetErrorLines()[0], fmt.Sprintf("Unable to install plugin from URL \"%s\".", jiraURL))
		s.Require().Contains(printer.GetErrorLines()[0], "Unable to install plugin. A plugin with the same ID is already installed.")

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
	})

	s.RunForSystemAdminAndLocal("install an already installed plugin with force", func(c client.Client) {
		printer.Clean()
		defer removePluginIfInstalled(s, jiraPluginID)

		err := pluginInstallURLCmdF(c, &cobra.Command{}, []string{jiraURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(jiraPluginID, printer.GetLines()[0].(*model.Manifest).Id)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", true, "")
		err = pluginInstallURLCmdF(c, cmd, []string{jiraURL})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(jiraPluginID, printer.GetLines()[1].(*model.Manifest).Id)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
	})
}

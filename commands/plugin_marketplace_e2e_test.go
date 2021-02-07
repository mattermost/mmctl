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

func (s *MmctlE2ETestSuite) TestPluginMarketplaceInstallCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("install a plugin", func(c client.Client) {
		printer.Clean()

		const (
			pluginID      = "jira"
			pluginVersion = "3.0.0"
		)

		defer removePluginIfInstalled(s, pluginID)

		err := pluginMarketplaceInstallCmdF(c, &cobra.Command{}, []string{pluginID, pluginVersion})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)

		manifest := printer.GetLines()[0].(*model.Manifest)
		s.Require().Equal(pluginID, manifest.Id)
		s.Require().Equal(pluginVersion, manifest.Version)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
		s.Require().Equal(pluginID, plugins.Inactive[0].Id)
		s.Require().Equal(pluginVersion, plugins.Inactive[0].Version)
	})

	s.Run("install a plugin without permissions", func() {
		printer.Clean()

		const (
			pluginID      = "jira"
			pluginVersion = "3.0.0"
		)

		defer removePluginIfInstalled(s, pluginID)

		err := pluginMarketplaceInstallCmdF(s.th.Client, &cobra.Command{}, []string{pluginID, pluginVersion})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "You do not have the appropriate permissions.")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 0)
	})

	s.RunForSystemAdminAndLocal("install a plugin without version", func(c client.Client) {
		printer.Clean()

		const (
			pluginID = "jira"
		)

		defer removePluginIfInstalled(s, pluginID)

		err := pluginMarketplaceInstallCmdF(c, &cobra.Command{}, []string{pluginID})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)

		manifest := printer.GetLines()[0].(*model.Manifest)
		s.Require().Equal(pluginID, manifest.Id)
		s.Require().NotEmpty(manifest.Version)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
		s.Require().Equal(pluginID, plugins.Inactive[0].Id)
		s.Require().NotEmpty(plugins.Inactive[0].Version)
	})

	s.RunForSystemAdminAndLocal("install a plugin with invalid version", func(c client.Client) {
		printer.Clean()

		const (
			pluginID      = "jira"
			pluginVersion = "invalid-version"
		)

		defer removePluginIfInstalled(s, pluginID)

		err := pluginMarketplaceInstallCmdF(c, &cobra.Command{}, []string{pluginID, pluginVersion})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "Could not find the requested marketplace plugin.")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 0)
	})

	s.RunForSystemAdminAndLocal("install a nonexistent plugin", func(c client.Client) {
		printer.Clean()

		const (
			pluginID = "a-nonexistent-plugin"
		)

		defer removePluginIfInstalled(s, pluginID)

		err := pluginMarketplaceInstallCmdF(c, &cobra.Command{}, []string{pluginID})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), fmt.Sprintf(`couldn't find a plugin with id "%s"`, pluginID))
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 0)
	})
}

func removePluginIfInstalled(s *MmctlE2ETestSuite, pluginID string) {
	appErr := s.th.App.RemovePlugin(pluginID)
	if appErr != nil {
		s.Require().Contains(appErr.Error(), "Plugin is not installed.")
	}
}

func (s *MmctlE2ETestSuite) TestPluginMarketplaceListCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("List Marketplace Plugins for Admin User", func(c client.Client) {
		printer.Clean()

		err := pluginMarketplaceListCmdF(c, &cobra.Command{}, nil)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("List Marketplace Plugins for non-admin User", func() {
		printer.Clean()

		err := pluginMarketplaceListCmdF(s.th.Client, &cobra.Command{}, nil)

		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "You do not have the appropriate permissions.")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})
}

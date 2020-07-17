// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestPluginMarketplaceInstallCmd() {
	s.Run("Install a valid plugin", func() {
		printer.Clean()

		id := "myplugin"
		version := "2.0.0"
		args := []string{id, version}
		pluginRequest := &model.InstallMarketplacePluginRequest{Id: id, Version: version}
		manifest := &model.Manifest{Name: "My Plugin", Id: id}

		s.client.
			EXPECT().
			InstallMarketplacePlugin(pluginRequest).
			Return(manifest, &model.Response{}).
			Times(1)

		err := pluginMarketplaceInstallCmdF(s.client, &cobra.Command{}, args)
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(manifest, printer.GetLines()[0])
	})

	s.Run("Install an invalid plugin", func() {
		printer.Clean()

		id := "myplugin"
		version := "2.0.0"
		args := []string{id, version}
		pluginRequest := &model.InstallMarketplacePluginRequest{Id: id, Version: version}

		s.client.
			EXPECT().
			InstallMarketplacePlugin(pluginRequest).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Mock error"}}).
			Times(1)

		err := pluginMarketplaceInstallCmdF(s.client, &cobra.Command{}, args)
		s.Require().Error(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})
}

func createMarketplacePlugin(name string) *model.MarketplacePlugin {
	return &model.MarketplacePlugin{
		BaseMarketplacePlugin: &model.BaseMarketplacePlugin{
			Manifest: &model.Manifest{Name: name},
		},
	}
}

func (s *MmctlUnitTestSuite) TestPluginMarketplaceListCmd() {
	s.Run("List honoring pagination flags", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().Int("page", 0, "")
		cmd.Flags().Int("per-page", 1, "")
		pluginFilter := &model.MarketplacePluginFilter{Page: 0, PerPage: 1}
		mockPlugin := createMarketplacePlugin("My Plugin")
		plugins := []*model.MarketplacePlugin{mockPlugin}

		s.client.
			EXPECT().
			GetMarketplacePlugins(pluginFilter).
			Return(plugins, &model.Response{}).
			Times(1)

		err := pluginMarketplaceListCmdF(s.client, cmd, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(mockPlugin, printer.GetLines()[0])
	})

	s.Run("List all plugins", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", 1, "")
		cmd.Flags().Bool("all", true, "")
		mockPlugin1 := createMarketplacePlugin("My Plugin One")
		mockPlugin2 := createMarketplacePlugin("My Plugin Two")

		s.client.
			EXPECT().
			GetMarketplacePlugins(&model.MarketplacePluginFilter{Page: 0, PerPage: 1}).
			Return([]*model.MarketplacePlugin{mockPlugin1}, &model.Response{}).
			Times(1)

		s.client.
			EXPECT().
			GetMarketplacePlugins(&model.MarketplacePluginFilter{Page: 1, PerPage: 1}).
			Return([]*model.MarketplacePlugin{mockPlugin2}, &model.Response{}).
			Times(1)

		s.client.
			EXPECT().
			GetMarketplacePlugins(&model.MarketplacePluginFilter{Page: 2, PerPage: 1}).
			Return([]*model.MarketplacePlugin{}, &model.Response{}).
			Times(1)

		err := pluginMarketplaceListCmdF(s.client, cmd, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(mockPlugin1, printer.GetLines()[0])
		s.Require().Equal(mockPlugin2, printer.GetLines()[1])
	})

	s.Run("List all plugins with errors", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", 200, "")

		s.client.
			EXPECT().
			GetMarketplacePlugins(&model.MarketplacePluginFilter{Page: 0, PerPage: 200}).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Mock error"}}).
			Times(1)

		err := pluginMarketplaceListCmdF(s.client, cmd, []string{})
		s.Require().Error(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("List honoring filter and local only flags", func() {
		printer.Clean()

		filter := "jit"
		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", 200, "")
		cmd.Flags().String("filter", filter, "")
		cmd.Flags().Bool("local-only", true, "")
		pluginFilter := &model.MarketplacePluginFilter{Page: 0, PerPage: 200, Filter: filter, LocalOnly: true}
		mockPlugin := createMarketplacePlugin("Jitsi")
		plugins := []*model.MarketplacePlugin{mockPlugin}

		s.client.
			EXPECT().
			GetMarketplacePlugins(pluginFilter).
			Return(plugins, &model.Response{}).
			Times(1)

		err := pluginMarketplaceListCmdF(s.client, cmd, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(mockPlugin, printer.GetLines()[0])
	})
}

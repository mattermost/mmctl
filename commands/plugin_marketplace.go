// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var PluginMarketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Management of marketplace plugins",
}

var PluginMarketplaceInstallCmd = &cobra.Command{
	Use:   "install <id> [version]",
	Short: "Install a plugin from the marketplace",
	Long:  "Installs a plugin listed in the marketplace server",
	Example: `  # you can specify with both the plugin id and its version
  $ mmctl plugin marketplace install jitsi 2.0.0

  # if you don't specify the version, the latest one will be installed
  $ mmctl plugin marketplace install jitsi`,
	Args: cobra.MinimumNArgs(1),
	RunE: withClient(pluginMarketplaceInstallCmdF),
}

var PluginMarketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List marketplace plugins",
	Long:  "Gets all plugins from the marketplace server, merging data from locally installed plugins as well as prepackaged plugins shipped with the server",
	Example: `  # You can list all the plugins
  $ mmctl plugin marketplace list --all

  # Pagination options can be used too
  $ mmctl plugin marketplace list --page 2 --per-page 10

  # Filtering will narrow down the search
  $ mmctl plugin marketplace list --filter jit

  # You can only retrieve local plugins
  $ mmctl plugin marketplace list --local-only`,
	Args: cobra.NoArgs,
	RunE: withClient(pluginMarketplaceListCmdF),
}

func init() {
	PluginMarketplaceListCmd.Flags().Int("page", 0, "Page number to fetch for the list of users")
	PluginMarketplaceListCmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	PluginMarketplaceListCmd.Flags().Bool("all", false, "Fetch all plugins. --page flag will be ignore if provided")
	PluginMarketplaceListCmd.Flags().String("filter", "", "Filter plugins by ID, name or description")
	PluginMarketplaceListCmd.Flags().Bool("local-only", false, "Only retrieve local plugins")

	PluginMarketplaceCmd.AddCommand(
		PluginMarketplaceInstallCmd,
		PluginMarketplaceListCmd,
	)

	PluginCmd.AddCommand(
		PluginMarketplaceCmd,
	)
}

func pluginMarketplaceInstallCmdF(c client.Client, _ *cobra.Command, args []string) error {
	id := args[0]
	var version string

	if len(args) < 2 {
		plugins, res := c.GetMarketplacePlugins(&model.MarketplacePluginFilter{Filter: id, PerPage: 200})
		if res.Error != nil {
			return errors.Wrap(res.Error, "Failed to fetch plugins")
		}

		for _, plugin := range plugins {
			if plugin.Manifest.Id == id {
				version = plugin.Manifest.Version
			}
		}

		if version == "" {
			return fmt.Errorf("couldn't find a plugin with id %q", id)
		}
	} else {
		version = args[1]
	}

	pluginRequest := &model.InstallMarketplacePluginRequest{Id: id, Version: version}
	manifest, resp := c.InstallMarketplacePlugin(pluginRequest)
	if resp.Error != nil {
		return errors.Wrap(resp.Error, "couldn't install plugin from marketplace")
	}

	printer.PrintT("Plugin {{.Name}} successfully installed", manifest)

	return nil
}

func pluginMarketplaceListCmdF(c client.Client, cmd *cobra.Command, _ []string) error {
	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	showAll, _ := cmd.Flags().GetBool("all")
	filter, _ := cmd.Flags().GetString("filter")
	localOnly, _ := cmd.Flags().GetBool("local-only")

	if showAll {
		page = 0
	}

	for {
		pluginFilter := &model.MarketplacePluginFilter{
			Page:      page,
			PerPage:   perPage,
			Filter:    filter,
			LocalOnly: localOnly,
		}

		plugins, res := c.GetMarketplacePlugins(pluginFilter)
		if res.Error != nil {
			return errors.Wrap(res.Error, "Failed to fetch plugins")
		}
		if len(plugins) == 0 {
			break
		}

		for _, plugin := range plugins {
			printer.PrintT("{{.Manifest.Id}}: {{.Manifest.Name}}, Version: {{.Manifest.Version}}", plugin)
		}

		if !showAll {
			break
		}
		page++
	}

	return nil
}

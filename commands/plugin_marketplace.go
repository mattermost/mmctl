// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func PluginMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Management of marketplace plugins",
	}

	cmd.AddCommand(
		InstallPluginMarketplaceCmd(),
		ListPluginMarketplaceCmd(),
	)

	return cmd
}

func InstallPluginMarketplaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "install <id> <version>",
		Short:   "Install a plugin from the marketplace",
		Long:    "",
		Example: "  plugin marketplace install jitsi 2.0.0",
		Args:    cobra.ExactArgs(2),
		RunE:    withClient(installPluginMarketplaceCmdF),
	}
}

func ListPluginMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List marketplace plugins",
		Long:  "",
		Example: `  plugin marketplace list
  plugin marketplace list --all
  plugin marketplace list --filter jit
  plugin marketplace list --local-only --page 2 --per-page 10`,
		Args: cobra.NoArgs,
		RunE: withClient(listPluginMarketplaceCmdF),
	}

	cmd.Flags().Int("page", 0, "Page number to fetch for the list of users")
	cmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	cmd.Flags().Bool("all", false, "Fetch all plugins. --page flag will be ignore if provided")
	cmd.Flags().String("filter", "", "Filter plugins by ID, name or description")
	cmd.Flags().Bool("local-only", false, "Only retrieve local plugins")

	return cmd
}

func init() {
	PluginCmd.AddCommand(
		PluginMarketplaceCmd(),
	)
}

func installPluginMarketplaceCmdF(c client.Client, _ *cobra.Command, args []string) error {
	pluginRequest := &model.InstallMarketplacePluginRequest{Id: args[0], Version: args[1]}
	manifest, resp := c.InstallMarketplacePlugin(pluginRequest)
	if resp.Error != nil {
		return errors.Wrap(resp.Error, "couldn't install plugin from marketplace")
	}

	printer.PrintT("Plugin {{.Name}} successfully installed", manifest)

	return nil
}

func listPluginMarketplaceCmdF(c client.Client, cmd *cobra.Command, _ []string) error {
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

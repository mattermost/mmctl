// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"os"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var PluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Management of plugins",
}

var PluginAddCmd = &cobra.Command{
	Use:     "add [plugins]",
	Short:   "Add plugins",
	Long:    "Add plugins to your Mattermost server.",
	Example: `  plugin add hovercardexample.tar.gz pluginexample.tar.gz`,
	RunE:    withClient(pluginAddCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var PluginDeleteCmd = &cobra.Command{
	Use:     "delete [plugins]",
	Short:   "Delete plugins",
	Long:    "Delete previously uploaded plugins from your Mattermost server.",
	Example: `  plugin delete hovercardexample pluginexample`,
	RunE:    withClient(pluginDeleteCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var PluginEnableCmd = &cobra.Command{
	Use:     "enable [plugins]",
	Short:   "Enable plugins",
	Long:    "Enable plugins for use on your Mattermost server.",
	Example: `  plugin enable hovercardexample pluginexample`,
	RunE:    withClient(pluginEnableCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var PluginDisableCmd = &cobra.Command{
	Use:     "disable [plugins]",
	Short:   "Disable plugins",
	Long:    "Disable plugins. Disabled plugins are immediately removed from the user interface and logged out of all sessions.",
	Example: `  plugin disable hovercardexample pluginexample`,
	RunE:    withClient(pluginDisableCmdF),
	Args:    cobra.MinimumNArgs(1),
}

var PluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List plugins",
	Long:    "List all enabled and disabled plugins installed on your Mattermost server.",
	Example: `  plugin list`,
	RunE:    withClient(pluginListCmdF),
}

func init() {
	PluginCmd.AddCommand(
		PluginAddCmd,
		PluginDeleteCmd,
		PluginEnableCmd,
		PluginDisableCmd,
		PluginListCmd,
	)
	RootCmd.AddCommand(PluginCmd)
}

func pluginAddCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	for i, plugin := range args {
		fileReader, err := os.Open(plugin)
		if err != nil {
			return err
		}

		if _, response := c.UploadPlugin(fileReader); response.Error != nil {
			printer.PrintError("Unable to add plugin: " + args[i] + ". Error: " + response.Error.Error())
		} else {
			printer.Print("Added plugin: " + plugin)
		}
		fileReader.Close()
	}

	return nil
}

func pluginDeleteCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	for _, plugin := range args {
		if _, response := c.RemovePlugin(plugin); response.Error != nil {
			printer.PrintError("Unable to delete plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			printer.Print("Deleted plugin: " + plugin)
		}
	}

	return nil
}

func pluginEnableCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	for _, plugin := range args {
		if _, response := c.EnablePlugin(plugin); response.Error != nil {
			printer.PrintError("Unable to enable plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			printer.Print("Enabled plugin: " + plugin)
		}
	}

	return nil
}

func pluginDisableCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	for _, plugin := range args {
		if _, response := c.DisablePlugin(plugin); response.Error != nil {
			printer.PrintError("Unable to disable plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			printer.Print("Disabled plugin: " + plugin)
		}
	}

	return nil
}

func pluginListCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	pluginsResp, response := c.GetPlugins()
	if response.Error != nil {
		return errors.New("Unable to list plugins. Error: " + response.Error.Error())
	}

	format, _ := cmd.Flags().GetString("format")
	if format == printer.FORMAT_JSON {
		printer.Print(pluginsResp)
	} else {
		printer.Print("Listing enabled plugins")
		for _, plugin := range pluginsResp.Active {
			printer.PrintT("{{.Manifest.Id}}: {{.Manifest.Name}}, Version: {{.Manifest.Version}}", plugin)
		}

		printer.Print("Listing disabled plugins")
		for _, plugin := range pluginsResp.Inactive {
			printer.PrintT("{{.Manifest.Id}}: {{.Manifest.Name}}, Version: {{.Manifest.Version}}", plugin)
		}
	}

	return nil
}

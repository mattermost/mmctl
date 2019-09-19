package commands

import (
	"errors"
	"os"

	"github.com/mattermost/mattermost-server/model"

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
}

var PluginDeleteCmd = &cobra.Command{
	Use:     "delete [plugins]",
	Short:   "Delete plugins",
	Long:    "Delete previously uploaded plugins from your Mattermost server.",
	Example: `  plugin delete hovercardexample pluginexample`,
	RunE:    withClient(pluginDeleteCmdF),
}

var PluginEnableCmd = &cobra.Command{
	Use:     "enable [plugins]",
	Short:   "Enable plugins",
	Long:    "Enable plugins for use on your Mattermost server.",
	Example: `  plugin enable hovercardexample pluginexample`,
	RunE:    withClient(pluginEnableCmdF),
}

var PluginDisableCmd = &cobra.Command{
	Use:     "disable [plugins]",
	Short:   "Disable plugins",
	Long:    "Disable plugins. Disabled plugins are immediately removed from the user interface and logged out of all sessions.",
	Example: `  plugin disable hovercardexample pluginexample`,
	RunE:    withClient(pluginDisableCmdF),
}

var PluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List plugins",
	Long:    "List all active and inactive plugins installed on your Mattermost server.",
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

func pluginAddCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for i, plugin := range args {
		fileReader, err := os.Open(plugin)
		if err != nil {
			return err
		}

		if _, response := c.UploadPlugin(fileReader); response.Error != nil {
			CommandPrintErrorln("Unable to add plugin: " + args[i] + ". Error: " + response.Error.Error())
		} else {
			CommandPrettyPrintln("Added plugin: " + plugin)
		}
		fileReader.Close()
	}

	return nil
}

func pluginDeleteCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for _, plugin := range args {
		if _, response := c.RemovePlugin(plugin); response.Error != nil {
			CommandPrintErrorln("Unable to delete plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			CommandPrettyPrintln("Deleted plugin: " + plugin)
		}
	}

	return nil
}

func pluginEnableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for _, plugin := range args {
		if _, response := c.EnablePlugin(plugin); response.Error != nil {
			CommandPrintErrorln("Unable to enable plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			CommandPrettyPrintln("Enabled plugin: " + plugin)
		}
	}

	return nil
}

func pluginDisableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Expected at least one argument. See help text for details.")
	}

	for _, plugin := range args {
		if _, response := c.DisablePlugin(plugin); response.Error != nil {
			CommandPrintErrorln("Unable to disable plugin: " + plugin + ". Error: " + response.Error.Error())
		} else {
			CommandPrettyPrintln("Disabled plugin: " + plugin)
		}
	}

	return nil
}

func pluginListCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	pluginsResp, response := c.GetPlugins()
	if response.Error != nil {
		return errors.New("Unable to list plugins. Error: " + response.Error.Error())
	}

	CommandPrettyPrintln("Listing active plugins")
	for _, plugin := range pluginsResp.Active {
		CommandPrettyPrintln(plugin.Manifest.Id + ": " + plugin.Manifest.Name + ", Version: " + plugin.Manifest.Version)
	}

	CommandPrettyPrintln("Listing inactive plugins")
	for _, plugin := range pluginsResp.Inactive {
		CommandPrettyPrintln(plugin.Manifest.Id + ": " + plugin.Manifest.Name + ", Version: " + plugin.Manifest.Version)
	}

	return nil
}

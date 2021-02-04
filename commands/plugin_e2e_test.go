package commands

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

const (
	pluginID            = "com.mattermost.demo-plugin"
	pluginURL           = filepath.Join(os.Getenv("MM_SERVER_PATH"), "tests", "testplugin.tar.gz")
	nonExistentPluginID = "nonExistentPluginID"
)

func (s *MmctlE2ETestSuite) TestPluginEnableCmd() {
	s.SetupTestHelper().InitBasic()
	installPlugin(s, pluginID, pluginURL)
	defer removePlugin(s, pluginID)

	s.RunForSystemAdminAndLocal("Successful enable plugin", func(c client.Client) {
		printer.Clean()

		appErr := s.th.App.DisablePlugin(pluginID)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)

		cmd := &cobra.Command{}
		err := pluginEnableCmdF(c, cmd, []string{pluginID})
		s.Require().Nil(err)

		plugins, appErr = s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 1)
		s.Require().Len(plugins.Inactive, 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled plugin: "+pluginID)

	})

	s.Run("error for enable plugin", func() {
		printer.Clean()

		appErr := s.th.App.DisablePlugin(pluginID)
		s.Require().Nil(appErr)

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)

		cmd := &cobra.Command{}
		_ = pluginEnableCmdF(s.th.Client, cmd, []string{pluginID})
		//err = pluginEnableCmdF(c, cmd, []string{pluginID})
		//s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+pluginID+". Error: : You do not have the appropriate permissions., ")

		plugins, appErr = s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
	})

	s.RunForSystemAdminAndLocal("error for enabling non existent plugin", func(c client.Client) {
		printer.Clean()

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)

		cmd := &cobra.Command{}
		_ = pluginEnableCmdF(c, cmd, []string{nonExistentPluginID})
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+nonExistentPluginID+". Error: : Plugin is not installed., ")

		plugins, appErr = s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
	})

	s.Run("error for enabling non existent plugin", func() {
		printer.Clean()

		plugins, appErr := s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)

		cmd := &cobra.Command{}
		_ = pluginEnableCmdF(s.th.Client, cmd, []string{nonExistentPluginID})
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+nonExistentPluginID+". Error: : You do not have the appropriate permissions., ")

		plugins, appErr = s.th.App.GetPlugins()
		s.Require().Nil(appErr)
		s.Require().Len(plugins.Active, 0)
		s.Require().Len(plugins.Inactive, 1)
	})

	s.RunForAllClients("error when plugins are disabled", func(c client.Client) {
		printer.Clean()

		s.th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.PluginSettings.Enable = false
		})

		cmd := &cobra.Command{}
		_ = pluginEnableCmdF(c, cmd, []string{pluginID})
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+pluginID+". Error: : Plugins have been disabled. Please check your logs for details., ")
	})
}

func installPlugin(s *MmctlE2ETestSuite, pluginID string, pluginURL string) {
	s.th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PluginSettings.Enable = true
	})

	pluginFileBytes, err := s.th.App.DownloadFromURL(pluginURL)
	s.Require().Nil(err)
	_, err = s.th.App.InstallPlugin(bytes.NewReader(pluginFileBytes), true)
	s.Require().Nil(err)
}

func removePlugin(s *MmctlE2ETestSuite, pluginID string) {
	s.th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PluginSettings.Enable = true
	})
	appErr := s.th.App.RemovePlugin(pluginID)
	if appErr != nil {
		s.Require().Contains(appErr.Error(), "Plugin is not installed.")
	}
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlUnitTestSuite) TestConfigGetCmd() {
	s.Run("Get a string config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.DriverName"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(*(printer.GetLines()[0].(*string)), "mysql")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get an int config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.MaxIdleConns"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(*(printer.GetLines()[0].(*int)), 20)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get an int64 config value for a given key", func() {
		printer.Clean()
		args := []string{"FileSettings.MaxFileSize"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(*(printer.GetLines()[0].(*int64)), int64(52428800))
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get a boolean config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.Trace"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(*(printer.GetLines()[0].(*bool)), false)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get a slice of string config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.DataSourceReplicas"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], []string{})
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get config struct for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()
		sqlSettings := model.SqlSettings{}
		sqlSettings.SetDefaults(false)

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], sqlSettings)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get error if the key doesn't exists", func() {
		printer.Clean()
		args := []string{"SqlSettings.WrongKey"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()
		sqlSettings := model.SqlSettings{}
		sqlSettings.SetDefaults(false)

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should handle the response error", func() {
		printer.Clean()
		args := []string{"SqlSettings.DriverName"}
		outputConfig := &model.Config{}
		outputConfig.SetDefaults()
		sqlSettings := model.SqlSettings{}
		sqlSettings.SetDefaults(false)

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{StatusCode: 500, Error: &model.AppError{}}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get value if the key points to a map element", func() {
		outputConfig := &model.Config{}
		pluginState := &model.PluginState{Enable: true}
		pluginSettings := map[string]interface{}{
			"test1": 1,
			"test2": []string{"a", "b"},
			"test3": map[string]string{"a": "b"},
		}
		outputConfig.PluginSettings.PluginStates = map[string]*model.PluginState{
			"com.mattermost.testplugin": pluginState,
		}
		outputConfig.PluginSettings.Plugins = map[string]map[string]interface{}{
			"com.mattermost.testplugin": pluginSettings,
		}

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(7)

		printer.Clean()
		err := configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.PluginStates.com.mattermost.testplugin"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], pluginState)
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], pluginSettings)
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test2"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], []string{"a", "b"})
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test3"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], map[string]string{"a": "b"})
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configGetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test3.a"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "b")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get error value if the key points to a missing map element", func() {
		printer.Clean()
		args := []string{"PluginSettings.PluginStates.com.mattermost.testplugin.x"}
		outputConfig := &model.Config{}
		pluginState := &model.PluginState{Enable: true}
		outputConfig.PluginSettings.PluginStates = map[string]*model.PluginState{
			"com.mattermost.testplugin": pluginState,
		}

		s.client.
			EXPECT().
			GetConfig().
			Return(outputConfig, &model.Response{Error: nil}).
			Times(0)

		err := configGetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestConfigSetCmd() {
	s.Run("Set a string config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.DriverName", "postgres"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		changedValue := "postgres"
		inputConfig.SqlSettings.DriverName = &changedValue

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set an int config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.MaxIdleConns", "20"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		changedValue := 20
		inputConfig.SqlSettings.MaxIdleConns = &changedValue

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set an int64 config value for a given key", func() {
		printer.Clean()
		args := []string{"FileSettings.MaxFileSize", "52428800"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		changedValue := int64(52428800)
		inputConfig.FileSettings.MaxFileSize = &changedValue

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set a boolean config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.Trace", "true"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		changedValue := true
		inputConfig.SqlSettings.Trace = &changedValue

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set a slice of string config value for a given key", func() {
		printer.Clean()
		args := []string{"SqlSettings.DataSourceReplicas", "test1", "test2"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		inputConfig.SqlSettings.DataSourceReplicas = []string{"test1", "test2"}

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should get an error if a string is passed while trying to set a slice", func() {
		printer.Clean()
		args := []string{"SqlSettings.DataSourceReplicas", "[\"test1\", \"test2\"]"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		inputConfig.SqlSettings.DataSourceReplicas = []string{"test1", "test2"}

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Get error if the key doesn't exists", func() {
		printer.Clean()
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		args := []string{"SqlSettings.WrongKey", "test1"}
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should handle response error from the server", func() {
		printer.Clean()
		args := []string{"SqlSettings.DriverName", "postgres"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		changedValue := "postgres"
		inputConfig.SqlSettings.DriverName = &changedValue

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{StatusCode: 500, Error: &model.AppError{}}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Set a field inside a map", func() {
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		defaultConfig.PluginSettings.PluginStates = map[string]*model.PluginState{
			"com.mattermost.testplugin": {Enable: false},
		}
		pluginSettings := map[string]interface{}{
			"test1": 1,
			"test2": []string{"a", "b"},
			"test3": map[string]interface{}{"a": "b"},
		}
		defaultConfig.PluginSettings.Plugins = map[string]map[string]interface{}{
			"com.mattermost.testplugin": pluginSettings,
		}

		inputConfig := &model.Config{}
		inputConfig.SetDefaults()
		inputConfig.PluginSettings.PluginStates = map[string]*model.PluginState{
			"com.mattermost.testplugin": {Enable: true},
		}
		inputConfig.PluginSettings.Plugins = map[string]map[string]interface{}{
			"com.mattermost.testplugin": pluginSettings,
		}
		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(3)

		s.client.
			EXPECT().
			PatchConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(3)

		printer.Clean()
		err := configSetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.PluginStates.com.mattermost.testplugin.Enable", "true"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configSetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test1", "123"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)

		printer.Clean()
		err = configSetCmdF(s.client, &cobra.Command{}, []string{"PluginSettings.Plugins.com.mattermost.testplugin.test3.a", "123"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to set a field inside a map for incorrect field, get error", func() {
		printer.Clean()
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()
		defaultConfig.PluginSettings.PluginStates = map[string]*model.PluginState{
			"com.mattermost.testplugin": {Enable: true},
		}
		args := []string{"PluginSettings.PluginStates.com.mattermost.testplugin.x", "true"}

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestConfigResetCmd() {
	s.Run("Reset a single key", func() {
		printer.Clean()
		args := []string{"SqlSettings.DriverName"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateConfig(defaultConfig).
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		resetCmd := &cobra.Command{}
		resetCmd.Flags().Bool("confirm", true, "")
		err := configResetCmdF(s.client, resetCmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], defaultConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Reset a whole config section", func() {
		printer.Clean()
		args := []string{"SqlSettings"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			UpdateConfig(defaultConfig).
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		resetCmd := &cobra.Command{}
		resetCmd.Flags().Bool("confirm", true, "")
		_ = resetCmd.ParseFlags([]string{"confirm"})
		err := configResetCmdF(s.client, resetCmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], defaultConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should fail if the key doesn't exists", func() {
		printer.Clean()
		args := []string{"WrongKey"}
		defaultConfig := &model.Config{}
		defaultConfig.SetDefaults()

		s.client.
			EXPECT().
			GetConfig().
			Return(defaultConfig, &model.Response{Error: nil}).
			Times(1)

		resetCmd := &cobra.Command{}
		resetCmd.Flags().Bool("confirm", true, "")
		_ = resetCmd.ParseFlags([]string{"confirm"})
		err := configResetCmdF(s.client, resetCmd, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestConfigShowCmd() {
	s.Run("Should show config", func() {
		printer.Clean()
		mockConfig := &model.Config{}

		s.client.
			EXPECT().
			GetConfig().
			Return(mockConfig, &model.Response{Error: nil}).
			Times(1)

		err := configShowCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(mockConfig, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should return an error", func() {
		printer.Clean()
		configError := &model.AppError{Message: "Config Error"}

		s.client.
			EXPECT().
			GetConfig().
			Return(nil, &model.Response{Error: configError}).
			Times(1)

		err := configShowCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.EqualError(err, configError.Error())
	})
}

func (s *MmctlUnitTestSuite) TestConfigReloadCmd() {
	s.Run("Should reload config", func() {
		printer.Clean()

		s.client.
			EXPECT().
			ReloadConfig().
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := configReloadCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Should fail on error when reload config", func() {
		printer.Clean()

		s.client.
			EXPECT().
			ReloadConfig().
			Return(false, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := configReloadCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
	})
}

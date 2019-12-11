package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
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
		s.Require().Equal(printer.GetLines()[0].([]string), []string{})
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
		s.Require().Equal(printer.GetLines()[0].(model.SqlSettings), sqlSettings)
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
			UpdateConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0].(*model.Config), inputConfig)
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
			UpdateConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0].(*model.Config), inputConfig)
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
			UpdateConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0].(*model.Config), inputConfig)
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
			UpdateConfig(inputConfig).
			Return(inputConfig, &model.Response{Error: nil}).
			Times(1)

		err := configSetCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0].(*model.Config), inputConfig)
		s.Require().Len(printer.GetErrorLines(), 0)
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
			UpdateConfig(inputConfig).
			Return(inputConfig, &model.Response{StatusCode: 500, Error: &model.AppError{}}).
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
		s.Require().Equal(printer.GetLines()[0].(*model.Config), defaultConfig)
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
		resetCmd.ParseFlags([]string{"confirm"})
		err := configResetCmdF(s.client, resetCmd, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0].(*model.Config), defaultConfig)
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
		resetCmd.ParseFlags([]string{"confirm"})
		err := configResetCmdF(s.client, resetCmd, args)
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

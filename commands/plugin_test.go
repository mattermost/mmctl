package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestPluginDisableCmd() {
	s.Run("Disable 1 plugin", func() {
		printer.Clean()
		arg := "plug1"

		s.client.
			EXPECT().
			DisablePlugin(arg).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := pluginDisableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled plugin: "+arg)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Fail to disable 1 plugin", func() {
		printer.Clean()
		arg := "fail1"
		mockError := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			DisablePlugin(arg).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := pluginDisableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to disable plugin: "+arg+". Error: "+mockError.Error())
	})

	s.Run("Disble several plugin with some errors", func() {
		printer.Clean()
		args := []string{"fail1", "plug2", "plug3", "fail4"}
		mockError := &model.AppError{Message: "Mock Error"}

		for _, arg := range args {
			if strings.HasPrefix(arg, "fail") {
				s.client.
					EXPECT().
					DisablePlugin(arg).
					Return(false, &model.Response{Error: mockError}).
					Times(1)
			} else {
				s.client.
					EXPECT().
					DisablePlugin(arg).
					Return(false, &model.Response{Error: nil}).
					Times(1)
			}
		}

		err := pluginDisableCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], "Disabled plugin: "+args[1])
		s.Require().Equal(printer.GetLines()[1], "Disabled plugin: "+args[2])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to disable plugin: "+args[0]+". Error: "+mockError.Error())
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to disable plugin: "+args[3]+". Error: "+mockError.Error())
	})
}

func (s *MmctlUnitTestSuite) TestPluginEnableCmd() {
	s.Run("Enable 1 plugin", func() {
		printer.Clean()
		pluginArg := "test-plugin"

		s.client.
			EXPECT().
			EnablePlugin(pluginArg).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := pluginEnableCmdF(s.client, &cobra.Command{}, []string{pluginArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled plugin: "+pluginArg)
	})

	s.Run("Enable multiple plugins", func() {
		printer.Clean()
		plugins := []string{"plugin1", "plugin2", "plugin3"}

		for _, plugin := range plugins {
			s.client.
				EXPECT().
				EnablePlugin(plugin).
				Return(false, &model.Response{Error: nil}).
				Times(1)
		}

		err := pluginEnableCmdF(s.client, &cobra.Command{}, plugins)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 3)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(printer.GetLines()[0], "Enabled plugin: "+plugins[0])
		s.Require().Equal(printer.GetLines()[1], "Enabled plugin: "+plugins[1])
		s.Require().Equal(printer.GetLines()[2], "Enabled plugin: "+plugins[2])
	})

	s.Run("Fail to enable plugin", func() {
		printer.Clean()
		pluginArg := "fail-plugin"
		mockErr := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			EnablePlugin(pluginArg).
			Return(false, &model.Response{Error: mockErr}).
			Times(1)

		err := pluginEnableCmdF(s.client, &cobra.Command{}, []string{pluginArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+pluginArg+". Error: "+mockErr.Error())
	})

	s.Run("Enable multiple plugins with some having errors", func() {
		printer.Clean()
		okPlugins := []string{"ok-plugin-1", "ok-plugin-2"}
		failPlugins := []string{"fail-plugin-1", "fail-plugin-2"}
		allPlugins := append(okPlugins, failPlugins...)

		mockErr := &model.AppError{Message: "Mock Error"}

		for _, plugin := range okPlugins {
			s.client.
				EXPECT().
				EnablePlugin(plugin).
				Return(false, &model.Response{Error: nil}).
				Times(1)
		}

		for _, plugin := range failPlugins {
			s.client.
				EXPECT().
				EnablePlugin(plugin).
				Return(false, &model.Response{Error: mockErr}).
				Times(1)
		}

		err := pluginEnableCmdF(s.client, &cobra.Command{}, allPlugins)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], "Enabled plugin: "+okPlugins[0])
		s.Require().Equal(printer.GetLines()[1], "Enabled plugin: "+okPlugins[1])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to enable plugin: "+failPlugins[0]+". Error: "+mockErr.Error())
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to enable plugin: "+failPlugins[1]+". Error: "+mockErr.Error())
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestPluginAddCmd() {
	s.Run("Add 1 plugin", func() {
		printer.Clean()
		tmpFile, err := ioutil.TempFile("", "tmpPlugin")
		s.Require().Nil(err)
		defer os.Remove(tmpFile.Name())

		pluginName := tmpFile.Name()

		s.client.
			EXPECT().
			UploadPlugin(gomock.AssignableToTypeOf(tmpFile)).
			Return(&model.Manifest{}, &model.Response{Error: nil}).
			Times(1)

		err = pluginAddCmdF(s.client, &cobra.Command{}, []string{pluginName})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Added plugin: "+pluginName)
	})

	s.Run("Add 1 plugin no file", func() {
		printer.Clean()
		err := pluginAddCmdF(s.client, &cobra.Command{}, []string{"non_existent_plugin"})
		s.Require().NotNil(err)
		s.Require().True(err.Error() == "open non_existent_plugin: no such file or directory" || err.Error() == "open non_existent_plugin: The system cannot find the file specified.")
	})

	s.Run("Add 1 plugin with error", func() {
		printer.Clean()
		tmpFile, err := ioutil.TempFile("", "tmpPlugin")
		s.Require().Nil(err)
		defer os.Remove(tmpFile.Name())

		pluginName := tmpFile.Name()
		mockError := &model.AppError{Message: "Plugin Add Error"}

		s.client.
			EXPECT().
			UploadPlugin(gomock.AssignableToTypeOf(tmpFile)).
			Return(&model.Manifest{}, &model.Response{Error: mockError}).
			Times(1)

		err = pluginAddCmdF(s.client, &cobra.Command{}, []string{pluginName})
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to add plugin: "+pluginName+". Error: "+mockError.Error())
	})

	s.Run("Add several plugins with some error", func() {
		printer.Clean()
		args := []string{"fail", "ok", "fail"}
		mockError := &model.AppError{Message: "Plugin Add Error"}

		for idx, arg := range args {
			tmpFile, err := ioutil.TempFile("", "tmpPlugin")
			s.Require().Nil(err)
			defer os.Remove(tmpFile.Name())
			if arg == "fail" {
				s.client.
					EXPECT().
					UploadPlugin(gomock.AssignableToTypeOf(tmpFile)).
					Return(nil, &model.Response{Error: mockError}).
					Times(1)
			} else {
				s.client.
					EXPECT().
					UploadPlugin(gomock.AssignableToTypeOf(tmpFile)).
					Return(&model.Manifest{}, &model.Response{Error: nil}).
					Times(1)
			}
			args[idx] = tmpFile.Name()
		}

		err := pluginAddCmdF(s.client, &cobra.Command{}, args)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Added plugin: "+args[1])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to add plugin: "+args[0]+". Error: "+mockError.Error())
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to add plugin: "+args[2]+". Error: "+mockError.Error())
	})
}

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

func (s *MmctlUnitTestSuite) TestPluginListCmd() {
	s.Run("List JSON plugins", func() {
		printer.Clean()
		mockList := &model.PluginsResponse{
			Active: []*model.PluginInfo{
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id1",
						Name:    "name1",
						Version: "v1",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id2",
						Name:    "name2",
						Version: "v2",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id3",
						Name:    "name3",
						Version: "v3",
					},
				},
			}, Inactive: []*model.PluginInfo{
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id4",
						Name:    "name4",
						Version: "v4",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id5",
						Name:    "name5",
						Version: "v5",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id6",
						Name:    "name6",
						Version: "v6",
					},
				},
			},
		}

		s.client.
			EXPECT().
			GetPlugins().
			Return(mockList, &model.Response{Error: nil}).
			Times(1)

		err := pluginListCmdF(s.client, &cobra.Command{}, nil)
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 8)

		s.Require().Equal("Listing enabled plugins", printer.GetLines()[0])
		for i, plugin := range mockList.Active {
			s.Require().Equal(plugin, printer.GetLines()[i+1])
		}

		s.Require().Equal("Listing disabled plugins", printer.GetLines()[4])
		for i, plugin := range mockList.Inactive {
			s.Require().Equal(plugin, printer.GetLines()[i+5])
		}
	})

	s.Run("List Plain Plugins", func() {
		printer.Clean()
		printer.SetFormat(printer.FORMAT_PLAIN)
		defer printer.SetFormat(printer.FORMAT_JSON)

		mockList := &model.PluginsResponse{
			Active: []*model.PluginInfo{
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id1",
						Name:    "name1",
						Version: "v1",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id2",
						Name:    "name2",
						Version: "v2",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id3",
						Name:    "name3",
						Version: "v3",
					},
				},
			}, Inactive: []*model.PluginInfo{
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id4",
						Name:    "name4",
						Version: "v4",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id5",
						Name:    "name5",
						Version: "v5",
					},
				},
				&model.PluginInfo{
					Manifest: model.Manifest{
						Id:      "id6",
						Name:    "name6",
						Version: "v6",
					},
				},
			},
		}

		s.client.
			EXPECT().
			GetPlugins().
			Return(mockList, &model.Response{Error: nil}).
			Times(1)

		err := pluginListCmdF(s.client, &cobra.Command{}, nil)
		s.Require().NoError(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 8)

		s.Require().Equal("Listing enabled plugins", printer.GetLines()[0])
		for i, plugin := range mockList.Active {
			s.Require().Equal(plugin.Id+": "+plugin.Name+", Version: "+plugin.Version, printer.GetLines()[i+1])
		}

		s.Require().Equal("Listing disabled plugins", printer.GetLines()[4])
		for i, plugin := range mockList.Inactive {
			s.Require().Equal(plugin.Id+": "+plugin.Name+", Version: "+plugin.Version, printer.GetLines()[i+5])
		}
	})

	s.Run("GetPlugins returns error", func() {
		printer.Clean()
		mockError := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			GetPlugins().
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := pluginListCmdF(s.client, &cobra.Command{}, nil)
		s.Require().NotNil(err)
		s.Require().Equal(err, errors.New("Unable to list plugins. Error: "+mockError.Error()))
	})
}

func (s *MmctlUnitTestSuite) TestPluginDeleteCmd() {
	s.Run("Delete one plugin with error", func() {
		printer.Clean()
		args := "plugin"
		mockError := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			RemovePlugin(args).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := pluginDeleteCmdF(s.client, &cobra.Command{}, []string{args})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to delete plugin: "+args+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
	})

	s.Run("Delete one plugin with no error", func() {
		printer.Clean()
		args := "plugin"

		s.client.
			EXPECT().
			RemovePlugin(args).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := pluginDeleteCmdF(s.client, &cobra.Command{}, []string{args})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("Deleted plugin: "+args, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete several plugins", func() {
		printer.Clean()
		args := []string{
			"plugin0",
			"error1",
			"error2",
			"plugin3",
		}
		mockErrors := []*model.AppError{
			&model.AppError{Message: "Mock Error1"},
			&model.AppError{Message: "Mock Error2"},
		}

		s.client.
			EXPECT().
			RemovePlugin(args[0]).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			RemovePlugin(args[1]).
			Return(false, &model.Response{Error: mockErrors[0]}).
			Times(1)

		s.client.
			EXPECT().
			RemovePlugin(args[2]).
			Return(false, &model.Response{Error: mockErrors[1]}).
			Times(1)

		s.client.
			EXPECT().
			RemovePlugin(args[3]).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := pluginDeleteCmdF(s.client, &cobra.Command{}, args)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal("Deleted plugin: "+args[0], printer.GetLines()[0])
		s.Require().Equal("Deleted plugin: "+args[3], printer.GetLines()[1])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("Unable to delete plugin: "+args[1]+". Error: "+mockErrors[0].Error(), printer.GetErrorLines()[0])
		s.Require().Equal("Unable to delete plugin: "+args[2]+". Error: "+mockErrors[1].Error(), printer.GetErrorLines()[1])
	})
}

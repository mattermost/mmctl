package commands

import (
	"errors"
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
		s.Require().Len(printer.GetLines(), 8)
		s.Require().Equal("Listing active plugins", printer.GetLines()[0])
		for i, plugin := range mockList.Active {
			s.Require().Equal(plugin, printer.GetLines()[i+1])
		}
		s.Require().Equal("Listing inactive plugins", printer.GetLines()[4])
		for i, plugin := range mockList.Inactive {
			s.Require().Equal(plugin, printer.GetLines()[i+5])
		}
		s.Require().Len(printer.GetErrorLines(), 0)
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

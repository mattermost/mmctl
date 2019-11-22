package commands

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestPluginAddCmd() {
	s.Run("Add without args", func() {
		err := pluginAddCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Error(err)
	})

	s.Run("Add 1 plugin", func() {
		printer.Clean()
		tmpFile, err := ioutil.TempFile("", "tmpPlugin")
		s.Require().Nil(err)
		defer os.Remove(tmpFile.Name())

		pluginName := tmpFile.Name()

		s.client.
			EXPECT().
			UploadPlugin(gomock.Any()).
			Return(&model.Manifest{}, &model.Response{Error: nil}).
			Times(1)

		err = pluginAddCmdF(s.client, &cobra.Command{}, []string{pluginName})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Added plugin: "+pluginName)
	})

	s.Run("Add 1 plugin no file", func() {
		err := pluginAddCmdF(s.client, &cobra.Command{}, []string{"non_existent_plugin"})
		s.Require().EqualError(err, "open non_existent_plugin: no such file or directory")
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
			UploadPlugin(gomock.Any()).
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
					UploadPlugin(gomock.Any()).
					Return(nil, &model.Response{Error: mockError}).
					Times(1)
			} else {
				s.client.
					EXPECT().
					UploadPlugin(gomock.Any()).
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

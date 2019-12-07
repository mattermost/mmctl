package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestConfigGetCmdF() {
	s.Run("Get nested config setting", func() {
		printer.Clean()

		configNameArg := "ServiceSettings.WebsocketSecurePort"
		configValue := 443
		mockConfig := &model.Config{
			ServiceSettings: model.ServiceSettings{
				WebsocketSecurePort: &configValue,
			},
		}

		s.client.
			EXPECT().
			GetConfig().
			Return(mockConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, []string{configNameArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(&configValue, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Get config setting block", func() {
		printer.Clean()

		configNameArg := "TeamSettings"
		configSiteName := "Mattermost"
		configMaxUsersPerTeam := 50
		mockTeamSettings := model.TeamSettings{
			SiteName:        &configSiteName,
			MaxUsersPerTeam: &configMaxUsersPerTeam,
		}
		mockConfig := &model.Config{
			TeamSettings: mockTeamSettings,
		}

		s.client.
			EXPECT().
			GetConfig().
			Return(mockConfig, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, []string{configNameArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(mockTeamSettings, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Attempt to get nonexistent setting key", func() {
		printer.Clean()

		configNameArg := "Does.Not.Exist"

		s.client.
			EXPECT().
			GetConfig().
			Return(&model.Config{}, &model.Response{Error: nil}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, []string{configNameArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Invalid key")
	})

	s.Run("Fail to communicate with server while getting config", func() {
		printer.Clean()

		configNameArg := "ServiceSettings.WebsocketSecurePort"
		configValue := 443
		mockConfig := &model.Config{
			ServiceSettings: model.ServiceSettings{
				WebsocketSecurePort: &configValue,
			},
		}

		mockErrorMessage := "Mock Internal Server Error"
		mockError := &model.AppError{Message: mockErrorMessage}

		s.client.
			EXPECT().
			GetConfig().
			Return(mockConfig, &model.Response{Error: mockError}).
			Times(1)

		err := configGetCmdF(s.client, &cobra.Command{}, []string{configNameArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, mockError.Error())
	})
}

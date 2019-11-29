package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestLdapSyncCmd() {
	s.Run("Sync without errors", func() {
		printer.Clean()
		outputMessage := map[string]interface{}{"status": "ok"}

		s.client.
			EXPECT().
			SyncLdap().
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := ldapSyncCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Not able to Sync", func() {
		printer.Clean()
		outputMessage := map[string]interface{}{"status": "error"}

		s.client.
			EXPECT().
			SyncLdap().
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := ldapSyncCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Sync with response error", func() {
		printer.Clean()
		mockError := &model.AppError{Message: "Mock Error"}

		s.client.
			EXPECT().
			SyncLdap().
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := ldapSyncCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Equal(err, mockError)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

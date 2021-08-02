// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestLdapSyncCmd() {
	s.Run("Sync without errors", func() {
		printer.Clean()
		outputMessage := map[string]interface{}{"status": "ok"}

		s.client.
			EXPECT().
			SyncLdap(false).
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
			SyncLdap(false).
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
			SyncLdap(false).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := ldapSyncCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Equal(err, mockError)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Sync with includeRemoveMembers", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("include-removed-members", true, "")

		s.client.
			EXPECT().
			SyncLdap(true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := ldapSyncCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
	})
}

func (s *MmctlUnitTestSuite) TestLdapMigrateID() {
	s.Run("Run successfully without errors", func() {
		printer.Clean()

		s.client.
			EXPECT().
			MigrateIdLdap("test-id").
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := ldapIDMigrateCmdF(s.client, &cobra.Command{}, []string{"test-id"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Contains(printer.GetLines()[0], "test-id")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Unable to migrate", func() {
		printer.Clean()

		s.client.
			EXPECT().
			MigrateIdLdap("test-id").
			Return(false, &model.Response{Error: &model.AppError{Message: "test-error"}}).
			Times(1)

		err := ldapIDMigrateCmdF(s.client, &cobra.Command{}, []string{"test-id"})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
	})
}

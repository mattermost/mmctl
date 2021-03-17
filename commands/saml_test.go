// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestSamlAuthDataReset() {
	s.Run("Reset auth data without errors", func() {
		printer.Clean()
		outputMessage := "1 user records were changed.\n"

		s.client.
			EXPECT().
			ResetSamlAuthDataToEmail(false, false, []string{}).
			Return(int64(1), &model.Response{Error: nil})

		err := samlAuthDataResetCmdF(s.client, &cobra.Command{}, nil)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
	s.Run("Reset auth data dry run", func() {
		printer.Clean()
		outputMessage := "1 user records would be affected.\n"

		cmd := &cobra.Command{}
		cmd.Flags().Bool("dry-run", true, "")

		s.client.
			EXPECT().
			ResetSamlAuthDataToEmail(false, true, []string{}).
			Return(int64(1), &model.Response{Error: nil})

		err := samlAuthDataResetCmdF(s.client, cmd, nil)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

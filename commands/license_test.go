// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlUnitTestSuite) TestRemoveLicenseCmd() {
	s.Run("Remove license successfully", func() {
		printer.Clean()

		s.client.
			EXPECT().
			RemoveLicenseFile().
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := removeLicenseCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(printer.GetLines()[0], "Removed license")
	})

	s.Run("Fail to remove license", func() {
		printer.Clean()
		mockErr := &model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			RemoveLicenseFile().
			Return(false, &model.Response{Error: mockErr}).
			Times(1)

		err := removeLicenseCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(err, mockErr)
	})
}

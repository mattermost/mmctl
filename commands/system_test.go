// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestGetBusyCmd() {
	s.Run("GetBusy when not set", func() {
		printer.Clean()
		outputMessage := "busy:false expires:0"

		s.client.
			EXPECT().
			GetServerBusy().
			Return(&model.ServerBusyState{}, &model.Response{Error: nil}).
			Times(1)

		err := getBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("GetBusy when set", func() {
		printer.Clean()
		expires := time.Now().Add(time.Minute * 15).Unix()
		outputMessage := fmt.Sprintf("busy:%t expires:%d", true, expires)

		s.client.
			EXPECT().
			GetServerBusy().
			Return(&model.ServerBusyState{Busy: true, Expires: expires}, &model.Response{Error: nil}).
			Times(1)

		err := archiveCommandCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], outputMessage)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestGetBusyCmd() {
	s.Run("GetBusy when not set", func() {
		printer.Clean()
		sbs := &model.ServerBusyState{}

		s.client.
			EXPECT().
			GetServerBusy().
			Return(sbs, &model.Response{Error: nil}).
			Times(1)

		err := getBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], sbs)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("GetBusy when set", func() {
		printer.Clean()
		const minutes = 15
		expires := time.Now().Add(time.Minute * minutes).Unix()
		sbs := &model.ServerBusyState{Busy: true, Expires: expires}

		s.client.
			EXPECT().
			GetServerBusy().
			Return(sbs, &model.Response{Error: nil}).
			Times(1)

		err := getBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], sbs)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("GetBusy with error", func() {
		printer.Clean()
		s.client.
			EXPECT().
			GetServerBusy().
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := getBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
	})
}

func (s *MmctlUnitTestSuite) TestSetBusyCmd() {
	s.Run("SetBusy 900 seconds", func() {
		printer.Clean()
		const minutes = 15

		s.client.
			EXPECT().
			SetServerBusy(minutes*60).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := setBusyCmdF(s.client, &cobra.Command{}, []string{strconv.Itoa(minutes * 60)})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Busy state set")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("SetBusy with missing arg", func() {
		printer.Clean()
		s.client.
			EXPECT().
			SetServerBusy(3600). // endpoint defaults to 3600 seconds
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := setBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Busy state set")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("SetBusy invalid seconds", func() {
		printer.Clean()

		err := setBusyCmdF(s.client, &cobra.Command{}, []string{strconv.Itoa(-1)})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
	})

	s.Run("SetBusy zero seconds", func() {
		printer.Clean()

		err := setBusyCmdF(s.client, &cobra.Command{}, []string{strconv.Itoa(0)})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
	})
}

func (s *MmctlUnitTestSuite) TestClearBusyCmd() {
	s.Run("ClearBusy", func() {
		printer.Clean()
		s.client.
			EXPECT().
			ClearServerBusy().
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := clearBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Busy state cleared")
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("ClearBusy with error", func() {
		printer.Clean()
		s.client.
			EXPECT().
			ClearServerBusy().
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := clearBusyCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
	})
}

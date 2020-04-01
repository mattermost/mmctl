// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"bytes"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
	"io/ioutil"

	"github.com/mattermost/mmctl/printer"
)

const (
	fakeLicensePayload = "This is the license."
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

func ReadFile(filename string) ([]byte, error) {
	buf := bytes.NewBufferString(fakeLicensePayload)
	return ioutil.ReadAll(buf)
}

func (s *MmctlUnitTestSuite) TestUploadLicenseCmdF() {
	s.Run("Upload license successfully", func() {
		printer.Clean()
		path := "/path/to/file"
		mockLicenseFile := []byte(fakeLicensePayload)
		customReadFile = ReadFile
		s.client.
			EXPECT().
			UploadLicenseFile(mockLicenseFile).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := uploadLicenseCmdF(s.client, &cobra.Command{}, []string{path})
		s.Require().Nil(err)
	})

	s.Run("Fail to upload license if file not found", func() {
		printer.Clean()

		path := "/path/to/nonexistentfile"
		mockLicenseFile := []byte(fakeLicensePayload)
		errMsg := "open " + path + ": no such file or directory"
		mockErr := &model.AppError{Message: errMsg}
		customReadFile = ReadFile
		s.client.
			EXPECT().
			UploadLicenseFile(mockLicenseFile).
			Return(false, &model.Response{Error: mockErr}).
			Times(1)

		err := uploadLicenseCmdF(s.client, &cobra.Command{}, []string{path})
		s.Require().EqualError(err, ": "+errMsg+", ")
	})

	s.Run("Fail to upload license if no path is given", func() {
		printer.Clean()
		err := uploadLicenseCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().EqualError(err, fmt.Sprintf("enter one license file to upload"))
	})
}

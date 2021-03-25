// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"net/http"

	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestImportListAvailableCmdF() {
	s.Run("no imports", func() {
		printer.Clean()
		var mockImports []string

		s.client.
			EXPECT().
			ListImports().
			Return(mockImports, &model.Response{Error: nil}).
			Times(1)

		err := importListAvailableCmdF(s.client, &cobra.Command{}, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Equal("No import files found", printer.GetLines()[0])
	})

	s.Run("some imports", func() {
		printer.Clean()
		mockImports := []string{
			"import1.zip",
			"import2.zip",
			"import3.zip",
		}

		s.client.
			EXPECT().
			ListImports().
			Return(mockImports, &model.Response{Error: nil}).
			Times(1)

		err := importListAvailableCmdF(s.client, &cobra.Command{}, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), len(mockImports))
		s.Len(printer.GetErrorLines(), 0)
		for i, line := range printer.GetLines() {
			s.Equal(mockImports[i], line)
		}
	})
}

func (s *MmctlUnitTestSuite) TestImportListIncompleteCmdF() {
	s.Run("no incomplete uploads", func() {
		printer.Clean()
		var mockUploads []*model.UploadSession

		s.client.
			EXPECT().
			GetUploadsForUser("me").
			Return(mockUploads, &model.Response{Error: nil}).
			Times(1)

		err := importListIncompleteCmdF(s.client, &cobra.Command{}, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Empty(printer.GetErrorLines())
		s.Equal("No incomplete import uploads found", printer.GetLines()[0])
	})

	s.Run("some incomplete uploads", func() {
		printer.Clean()
		mockUploads := []*model.UploadSession{
			{
				Id:   model.NewId(),
				Type: model.UploadTypeImport,
			},
			{
				Id:   model.NewId(),
				Type: model.UploadTypeAttachment,
			},
			{
				Id:   model.NewId(),
				Type: model.UploadTypeImport,
			},
		}

		s.client.
			EXPECT().
			GetUploadsForUser("me").
			Return(mockUploads, &model.Response{Error: nil}).
			Times(1)

		err := importListIncompleteCmdF(s.client, &cobra.Command{}, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 2)
		s.Empty(printer.GetErrorLines())
		s.Equal(mockUploads[0], printer.GetLines()[0].(*model.UploadSession))
		s.Equal(mockUploads[2], printer.GetLines()[1].(*model.UploadSession))
	})
}

func (s *MmctlUnitTestSuite) TestImportJobShowCmdF() {
	s.Run("not found", func() {
		printer.Clean()

		jobID := model.NewId()

		s.client.
			EXPECT().
			GetJob(jobID).
			Return(nil, &model.Response{Error: model.NewAppError("not found", "", nil, "not found", http.StatusNotFound)}).
			Times(1)

		err := importJobShowCmdF(s.client, &cobra.Command{}, []string{jobID})
		s.Require().NotNil(err)
		s.Empty(printer.GetLines())
		s.Empty(printer.GetErrorLines())
	})

	s.Run("found", func() {
		printer.Clean()
		mockJob := &model.Job{
			Id: model.NewId(),
		}

		s.client.
			EXPECT().
			GetJob(mockJob.Id).
			Return(mockJob, &model.Response{Error: nil}).
			Times(1)

		err := importJobShowCmdF(s.client, &cobra.Command{}, []string{mockJob.Id})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Empty(printer.GetErrorLines())
		s.Equal(mockJob, printer.GetLines()[0].(*model.Job))
	})
}

func (s *MmctlUnitTestSuite) TestImportJobListCmdF() {
	s.Run("no import jobs", func() {
		printer.Clean()
		var mockJobs []*model.Job

		cmd := &cobra.Command{}
		perPage := 10
		cmd.Flags().Int("page", 0, "")
		cmd.Flags().Int("per-page", perPage, "")
		cmd.Flags().Bool("all", false, "")

		s.client.
			EXPECT().
			GetJobsByType(model.JOB_TYPE_IMPORT_PROCESS, 0, perPage).
			Return(mockJobs, &model.Response{Error: nil}).
			Times(1)

		err := importJobListCmdF(s.client, cmd, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Empty(printer.GetErrorLines())
		s.Equal("No jobs found", printer.GetLines()[0])
	})

	s.Run("some import jobs", func() {
		printer.Clean()
		mockJobs := []*model.Job{
			{
				Id: model.NewId(),
			},
			{
				Id: model.NewId(),
			},
			{
				Id: model.NewId(),
			},
		}

		cmd := &cobra.Command{}
		perPage := 3
		cmd.Flags().Int("page", 0, "")
		cmd.Flags().Int("per-page", perPage, "")
		cmd.Flags().Bool("all", false, "")

		s.client.
			EXPECT().
			GetJobsByType(model.JOB_TYPE_IMPORT_PROCESS, 0, perPage).
			Return(mockJobs, &model.Response{Error: nil}).
			Times(1)

		err := importJobListCmdF(s.client, cmd, nil)
		s.Require().Nil(err)
		s.Len(printer.GetLines(), len(mockJobs))
		s.Empty(printer.GetErrorLines())
		for i, line := range printer.GetLines() {
			s.Equal(mockJobs[i], line.(*model.Job))
		}
	})
}

func (s *MmctlUnitTestSuite) TestImportProcessCmdF() {
	printer.Clean()
	importFile := "import.zip"
	mockJob := &model.Job{
		Type: model.JOB_TYPE_IMPORT_PROCESS,
		Data: map[string]string{"import_file": importFile},
	}

	s.client.
		EXPECT().
		CreateJob(mockJob).
		Return(mockJob, &model.Response{Error: nil}).
		Times(1)

	err := importProcessCmdF(s.client, &cobra.Command{}, []string{importFile})
	s.Require().Nil(err)
	s.Len(printer.GetLines(), 1)
	s.Empty(printer.GetErrorLines())
	s.Equal(mockJob, printer.GetLines()[0].(*model.Job))
}

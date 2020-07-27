// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"

	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestIntegrityCmd() {
	s.Run("Integrity check succeeds", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		mockData := model.RelationalIntegrityCheckData{
			ParentName:   "parent",
			ChildName:    "child",
			ParentIdAttr: "parentIdAttr",
			ChildIdAttr:  "childIdAttr",
			Records: []model.OrphanedRecord{
				{
					ParentId: model.NewString("parentId"),
					ChildId:  model.NewString("childId"),
				},
			},
		}
		mockResults := []model.IntegrityCheckResult{
			{
				Data: mockData,
				Err:  nil,
			},
		}
		s.client.
			EXPECT().
			CheckIntegrity().
			Return(mockResults, &model.Response{Error: nil}).
			Times(1)

		err := integrityCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(mockData, printer.GetLines()[0])
	})

	s.Run("Integrity check fails", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			CheckIntegrity().
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := integrityCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal("unable to perform integrity check. Error: : , ", err.Error())
	})

	s.Run("Integrity check with errors", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		mockData := model.RelationalIntegrityCheckData{
			ParentName:   "parent",
			ChildName:    "child",
			ParentIdAttr: "parentIdAttr",
			ChildIdAttr:  "childIdAttr",
			Records: []model.OrphanedRecord{
				{
					ParentId: model.NewString("parentId"),
					ChildId:  model.NewString("childId"),
				},
			},
		}
		mockResults := []model.IntegrityCheckResult{
			{
				Data: nil,
				Err:  errors.New("test error"),
			},
			{
				Data: mockData,
				Err:  nil,
			},
		}
		s.client.
			EXPECT().
			CheckIntegrity().
			Return(mockResults, &model.Response{Error: nil}).
			Times(1)

		err := integrityCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(mockData, printer.GetLines()[0])
		s.Require().Equal("test error", printer.GetErrorLines()[0])
	})
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

const fakeUserID = "nouser"

var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Management of imports",
}

var ImportUploadCmd = &cobra.Command{
	Use:     "upload [filepath]",
	Short:   "Upload import files",
	Example: " import upload import_file.zip",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(importUploadCmdF),
}

var ImportListCmd = &cobra.Command{
	Use:   "list",
	Short: "List import files",
}

var ImportListAvailableCmd = &cobra.Command{
	Use:     "available",
	Short:   "List available import files",
	Example: " import list available",
	Args:    cobra.NoArgs,
	RunE:    withClient(importListAvailableCmdF),
}

var ImportListIncompleteCmd = &cobra.Command{
	Use:     "incomplete",
	Short:   "List incomplete import files uploads",
	Example: " import list incomplete",
	Args:    cobra.NoArgs,
	RunE:    withClient(importListIncompleteCmdF),
}

var ImportListJobsCmd = &cobra.Command{
	Use:     "jobs [importJobID]",
	Example: " import list jobs",
	Short:   "List import jobs",
	Args:    cobra.MaximumNArgs(1),
	RunE:    withClient(importListJobsCmdF),
}

var ImportProcessCmd = &cobra.Command{
	Use:     "process [importname]",
	Example: " import process 35uy6cwrqfnhdx3genrhqqznxc_import.zip",
	Short:   "Start an import job",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(importProcessCmdF),
}

func init() {
	ImportUploadCmd.Flags().Bool("resume", false, "Set to true to resume an incomplete import upload.")
	ImportUploadCmd.Flags().String("upload", "", "The ID of the import upload to resume.")
	ImportListJobsCmd.Flags().Int("limit", 10, "The maximum number of jobs to show.")
	ImportListCmd.AddCommand(
		ImportListAvailableCmd,
		ImportListIncompleteCmd,
		ImportListJobsCmd,
	)
	ImportCmd.AddCommand(
		ImportUploadCmd,
		ImportListCmd,
		ImportProcessCmd,
	)
	RootCmd.AddCommand(ImportCmd)
}

func importListIncompleteCmdF(c client.Client, command *cobra.Command, args []string) error {
	isLocal, _ := command.Flags().GetBool("local")
	userID := "me"
	if isLocal {
		userID = fakeUserID
	}

	uploads, resp := c.GetUploadsForUser(userID)
	if resp.Error != nil {
		return fmt.Errorf("failed to get uploads: %w", resp.Error)
	}

	var imports []*model.UploadSession
	for i := range uploads {
		if uploads[i].Type == model.UploadTypeImport {
			imports = append(imports, uploads[i])
		}
	}

	if len(imports) == 0 {
		printer.Print("No incomplete import uploads found")
		return nil
	}

	for _, us := range imports {
		completedPct := float64(us.FileOffset) / float64(us.FileSize) * 100
		printer.PrintT(fmt.Sprintf("  ID: {{.Id}}\n  Name: {{.Filename}}\n  Uploaded: {{.FileOffset}}/{{.FileSize}} (%0.0f%%)\n", completedPct), us)
	}

	return nil
}

func importListAvailableCmdF(c client.Client, command *cobra.Command, args []string) error {
	imports, resp := c.ListImports()
	if resp.Error != nil {
		return fmt.Errorf("failed to list imports: %w", resp.Error)
	}

	if len(imports) == 0 {
		printer.Print("No import files found")
		return nil
	}

	for _, name := range imports {
		printer.Print(name)
	}

	return nil
}

func importUploadCmdF(c client.Client, command *cobra.Command, args []string) error {
	filepath := args[0]

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat import file: %w", err)
	}

	shouldResume, _ := command.Flags().GetBool("resume")
	var us *model.UploadSession
	var resp *model.Response
	if shouldResume {
		uploadID, err := command.Flags().GetString("upload")
		if err != nil || !model.IsValidId(uploadID) {
			return errors.New("required upload ID is missing or invalid")
		}

		us, resp = c.GetUpload(uploadID)
		if resp.Error != nil {
			return fmt.Errorf("failed to get upload: %w", resp.Error)
		}

		if us.FileSize != info.Size() {
			return fmt.Errorf("file sizes do not match")
		}

		if _, err := file.Seek(us.FileOffset, io.SeekStart); err != nil {
			return fmt.Errorf("failed to get seek file: %w", err)
		}
	} else {
		isLocal, _ := command.Flags().GetBool("local")
		userID := "me"
		if isLocal {
			userID = fakeUserID
		}

		us, resp = c.CreateUpload(&model.UploadSession{
			Filename: info.Name(),
			FileSize: info.Size(),
			Type:     model.UploadTypeImport,
			UserId:   userID,
		})
		if resp.Error != nil {
			return fmt.Errorf("failed to create upload: %w", resp.Error)
		}

		printer.PrintT("Upload successfully created, ID: {{.Id}} ", us)
	}

	finfo, resp := c.UploadData(us.Id, file)
	if resp.Error != nil {
		return fmt.Errorf("failed to upload data: %w", resp.Error)
	}

	printer.PrintT("Import file successfully uploaded, name: {{.Id}}", finfo)

	return nil
}

func importProcessCmdF(c client.Client, command *cobra.Command, args []string) error {
	importFile := args[0]

	job, resp := c.CreateJob(&model.Job{
		Type: model.JOB_TYPE_IMPORT_PROCESS,
		Data: map[string]string{
			"import_file": importFile,
		},
	})
	if resp.Error != nil {
		return fmt.Errorf("failed to create import process job: %w", resp.Error)
	}

	printer.PrintT("Import process job successfully created, ID: {{.Id}}", job)

	return nil
}

func importListJobsCmdF(c client.Client, command *cobra.Command, args []string) error {
	var jobs []*model.Job
	var resp *model.Response
	if len(args) == 1 {
		var job *model.Job
		job, resp = c.GetJob(args[0])
		if resp.Error != nil {
			return fmt.Errorf("failed to get import job: %w", resp.Error)
		}
		jobs = append(jobs, job)
	} else {
		numJobs, _ := command.Flags().GetInt("limit")
		jobs, resp = c.GetJobsByType(model.JOB_TYPE_IMPORT_PROCESS, 0, numJobs)
		if resp.Error != nil {
			return fmt.Errorf("failed to get import jobs: %w", resp.Error)
		}

		if len(jobs) == 0 {
			printer.Print("No import jobs found")
			return nil
		}
	}

	for _, job := range jobs {
		printer.PrintT(fmt.Sprintf("  ID: {{.Id}}\n  Status: {{.Status}}\n  Created: %s\n  Started: %s\n",
			time.Unix(job.CreateAt/1000, 0), time.Unix(job.StartAt/1000, 0)), job)
	}

	return nil
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/spf13/cobra"
)

var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Management of imports",
}

var ImportUploadCmd = &cobra.Command{
	Use:     "upload [filepath]",
	Short:   "Upload import files",
	Example: "  import upload import_file.zip",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(importUploadCmdF),
}

var ImportListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List import files",
	Example: " import list",
}

var ImportListAvailableCmd = &cobra.Command{
	Use:     "available",
	Short:   "List available import files",
	Example: "  import list available",
	Args:    cobra.NoArgs,
	RunE:    withClient(importListAvailableCmdF),
}

var ImportJobCmd = &cobra.Command{
	Use:   "job",
	Short: "List and show import jobs",
}

var ImportListIncompleteCmd = &cobra.Command{
	Use:     "incomplete",
	Short:   "List incomplete import files uploads",
	Example: "  import list incomplete",
	Args:    cobra.NoArgs,
	RunE:    withClient(importListIncompleteCmdF),
}

var ImportJobListCmd = &cobra.Command{
	Use:     "list",
	Example: "  import job list",
	Short:   "List import jobs",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    withClient(importJobListCmdF),
}

var ImportJobShowCmd = &cobra.Command{
	Use:     "show [importJobID]",
	Example: " import job show f3d68qkkm7n8xgsfxwuo498rah",
	Short:   "Show import job",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(importJobShowCmdF),
}

var ImportProcessCmd = &cobra.Command{
	Use:     "process [importname]",
	Example: "  import process 35uy6cwrqfnhdx3genrhqqznxc_import.zip",
	Short:   "Start an import job",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(importProcessCmdF),
}

func init() {
	ImportUploadCmd.Flags().Bool("resume", false, "Set to true to resume an incomplete import upload.")
	ImportUploadCmd.Flags().String("upload", "", "The ID of the import upload to resume.")

	ImportJobListCmd.Flags().Int("page", 0, "Page number to fetch for the list of import jobs")
	ImportJobListCmd.Flags().Int("per-page", 200, "Number of import jobs to be fetched")
	ImportJobListCmd.Flags().Bool("all", false, "Fetch all import jobs. --page flag will be ignore if provided")

	ImportListCmd.AddCommand(
		ImportListAvailableCmd,
		ImportListIncompleteCmd,
	)
	ImportJobCmd.AddCommand(
		ImportJobListCmd,
		ImportJobShowCmd,
	)
	ImportCmd.AddCommand(
		ImportUploadCmd,
		ImportListCmd,
		ImportProcessCmd,
		ImportJobCmd,
	)
	RootCmd.AddCommand(ImportCmd)
}

func importListIncompleteCmdF(c client.Client, command *cobra.Command, args []string) error {
	isLocal, _ := command.Flags().GetBool("local")
	userID := "me"
	if isLocal {
		userID = model.UploadNoUserID
	}

	uploads, _, err := c.GetUploadsForUser(userID)
	if err != nil {
		return fmt.Errorf("failed to get uploads: %w", err)
	}

	var hasImports bool
	for _, us := range uploads {
		if us.Type == model.UploadTypeImport {
			completedPct := float64(us.FileOffset) / float64(us.FileSize) * 100
			printer.PrintT(fmt.Sprintf("  ID: {{.Id}}\n  Name: {{.Filename}}\n  Uploaded: {{.FileOffset}}/{{.FileSize}} (%0.0f%%)\n", completedPct), us)
			hasImports = true
		}
	}

	if !hasImports {
		printer.Print("No incomplete import uploads found")
		return nil
	}

	return nil
}

func importListAvailableCmdF(c client.Client, command *cobra.Command, args []string) error {
	imports, _, err := c.ListImports()
	if err != nil {
		return fmt.Errorf("failed to list imports: %w", err)
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
	if shouldResume {
		uploadID, nErr := command.Flags().GetString("upload")
		if nErr != nil || !model.IsValidId(uploadID) {
			return errors.New("upload session ID is missing or invalid")
		}

		us, _, err = c.GetUpload(uploadID)
		if err != nil {
			return fmt.Errorf("failed to get upload session: %w", err)
		}

		if us.FileSize != info.Size() {
			return fmt.Errorf("file sizes do not match")
		}

		if _, nErr := file.Seek(us.FileOffset, io.SeekStart); nErr != nil {
			return fmt.Errorf("failed to get seek file: %w", nErr)
		}
	} else {
		isLocal, _ := command.Flags().GetBool("local")
		userID := "me"
		if isLocal {
			userID = model.UploadNoUserID
		}

		us, _, err = c.CreateUpload(&model.UploadSession{
			Filename: info.Name(),
			FileSize: info.Size(),
			Type:     model.UploadTypeImport,
			UserId:   userID,
		})
		if err != nil {
			return fmt.Errorf("failed to create upload session: %w", err)
		}

		printer.PrintT("Upload session successfully created, ID: {{.Id}} ", us)
	}

	finfo, _, err := c.UploadData(us.Id, file)
	if err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}

	printer.PrintT("Import file successfully uploaded, name: {{.Id}}", finfo)

	return nil
}

func importProcessCmdF(c client.Client, command *cobra.Command, args []string) error {
	importFile := args[0]

	job, _, err := c.CreateJob(&model.Job{
		Type: model.JobTypeImportProcess,
		Data: map[string]string{
			"import_file": importFile,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create import process job: %w", err)
	}

	printer.PrintT("Import process job successfully created, ID: {{.Id}}", job)

	return nil
}

func printJob(job *model.Job) {
	if job.StartAt > 0 {
		printer.PrintT(fmt.Sprintf("  ID: {{.Id}}\n  Status: {{.Status}}\n  Created: %s\n  Started: %s\n",
			time.Unix(job.CreateAt/1000, 0), time.Unix(job.StartAt/1000, 0)), job)
	} else {
		printer.PrintT(fmt.Sprintf("  ID: {{.Id}}\n  Status: {{.Status}}\n  Created: %s\n\n",
			time.Unix(job.CreateAt/1000, 0)), job)
	}
}

func importJobShowCmdF(c client.Client, command *cobra.Command, args []string) error {
	job, _, err := c.GetJob(args[0])
	if err != nil {
		return fmt.Errorf("failed to get import job: %w", err)
	}

	printJob(job)

	return nil
}

func jobListCmdF(c client.Client, command *cobra.Command, jobType string) error {
	page, err := command.Flags().GetInt("page")
	if err != nil {
		return err
	}
	perPage, err := command.Flags().GetInt("per-page")
	if err != nil {
		return err
	}
	showAll, err := command.Flags().GetBool("all")
	if err != nil {
		return err
	}

	if showAll {
		page = 0
	}

	for {
		jobs, _, err := c.GetJobsByType(jobType, page, perPage)
		if err != nil {
			return fmt.Errorf("failed to get jobs: %w", err)
		}

		if len(jobs) == 0 {
			if !showAll || page == 0 {
				printer.Print("No jobs found")
			}
			return nil
		}

		for _, job := range jobs {
			printJob(job)
		}

		if !showAll {
			break
		}

		page++
	}

	return nil
}

func importJobListCmdF(c client.Client, command *cobra.Command, args []string) error {
	return jobListCmdF(c, command, model.JobTypeImportProcess)
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Management of exports",
}

var ExportCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create export file",
	Args:  cobra.NoArgs,
	RunE:  withClient(exportCreateCmdF),
}

var ExportDownloadCmd = &cobra.Command{
	Use:   "download [exportname] [filepath]",
	Short: "Download export files",
	Example: `  # you can indicate the name of the export and its destination path
  $ mmctl export download samplename sample_export.zip
  
  # or if you only indicate the name, the path would match it
  $ mmctl export download sample_export.zip`,
	Args: cobra.MinimumNArgs(1),
	RunE: withClient(exportDownloadCmdF),
}

var ExportDeleteCmd = &cobra.Command{
	Use:     "delete [exportname]",
	Aliases: []string{"rm"},
	Example: "  export delete export_file.zip",
	Short:   "Delete export file",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(exportDeleteCmdF),
}

var ExportListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List export files",
	Args:    cobra.NoArgs,
	RunE:    withClient(exportListCmdF),
}

var ExportJobCmd = &cobra.Command{
	Use:   "job",
	Short: "List and show export jobs",
}

var ExportJobListCmd = &cobra.Command{
	Use:     "list",
	Example: "  export job list",
	Short:   "List export jobs",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    withClient(exportJobListCmdF),
}

var ExportJobShowCmd = &cobra.Command{
	Use:     "show [exportJobID]",
	Example: "  export job show",
	Short:   "Show export job",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(exportJobShowCmdF),
}

func init() {
	ExportCreateCmd.Flags().Bool("attachments", false, "Set to true to include file attachments in the export file.")

	ExportDownloadCmd.Flags().Bool("resume", false, "Set to true to resume an export download.")

	ExportJobListCmd.Flags().Int("page", 0, "Page number to fetch for the list of export jobs")
	ExportJobListCmd.Flags().Int("per-page", 200, "Number of export jobs to be fetched")
	ExportJobListCmd.Flags().Bool("all", false, "Fetch all export jobs. --page flag will be ignore if provided")

	ExportJobCmd.AddCommand(
		ExportJobListCmd,
		ExportJobShowCmd,
	)
	ExportCmd.AddCommand(
		ExportCreateCmd,
		ExportListCmd,
		ExportDeleteCmd,
		ExportDownloadCmd,
		ExportJobCmd,
	)
	RootCmd.AddCommand(ExportCmd)
}

func exportCreateCmdF(c client.Client, command *cobra.Command, args []string) error {
	var data map[string]string
	withAttachments, _ := command.Flags().GetBool("attachments")
	if withAttachments {
		data = map[string]string{
			"include_attachments": "true",
		}
	}

	job, resp := c.CreateJob(&model.Job{
		Type: model.JOB_TYPE_EXPORT_PROCESS,
		Data: data,
	})
	if resp.Error != nil {
		return fmt.Errorf("failed to create export process job: %w", resp.Error)
	}

	printer.PrintT("Export process job successfully created, ID: {{.Id}}", job)

	return nil
}

func exportListCmdF(c client.Client, command *cobra.Command, args []string) error {
	exports, resp := c.ListExports()
	if resp.Error != nil {
		return fmt.Errorf("failed to list exports: %w", resp.Error)
	}

	if len(exports) == 0 {
		printer.Print("No export files found")
		return nil
	}

	for _, name := range exports {
		printer.Print(name)
	}

	return nil
}

func exportDeleteCmdF(c client.Client, command *cobra.Command, args []string) error {
	name := args[0]

	if _, resp := c.DeleteExport(name); resp.Error != nil {
		return fmt.Errorf("failed to delete export: %w", resp.Error)
	}

	printer.Print(fmt.Sprintf("Export file %q has been deleted", name))

	return nil
}

func exportDownloadCmdF(c client.Client, command *cobra.Command, args []string) error {
	var path string
	name := args[0]
	if len(args) > 1 {
		path = args[1]
	}
	if path == "" {
		path = name
	}

	resume, _ := command.Flags().GetBool("resume")

	var outFile *os.File
	info, err := os.Stat(path)
	switch {
	case err != nil && !os.IsNotExist(err):
		// some error occurred and not because file doesn't exist
		return fmt.Errorf("failed to stat export file: %w", err)
	case err == nil && info.Size() > 0 && !resume:
		// we exit to avoid overwriting an existing non-empty file
		return fmt.Errorf("export file already exists")
	case os.IsNotExist(err) && resume:
		// cannot resume if the file does not exist
		return fmt.Errorf("cannot resume download: export file does not exist")
	case err != nil:
		// file does not exist, we create it
		outFile, err = os.Create(path)
	default:
		// no error, file exists, we open it
		outFile, err = os.OpenFile(path, os.O_WRONLY, 0600)
	}

	if err != nil {
		return fmt.Errorf("failed to create/open export file: %w", err)
	}
	defer outFile.Close()

	off, err := outFile.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek export file: %w", err)
	}

	if _, resp := c.DownloadExport(name, outFile, off); resp.Error != nil {
		return fmt.Errorf("failed to download export file: %w", resp.Error)
	}

	return nil
}

func exportJobListCmdF(c client.Client, command *cobra.Command, args []string) error {
	return jobListCmdF(c, command, model.JOB_TYPE_EXPORT_PROCESS)
}

func exportJobShowCmdF(c client.Client, command *cobra.Command, args []string) error {
	job, resp := c.GetJob(args[0])
	if resp.Error != nil {
		return fmt.Errorf("failed to get export job: %w", resp.Error)
	}

	printJob(job)

	return nil
}

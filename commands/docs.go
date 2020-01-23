// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var DocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generates mmctl documentation",
	Args:  cobra.NoArgs,
	RunE:  docsCmdF,
}

func init() {
	DocsCmd.Flags().StringP("directory", "d", "docs", "The directory where the docs would be generated in.")

	RootCmd.AddCommand(DocsCmd)
}

func docsCmdF(cmd *cobra.Command, args []string) error {
	outDir, _ := cmd.Flags().GetString("directory")

	if fileInfo, err := os.Stat(outDir); os.IsNotExist(err) {
		if createErr := os.Mkdir(outDir, 0755); createErr != nil {
			return createErr
		}
	} else if err != nil {
		return err
	} else if !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("File \"%s\" is not a directory", outDir))
	}

	err := doc.GenMarkdownTree(RootCmd, outDir)
	if err != nil {
		return err
	}

	return nil
}

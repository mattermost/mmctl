// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var (
	// SHA1 from git, output of $(git rev-parse HEAD)
	BuildHash = "dev mode"
	// Version of the application
	Version = "6.2.5"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of mmctl.",
	Args:  cobra.NoArgs,
	Run:   versionCmdF,
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}

func versionCmdF(cmd *cobra.Command, args []string) {
	printer.Print("mmctl v" + Version + " -- " + BuildHash)
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var (
	BuildHash = "dev mode"
	Version   = "6.1.4"
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

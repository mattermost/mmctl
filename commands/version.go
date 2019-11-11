package commands

import (
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var (
	BuildHash = "dev mode"
	Version   = "0.1.0"
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

package commands

import (
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

type Command = cobra.Command

func Run(args []string) error {
	RootCmd.PersistentFlags().String("format", "plain", "the format of the command output [plain, json]")
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:   "mmctl",
	Short: "Remote client for the Open Source, self-hosted Slack-alternative",
	Long:  `Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com`,
	PersistentPreRun: func(command *cobra.Command, args []string) {
		format, _ := command.Flags().GetString("format")
		printer.SetFormat(format)
	},
	PersistentPostRun: func(command *cobra.Command, args []string) {
		printer.Flush()
	},
}

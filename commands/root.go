// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mmctl/printer"
)

func Run(args []string) error {
	viper.SetEnvPrefix("mmctl")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	RootCmd.PersistentFlags().String("format", "plain", "the format of the command output [plain, json]")
	_ = viper.BindPFlag("format", RootCmd.PersistentFlags().Lookup("format"))
	RootCmd.PersistentFlags().Bool("strict", false, "will only run commands if the mmctl version matches the server one")
	_ = viper.BindPFlag("strict", RootCmd.PersistentFlags().Lookup("strict"))
	RootCmd.PersistentFlags().Bool("insecure-sha1-intermediate", false, "allows to use insecure TLS protocols, such as SHA-1")
	_ = viper.BindPFlag("insecure-sha1-intermediate", RootCmd.PersistentFlags().Lookup("insecure-sha1-intermediate"))

	RootCmd.SetArgs(args)

	return RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:               "mmctl",
	Short:             "Remote client for the Open Source, self-hosted Slack-alternative",
	Long:              `Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com`,
	DisableAutoGenTag: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		format := viper.GetString("format")
		printer.SetFormat(format)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		printer.Flush()
	},
}

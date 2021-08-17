// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mmctl/printer"
)

func Run(args []string) error {
	viper.SetEnvPrefix("mmctl")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetDefault("local-socket-path", model.LocalModeSocketPath)
	viper.AutomaticEnv()

	RootCmd.PersistentFlags().String("config", filepath.Join(xdgConfigHomeVar, configParent, configFileName), "path to the configuration file")
	_ = viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	RootCmd.PersistentFlags().String("config-path", xdgConfigHomeVar, "path to the configuration directory.")
	_ = viper.BindPFlag("config-path", RootCmd.PersistentFlags().Lookup("config-path"))
	_ = RootCmd.Flags().MarkHidden("config-path")
	RootCmd.PersistentFlags().Bool("suppress-warnings", false, "disables printing warning messages")
	_ = viper.BindPFlag("suppress-warnings", RootCmd.PersistentFlags().Lookup("suppress-warnings"))
	RootCmd.PersistentFlags().String("format", "plain", "the format of the command output [plain, json]")
	_ = viper.BindPFlag("format", RootCmd.PersistentFlags().Lookup("format"))
	RootCmd.PersistentFlags().Bool("strict", false, "will only run commands if the mmctl version matches the server one")
	_ = viper.BindPFlag("strict", RootCmd.PersistentFlags().Lookup("strict"))
	RootCmd.PersistentFlags().Bool("insecure-sha1-intermediate", false, "allows to use insecure TLS protocols, such as SHA-1")
	_ = viper.BindPFlag("insecure-sha1-intermediate", RootCmd.PersistentFlags().Lookup("insecure-sha1-intermediate"))
	RootCmd.PersistentFlags().Bool("insecure-tls-version", false, "allows to use TLS versions 1.0 and 1.1")
	_ = viper.BindPFlag("insecure-tls-version", RootCmd.PersistentFlags().Lookup("insecure-tls-version"))
	RootCmd.PersistentFlags().Bool("local", false, "allows communicating with the server through a unix socket")
	_ = viper.BindPFlag("local", RootCmd.PersistentFlags().Lookup("local"))

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

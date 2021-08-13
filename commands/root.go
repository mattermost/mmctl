// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"
	"runtime/debug"
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

	RootCmd.PersistentFlags().String("config-path", xdgConfigHomeVar, fmt.Sprintf("path to the configuration directory. If \"%s/.%s\" exists it will take precedence over the default value", userHomeVar, configFileName))
	_ = viper.BindPFlag("config-path", RootCmd.PersistentFlags().Lookup("config-path"))
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
	RootCmd.PersistentFlags().Bool("short-stat", false, "short stat will provide useful statistical data")
	_ = RootCmd.PersistentFlags().MarkHidden("short-stat")
	RootCmd.PersistentFlags().Bool("no-stat", false, "the statistical data won't be displayed")
	_ = RootCmd.PersistentFlags().MarkHidden("no-stat")
	RootCmd.PersistentFlags().Bool("disable-pager", false, "disables paged output")
	_ = viper.BindPFlag("disable-pager", RootCmd.PersistentFlags().Lookup("disable-pager"))

	RootCmd.SetArgs(args)

	defer func() {
		if x := recover(); x != nil {
			printer.PrintError("Uh oh! Something unexpected happen :( Would you mind reporting it?\n")
			printer.PrintError(`https://github.com/mattermost/mmctl/issues/new?title=%5Bbug%5D%20panic%20on%20mmctl%20v` + Version + "&body=%3C!---%20Please%20provide%20the%20stack%20trace%20--%3E\n")
			printer.PrintError(string(debug.Stack()))

			os.Exit(1)
		}
	}()

	return RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:               "mmctl",
	Short:             "Remote client for the Open Source, self-hosted Slack-alternative",
	Long:              `Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com`,
	DisableAutoGenTag: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		format := viper.GetString("format")
		if viper.GetBool("disable-pager") {
			printer.OverrideEnablePager(false)
		}

		printer.SetFormat(format)
		printer.SetCommand(cmd)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		_ = printer.Flush()
	},
	SilenceUsage: true,
}

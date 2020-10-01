// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/v5/mlog/human"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Display logs in a human-readable format",
	Long:  "Display logs in a human-readable format. The \"--format\" flag cannot be used here.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("format") || viper.GetString("format") == printer.FormatJSON {
			return errors.New("format cannot be used")
		}
		return nil
	},
	RunE: withClient(logsCmdF),
}

func init() {
	LogsCmd.Flags().IntP("number", "n", 200, "Number of log lines to retrieve.")
	LogsCmd.Flags().BoolP("logrus", "l", false, "Use logrus for formatting.")
	RootCmd.AddCommand(LogsCmd)
}

func logsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	number, _ := cmd.Flags().GetInt("number")
	logLines, response := c.GetLogs(0, number)
	if response.Error != nil {
		return errors.New("Unable to retrieve logs. Error: " + response.Error.Error())
	}

	reader := bytes.NewReader([]byte(strings.Join(logLines, "")))

	var writer human.LogWriter
	if logrus, _ := cmd.Flags().GetBool("logrus"); logrus {
		writer = human.NewLogrusWriter(os.Stdout)
	} else {
		writer = human.NewSimpleWriter(os.Stdout)
	}
	human.ProcessLogs(reader, writer)

	return nil
}

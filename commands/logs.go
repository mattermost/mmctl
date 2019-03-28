package commands

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/mlog/human"
	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Display logs in a human-readable format",
	RunE:  logsCmdF,
}

func init() {
	LogsCmd.Flags().IntP("number", "n", 200, "Number of log lines to retrieve.")
	LogsCmd.Flags().BoolP("logrus", "l", false, "Use logrus for formatting.")
	RootCmd.AddCommand(LogsCmd)
}

func logsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	number, _ := command.Flags().GetInt("number")
	logLines, response := c.GetLogs(0, number)
	if response.Error != nil {
		return errors.New("Unable to retrieve logs. Error: " + response.Error.Error())
	}

	reader := bytes.NewReader([]byte(strings.Join(logLines, "")))

	var writer human.LogWriter
	if logrus, _ := command.Flags().GetBool("logrus"); logrus {
		writer = human.NewLogrusWriter(os.Stdout)
	} else {
		writer = human.NewSimpleWriter(os.Stdout)
	}
	human.ProcessLogs(reader, writer)

	return nil
}

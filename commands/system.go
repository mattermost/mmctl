// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "System management",
	Long:  `System management commands for interacting with the server state and configuration.`,
}

var SystemGetBusyCmd = &cobra.Command{
	Use:     "getbusy",
	Short:   "Get the current busy state",
	Long:    `Gets the server busy state (high load) and timestamp corresponding to when the server busy flag will be automatically cleared.`,
	Example: `  system getbusy`,
	Args:    cobra.NoArgs,
	RunE:    withClient(getBusyCmdF),
}

var SystemSetBusyCmd = &cobra.Command{
	Use:     "setbusy -s [seconds]",
	Short:   "Set the busy state to true",
	Long:    `Set the busy state to true for the specified number of seconds, which disables non-critical services.`,
	Example: `  system setbusy -s 3600`,
	Args:    cobra.NoArgs,
	RunE:    withClient(setBusyCmdF),
}

var SystemClearBusyCmd = &cobra.Command{
	Use:     "clearbusy",
	Short:   "Clears the busy state",
	Long:    `Clear the busy state, which re-enables non-critical services.`,
	Example: `  system clearbusy`,
	Args:    cobra.NoArgs,
	RunE:    withClient(clearBusyCmdF),
}

func init() {
	SystemSetBusyCmd.Flags().UintP("seconds", "s", 3600, "Number of seconds until server is automatically marked as not busy.")
	_ = SystemSetBusyCmd.MarkFlagRequired("seconds")
	SystemCmd.AddCommand(
		SystemGetBusyCmd,
		SystemSetBusyCmd,
		SystemClearBusyCmd,
	)
	RootCmd.AddCommand(SystemCmd)
}

func getBusyCmdF(c client.Client, cmd *cobra.Command, _ []string) error {
	printer.SetSingle(true)

	sbs, response := c.GetServerBusy()
	if response.Error != nil {
		printer.PrintError("Unable to get busy state: " + response.Error.Error())
		return response.Error
	}
	printer.PrintT("busy:{{.Busy}} expires:{{.Expires}}", sbs)
	return nil
}

func setBusyCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	seconds, err := cmd.Flags().GetUint("seconds")
	if err != nil || seconds == 0 {
		err = fmt.Errorf("seconds must be a number > 0")
		printer.PrintError(err.Error())
		return err
	}

	ok, response := c.SetServerBusy(int(seconds))
	if response.Error != nil || !ok {
		printer.PrintError(fmt.Sprintf("Unable to set busy state: %v", response.Error))
		return response.Error
	}
	printer.PrintT("Busy state set", map[string]string{"status": "ok"})
	return nil
}

func clearBusyCmdF(c client.Client, cmd *cobra.Command, _ []string) error {
	ok, response := c.ClearServerBusy()
	if response.Error != nil || !ok {
		printer.PrintError(fmt.Sprintf("Unable to clear busy state: %v", response.Error))
		return response.Error
	}
	printer.PrintT("Busy state cleared", map[string]string{"status": "ok"})
	return nil
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "System management",
}

var SystemGetBusyCmd = &cobra.Command{
	Use:     "getbusy",
	Short:   "Get the current busy state",
	Long:    `Gets the server busy state (high load) and timestamp corresponding to when the server busy flag will be automatically cleared.`,
	Example: `  system getbusy`,
	RunE:    withClient(getBusyCmdF),
}

var SystemSetBusyCmd = &cobra.Command{
	Use:     "setbusy [seconds]",
	Short:   "Set the busy state to true",
	Long:    `Set the busy state to true for the specified number of seconds, which disables non-critical services.`,
	Example: `  system setbusy 3600`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(setBusyCmdF),
}

var SystemClearBusyCmd = &cobra.Command{
	Use:     "clearbusy",
	Short:   "Clears the busy state",
	Long:    `Clear the busy state, which re-enables non-critical services.`,
	Example: `  system clearbusy`,
	RunE:    withClient(clearBusyCmdF),
}

func init() {
	SystemSetBusyCmd.Flags().Uint("seconds", 3600, "Number of seconds until server is automatically marked as not busy.")
	SystemCmd.AddCommand(
		SystemGetBusyCmd,
		SystemSetBusyCmd,
		SystemClearBusyCmd,
	)
	RootCmd.AddCommand(SystemCmd)
}

func getBusyCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	sbs, response := c.GetServerBusy()
	if response.Error != nil {
		return response.Error
	}
	printer.PrintT("busy:{{.Busy}} expires:{{.Expires}}", sbs)
	return nil
}

func setBusyCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	seconds, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		printer.PrintError("Seconds must be a number > 0")
		return err
	}

	ok, response := c.SetServerBusy(int(seconds))
	if response.Error != nil || !ok {
		printer.PrintError("Unable to set busy state")
		return response.Error
	}

	printer.Print("Busy state set")
	return nil
}

func clearBusyCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	ok, response := c.ClearServerBusy()
	if response.Error != nil || !ok {
		printer.PrintError("Unable to clear busy state")
		return response.Error
	}

	printer.Print("Busy state cleared")
	return nil
}

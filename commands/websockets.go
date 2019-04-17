package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var WebsocketCmd = &cobra.Command{
	Use:   "websocket",
	Short: "Display websocket in a human-readable format",
	RunE:  websocketCmdF,
}

func init() {
	RootCmd.AddCommand(WebsocketCmd)
}

func websocketCmdF(command *cobra.Command, args []string) error {
	c, err := InitWebSocketClient()
	if err != nil {
		return err
	}
	appErr := c.Connect()
	if appErr != nil {
		return errors.New(appErr.Error())
	}

	c.Listen()
	fmt.Println("Press CTRL+C to exit")
	for {
		event := <-c.EventChannel
		fmt.Println(event.ToJson())
	}
}

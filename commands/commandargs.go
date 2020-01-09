package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
)

// getCommandFromCommandArg retrieves a Command by command id. Future versions
// may allow lookup by team:trigger
func getCommandFromCommandArg(c client.Client, commandArg string) *model.Command {
	cmd, _ := c.GetCommandById(commandArg)
	return cmd
}

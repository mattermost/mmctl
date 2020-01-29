// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/client"

	"github.com/mattermost/mattermost-server/v5/model"
)

// getCommandFromCommandArg retrieves a Command by command id. Future versions
// may allow lookup by team:trigger
func getCommandFromCommandArg(c client.Client, commandArg string) *model.Command {
	cmd, _ := c.GetCommandById(commandArg)
	return cmd
}

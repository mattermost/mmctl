// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mmctl/client"

	"github.com/mattermost/mattermost-server/v5/model"
)

// getCommandFromCommandArg retrieves a Command by command id or team:trigger.
func getCommandFromCommandArg(c client.Client, commandArg string) *model.Command {
	if checkSlash(commandArg) {
		return nil
	}

	cmd := getCommandFromTeamTrigger(c, commandArg)
	if cmd == nil {
		cmd, _ = c.GetCommandById(commandArg)
	}
	return cmd
}

// getCommandFromTeamTrigger retrieves a Command via team:trigger syntax.
func getCommandFromTeamTrigger(c client.Client, teamTrigger string) *model.Command {
	if !strings.Contains(teamTrigger, ":") {
		return nil
	}

	arr := strings.Split(teamTrigger, ":")
	if len(arr) != 2 {
		return nil
	}

	team, _ := c.GetTeamByName(arr[0], "")
	if team == nil {
		return nil
	}

	trigger := arr[1]
	if len(trigger) == 0 {
		return nil
	}

	var cmd *model.Command
	list, _ := c.ListCommands(team.Id, false)
	if list == nil {
		return nil
	}

	for _, c := range list {
		if c.Trigger == trigger {
			cmd = c
		}
	}
	return cmd
}

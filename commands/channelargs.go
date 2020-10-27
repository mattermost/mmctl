// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
)

const channelArgSeparator = ":"

func getChannelsFromChannelArgs(c client.Client, channelArgs []string) []*model.Channel {
	channels := make([]*model.Channel, 0, len(channelArgs))
	for _, channelArg := range channelArgs {
		channel := getChannelFromChannelArg(c, channelArg)
		channels = append(channels, channel)
	}
	return channels
}

func parseChannelArg(channelArg string) (string, string) {
	result := strings.SplitN(channelArg, channelArgSeparator, 2)
	if len(result) == 1 {
		return "", channelArg
	}
	return result[0], result[1]
}

func getChannelFromChannelArg(c client.Client, channelArg string) *model.Channel {
	teamArg, channelPart := parseChannelArg(channelArg)
	if teamArg == "" && channelPart == "" {
		return nil
	}

	if checkDots(channelPart) || checkSlash(channelPart) {
		return nil
	}

	var channel *model.Channel
	if teamArg != "" {
		team := getTeamFromTeamArg(c, teamArg)
		if team == nil {
			return nil
		}

		channel, _ = c.GetChannelByNameIncludeDeleted(channelPart, team.Id, "")
	}

	if channel == nil {
		channel, _ = c.GetChannel(channelPart, "")
	}

	return channel
}

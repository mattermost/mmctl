package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/model"
)

const CHANNEL_ARG_SEPARATOR = ":"

func getChannelsFromChannelArgs(c *model.Client4, channelArgs []string) []*model.Channel {
	channels := make([]*model.Channel, 0, len(channelArgs))
	for _, channelArg := range channelArgs {
		channel := getChannelFromChannelArg(c, channelArg)
		channels = append(channels, channel)
	}
	return channels
}

func parseChannelArg(channelArg string) (string, string) {
	result := strings.SplitN(channelArg, CHANNEL_ARG_SEPARATOR, 2)
	if len(result) == 1 {
		return "", channelArg
	}
	return result[0], result[1]
}

func getChannelFromChannelArg(c *model.Client4, channelArg string) *model.Channel {
	teamArg, channelPart := parseChannelArg(channelArg)
	if teamArg == "" && channelPart == "" {
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

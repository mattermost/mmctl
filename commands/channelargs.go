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

func getChannelsFromArgs(c client.Client, channelArgs []string) ([]*model.Channel, *FindEntitySummary) {
	var (
		channels []*model.Channel
		errors   []error
	)
	for _, channelArg := range channelArgs {
		channel, err := getChannelFromArg(c, channelArg)
		if err != nil {
			errors = append(errors, err)
		} else {
			channels = append(channels, channel)
		}
	}
	if len(errors) > 0 {
		summary := &FindEntitySummary{
			Errors: errors,
		}
		return channels, summary
	}
	return channels, nil
}

func getChannelFromArg(c client.Client, arg string) (*model.Channel, error) {
	teamArg, channelArg := parseChannelArg(arg)
	if teamArg == "" && channelArg == "" {
		return nil, ErrEntityNotFound{Type: "channel", ID: arg}
	}
	if checkDots(channelArg) || checkSlash(channelArg) {
		return nil, ErrEntityNotFound{Type: "channel", ID: arg}
	}
	var (
		channel  *model.Channel
		response *model.Response
	)
	if teamArg != "" {
		team, err := getTeamFromArg(c, teamArg)
		if err != nil {
			return nil, err
		}
		channel, response = c.GetChannelByNameIncludeDeleted(channelArg, team.Id, "")
		if isErrorSevere(response) {
			return nil, response.Error
		}
	}
	if channel != nil {
		return channel, nil
	}
	channel, response = c.GetChannel(channelArg, "")
	if isErrorSevere(response) {
		return nil, response.Error
	}
	if channel == nil {
		return nil, ErrEntityNotFound{Type: "channel", ID: arg}
	}
	return channel, nil
}

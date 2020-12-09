// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/utils"
)

// CreateDefaultChannels creates channels in the given team for each channel returned by (*App).DefaultChannelNames.
//
func (a *App) CreateDefaultChannels(teamID string) ([]*model.Channel, *model.AppError) {
	displayNames := map[string]string{
		"town-square": utils.T("api.channel.create_default_channels.town_square"),
		"off-topic":   utils.T("api.channel.create_default_channels.off_topic"),
	}
	channels := []*model.Channel{}
	defaultChannelNames := a.DefaultChannelNames()
	for _, name := range defaultChannelNames {
		displayName := utils.TDefault(displayNames[name], name)
		channel := &model.Channel{DisplayName: displayName, Name: name, Type: model.CHANNEL_OPEN, TeamId: teamID}
		if _, err := a.CreateChannel(channel, false); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// DefaultChannelNames returns the list of system-wide default channel names.
//
// By default the list will be (not necessarily in this order):
//	['town-square', 'off-topic']
// However, if TeamSettings.ExperimentalDefaultChannels contains a list of channels then that list will replace
// 'off-topic' and be included in the return results in addition to 'town-square'. For example:
//	['town-square', 'game-of-thrones', 'wow']
//
func (a *App) DefaultChannelNames() []string {
	names := []string{"town-square"}

	if len(a.Config().TeamSettings.ExperimentalDefaultChannels) == 0 {
		names = append(names, "off-topic")
	} else {
		seenChannels := map[string]bool{"town-square": true}
		for _, channelName := range a.Config().TeamSettings.ExperimentalDefaultChannels {
			if !seenChannels[channelName] {
				names = append(names, channelName)
				seenChannels[channelName] = true
			}
		}
	}

	return names
}

func (a *App) JoinDefaultChannels(teamId string, user *model.User, shouldBeAdmin bool, userRequestorId string) *model.AppError {
	var requestor *model.User
	var nErr error
	if userRequestorId != "" {
		requestor, nErr = a.Srv().Store.User().Get(userRequestorId)
		if nErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(nErr, &nfErr):
				return model.NewAppError("JoinDefaultChannels", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
			default:
				return model.NewAppError("JoinDefaultChannels", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
	}

	var err *model.AppError
	for _, channelName := range a.DefaultChannelNames() {
		channel, channelErr := a.Srv().Store.Channel().GetByName(teamId, channelName, true)
		if channelErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(err, &nfErr):
				err = model.NewAppError("JoinDefaultChannels", "app.channel.get_by_name.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
			default:
				err = model.NewAppError("JoinDefaultChannels", "app.channel.get_by_name.existing.app_error", nil, channelErr.Error(), http.StatusInternalServerError)
			}
			continue
		}

		if channel.Type != model.CHANNEL_OPEN {
			continue
		}

		cm := &model.ChannelMember{
			ChannelId:   channel.Id,
			UserId:      user.Id,
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
			SchemeAdmin: shouldBeAdmin,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}

		_, nErr = a.Srv().Store.Channel().SaveMember(cm)
		if histErr := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); histErr != nil {
			mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(histErr))
			return model.NewAppError("JoinDefaultChannels", "app.channel_member_history.log_join_event.internal_error", nil, histErr.Error(), http.StatusInternalServerError)
		}

		if *a.Config().ServiceSettings.ExperimentalEnableDefaultChannelLeaveJoinMessages {
			a.postJoinMessageForDefaultChannel(user, requestor, channel)
		}

		a.invalidateCacheForChannelMembers(channel.Id)

		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ADDED, "", channel.Id, "", nil)
		message.Add("user_id", user.Id)
		message.Add("team_id", channel.TeamId)
		a.Publish(message)

	}

	if nErr != nil {
		var appErr *model.AppError
		var cErr *store.ErrConflict
		switch {
		case errors.As(nErr, &cErr):
			if cErr.Resource == "ChannelMembers" {
				return model.NewAppError("JoinDefaultChannels", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
			}
		case errors.As(nErr, &appErr):
			return appErr
		default:
			return model.NewAppError("JoinDefaultChannels", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	return nil
}

func (a *App) postJoinMessageForDefaultChannel(user *model.User, requestor *model.User, channel *model.Channel) {
	if channel.Name == model.DEFAULT_CHANNEL {
		if requestor == nil {
			if err := a.postJoinTeamMessage(user, channel); err != nil {
				mlog.Error("Failed to post join/leave message", mlog.Err(err))
			}
		} else {
			if err := a.postAddToTeamMessage(requestor, user, channel, ""); err != nil {
				mlog.Error("Failed to post join/leave message", mlog.Err(err))
			}
		}
	} else {
		if requestor == nil {
			if err := a.postJoinChannelMessage(user, channel); err != nil {
				mlog.Error("Failed to post join/leave message", mlog.Err(err))
			}
		} else {
			if err := a.PostAddToChannelMessage(requestor, user, channel, ""); err != nil {
				mlog.Error("Failed to post join/leave message", mlog.Err(err))
			}
		}
	}
}

func (a *App) CreateChannelWithUser(channel *model.Channel, userId string) (*model.Channel, *model.AppError) {
	if channel.IsGroupOrDirect() {
		return nil, model.NewAppError("CreateChannelWithUser", "api.channel.create_channel.direct_channel.app_error", nil, "", http.StatusBadRequest)
	}

	if len(channel.TeamId) == 0 {
		return nil, model.NewAppError("CreateChannelWithUser", "app.channel.create_channel.no_team_id.app_error", nil, "", http.StatusBadRequest)
	}

	// Get total number of channels on current team
	count, err := a.GetNumberOfChannelsOnTeam(channel.TeamId)
	if err != nil {
		return nil, err
	}

	if int64(count+1) > *a.Config().TeamSettings.MaxChannelsPerTeam {
		return nil, model.NewAppError("CreateChannelWithUser", "api.channel.create_channel.max_channel_limit.app_error", map[string]interface{}{"MaxChannelsPerTeam": *a.Config().TeamSettings.MaxChannelsPerTeam}, "", http.StatusBadRequest)
	}

	channel.CreatorId = userId

	rchannel, err := a.CreateChannel(channel, true)
	if err != nil {
		return nil, err
	}

	var user *model.User
	if user, err = a.GetUser(userId); err != nil {
		return nil, err
	}

	a.postJoinChannelMessage(user, channel)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_CREATED, "", "", userId, nil)
	message.Add("channel_id", channel.Id)
	message.Add("team_id", channel.TeamId)
	a.Publish(message)

	return rchannel, nil
}

// RenameChannel is used to rename the channel Name and the DisplayName fields
func (a *App) RenameChannel(channel *model.Channel, newChannelName string, newDisplayName string) (*model.Channel, *model.AppError) {
	if channel.Type == model.CHANNEL_DIRECT {
		return nil, model.NewAppError("RenameChannel", "api.channel.rename_channel.cant_rename_direct_messages.app_error", nil, "", http.StatusBadRequest)
	}

	if channel.Type == model.CHANNEL_GROUP {
		return nil, model.NewAppError("RenameChannel", "api.channel.rename_channel.cant_rename_group_messages.app_error", nil, "", http.StatusBadRequest)
	}

	channel.Name = newChannelName
	if newDisplayName != "" {
		channel.DisplayName = newDisplayName
	}

	newChannel, err := a.UpdateChannel(channel)
	if err != nil {
		return nil, err
	}

	return newChannel, nil
}

func (a *App) CreateChannel(channel *model.Channel, addMember bool) (*model.Channel, *model.AppError) {
	channel.DisplayName = strings.TrimSpace(channel.DisplayName)
	sc, nErr := a.Srv().Store.Channel().Save(channel, *a.Config().TeamSettings.MaxChannelsPerTeam)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.direct_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			return sc, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateChannel", "app.channel.create_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if addMember {
		user, nErr := a.Srv().Store.User().Get(channel.CreatorId)
		if nErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(nErr, &nfErr):
				return nil, model.NewAppError("CreateChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
			default:
				return nil, model.NewAppError("CreateChannel", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}

		cm := &model.ChannelMember{
			ChannelId:   sc.Id,
			UserId:      user.Id,
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
			SchemeAdmin: true,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}

		if _, nErr := a.Srv().Store.Channel().SaveMember(cm); nErr != nil {
			var appErr *model.AppError
			var cErr *store.ErrConflict
			switch {
			case errors.As(nErr, &cErr):
				switch cErr.Resource {
				case "ChannelMembers":
					return nil, model.NewAppError("CreateChannel", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
				}
			case errors.As(nErr, &appErr):
				return nil, appErr
			default:
				return nil, model.NewAppError("CreateChannel", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}

		if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(channel.CreatorId, sc.Id, model.GetMillis()); err != nil {
			mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
			return nil, model.NewAppError("CreateChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}

		a.InvalidateCacheForUser(channel.CreatorId)
	}

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.ChannelHasBeenCreated(pluginContext, sc)
				return true
			}, plugin.ChannelHasBeenCreatedId)
		})
	}

	return sc, nil
}

func (a *App) GetOrCreateDirectChannel(userId, otherUserId string) (*model.Channel, *model.AppError) {
	channel, nErr := a.Srv().Store.Channel().GetByName("", model.GetDMNameFromIds(userId, otherUserId), true)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		if errors.As(nErr, &nfErr) {
			var err *model.AppError
			channel, err = a.createDirectChannel(userId, otherUserId)
			if err != nil {
				if err.Id == store.CHANNEL_EXISTS_ERROR {
					return channel, nil
				}
				return nil, err
			}

			a.WaitForChannelMembership(channel.Id, userId)

			a.InvalidateCacheForUser(userId)
			a.InvalidateCacheForUser(otherUserId)

			if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
				a.Srv().Go(func() {
					pluginContext := a.PluginContext()
					pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
						hooks.ChannelHasBeenCreated(pluginContext, channel)
						return true
					}, plugin.ChannelHasBeenCreatedId)
				})
			}

			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_DIRECT_ADDED, "", channel.Id, "", nil)
			message.Add("teammate_id", otherUserId)
			a.Publish(message)

			return channel, nil
		}
		return nil, model.NewAppError("GetOrCreateDirectChannel", "web.incoming_webhook.channel.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	}
	return channel, nil
}

func (a *App) createDirectChannel(userId string, otherUserId string) (*model.Channel, *model.AppError) {
	uc1 := make(chan store.StoreResult, 1)
	uc2 := make(chan store.StoreResult, 1)
	go func() {
		user, err := a.Srv().Store.User().Get(userId)
		uc1 <- store.StoreResult{Data: user, NErr: err}
		close(uc1)
	}()
	go func() {
		user, err := a.Srv().Store.User().Get(otherUserId)
		uc2 <- store.StoreResult{Data: user, NErr: err}
		close(uc2)
	}()

	result := <-uc1
	if result.NErr != nil {
		return nil, model.NewAppError("CreateDirectChannel", "api.channel.create_direct_channel.invalid_user.app_error", nil, userId, http.StatusBadRequest)
	}
	user := result.Data.(*model.User)

	result = <-uc2
	if result.NErr != nil {
		return nil, model.NewAppError("CreateDirectChannel", "api.channel.create_direct_channel.invalid_user.app_error", nil, otherUserId, http.StatusBadRequest)
	}
	otherUser := result.Data.(*model.User)

	channel, nErr := a.Srv().Store.Channel().CreateDirectChannel(user, otherUser)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_direct_channel.not_direct.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("SqlChannelStore.Save", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			switch cErr.Resource {
			case "Channel":
				return channel, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
			case "ChannelMembers":
				return nil, model.NewAppError("CreateChannel", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
			}
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateDirectChannel", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(userId, channel.Id, model.GetMillis()); err != nil {
		mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
		return nil, model.NewAppError("CreateDirectChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
	}
	if userId != otherUserId {
		if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(otherUserId, channel.Id, model.GetMillis()); err != nil {
			mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
			return nil, model.NewAppError("CreateDirectChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return channel, nil
}

func (a *App) WaitForChannelMembership(channelId string, userId string) {
	if len(a.Config().SqlSettings.DataSourceReplicas) == 0 {
		return
	}

	now := model.GetMillis()

	for model.GetMillis()-now < 12000 {

		time.Sleep(100 * time.Millisecond)

		_, err := a.Srv().Store.Channel().GetMember(channelId, userId)

		// If the membership was found then return
		if err == nil {
			return
		}

		// If we received an error, but it wasn't a missing channel member then return
		var nfErr *store.ErrNotFound
		if !errors.As(err, &nfErr) {
			return
		}
	}

	mlog.Error("WaitForChannelMembership giving up", mlog.String("channel_id", channelId), mlog.String("user_id", userId))
}

func (a *App) CreateGroupChannel(userIds []string, creatorId string) (*model.Channel, *model.AppError) {
	channel, err := a.createGroupChannel(userIds, creatorId)
	if err != nil {
		if err.Id == store.CHANNEL_EXISTS_ERROR {
			return channel, nil
		}
		return nil, err
	}

	for _, userId := range userIds {
		if userId == creatorId {
			a.WaitForChannelMembership(channel.Id, creatorId)
		}

		a.InvalidateCacheForUser(userId)
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_GROUP_ADDED, "", channel.Id, "", nil)
	message.Add("teammate_ids", model.ArrayToJson(userIds))
	a.Publish(message)

	return channel, nil
}

func (a *App) createGroupChannel(userIds []string, creatorId string) (*model.Channel, *model.AppError) {
	if len(userIds) > model.CHANNEL_GROUP_MAX_USERS || len(userIds) < model.CHANNEL_GROUP_MIN_USERS {
		return nil, model.NewAppError("CreateGroupChannel", "api.channel.create_group.bad_size.app_error", nil, "", http.StatusBadRequest)
	}

	users, err := a.Srv().Store.User().GetProfileByIds(userIds, nil, true)
	if err != nil {
		return nil, model.NewAppError("createGroupChannel", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if len(users) != len(userIds) {
		return nil, model.NewAppError("CreateGroupChannel", "api.channel.create_group.bad_user.app_error", nil, "user_ids="+model.ArrayToJson(userIds), http.StatusBadRequest)
	}

	group := &model.Channel{
		Name:        model.GetGroupNameFromUserIds(userIds),
		DisplayName: model.GetGroupDisplayNameFromUsers(users, true),
		Type:        model.CHANNEL_GROUP,
	}

	channel, nErr := a.Srv().Store.Channel().Save(group, *a.Config().TeamSettings.MaxChannelsPerTeam)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.direct_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			return channel, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateChannel", "app.channel.create_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	for _, user := range users {
		cm := &model.ChannelMember{
			UserId:      user.Id,
			ChannelId:   group.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
		}

		if _, nErr = a.Srv().Store.Channel().SaveMember(cm); nErr != nil {
			var appErr *model.AppError
			var cErr *store.ErrConflict
			switch {
			case errors.As(nErr, &cErr):
				switch cErr.Resource {
				case "ChannelMembers":
					return nil, model.NewAppError("createGroupChannel", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
				}
			case errors.As(nErr, &appErr):
				return nil, appErr
			default:
				return nil, model.NewAppError("createGroupChannel", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
		if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); err != nil {
			mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
			return nil, model.NewAppError("createGroupChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return channel, nil
}

func (a *App) GetGroupChannel(userIds []string) (*model.Channel, *model.AppError) {
	if len(userIds) > model.CHANNEL_GROUP_MAX_USERS || len(userIds) < model.CHANNEL_GROUP_MIN_USERS {
		return nil, model.NewAppError("GetGroupChannel", "api.channel.create_group.bad_size.app_error", nil, "", http.StatusBadRequest)
	}

	users, err := a.Srv().Store.User().GetProfileByIds(userIds, nil, true)
	if err != nil {
		return nil, model.NewAppError("GetGroupChannel", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if len(users) != len(userIds) {
		return nil, model.NewAppError("GetGroupChannel", "api.channel.create_group.bad_user.app_error", nil, "user_ids="+model.ArrayToJson(userIds), http.StatusBadRequest)
	}

	channel, appErr := a.GetChannelByName(model.GetGroupNameFromUserIds(userIds), "", true)
	if appErr != nil {
		return nil, appErr
	}

	return channel, nil
}

// UpdateChannel updates a given channel by its Id. It also publishes the CHANNEL_UPDATED event.
func (a *App) UpdateChannel(channel *model.Channel) (*model.Channel, *model.AppError) {
	_, err := a.Srv().Store.Channel().Update(channel)
	if err != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("UpdateChannel", "app.channel.update.bad_id", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("UpdateChannel", "app.channel.update_channel.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	a.invalidateCacheForChannel(channel)

	messageWs := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_UPDATED, "", channel.Id, "", nil)
	messageWs.Add("channel", channel.ToJson())
	a.Publish(messageWs)

	return channel, nil
}

// CreateChannelScheme creates a new Scheme of scope channel and assigns it to the channel.
func (a *App) CreateChannelScheme(channel *model.Channel) (*model.Scheme, *model.AppError) {
	scheme, err := a.CreateScheme(&model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	})
	if err != nil {
		return nil, err
	}

	channel.SchemeId = &scheme.Id
	if _, err := a.UpdateChannelScheme(channel); err != nil {
		return nil, err
	}
	return scheme, nil
}

// DeleteChannelScheme deletes a channels scheme and sets its SchemeId to nil.
func (a *App) DeleteChannelScheme(channel *model.Channel) (*model.Channel, *model.AppError) {
	if channel.SchemeId != nil && len(*channel.SchemeId) != 0 {
		if _, err := a.DeleteScheme(*channel.SchemeId); err != nil {
			return nil, err
		}
	}
	channel.SchemeId = nil
	return a.UpdateChannelScheme(channel)
}

// UpdateChannelScheme saves the new SchemeId of the channel passed.
func (a *App) UpdateChannelScheme(channel *model.Channel) (*model.Channel, *model.AppError) {
	var oldChannel *model.Channel
	var err *model.AppError
	if oldChannel, err = a.GetChannel(channel.Id); err != nil {
		return nil, err
	}

	oldChannel.SchemeId = channel.SchemeId
	return a.UpdateChannel(oldChannel)
}

func (a *App) UpdateChannelPrivacy(oldChannel *model.Channel, user *model.User) (*model.Channel, *model.AppError) {
	channel, err := a.UpdateChannel(oldChannel)
	if err != nil {
		return channel, err
	}

	if err := a.postChannelPrivacyMessage(user, channel); err != nil {
		if channel.Type == model.CHANNEL_OPEN {
			channel.Type = model.CHANNEL_PRIVATE
		} else {
			channel.Type = model.CHANNEL_OPEN
		}
		// revert to previous channel privacy
		a.UpdateChannel(channel)
		return channel, err
	}

	a.invalidateCacheForChannel(channel)

	messageWs := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_CONVERTED, channel.TeamId, "", "", nil)
	messageWs.Add("channel_id", channel.Id)
	a.Publish(messageWs)

	return channel, nil
}

func (a *App) postChannelPrivacyMessage(user *model.User, channel *model.Channel) *model.AppError {
	message := (map[string]string{
		model.CHANNEL_OPEN:    utils.T("api.channel.change_channel_privacy.private_to_public"),
		model.CHANNEL_PRIVATE: utils.T("api.channel.change_channel_privacy.public_to_private"),
	})[channel.Type]
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      model.POST_CHANGE_CHANNEL_PRIVACY,
		UserId:    user.Id,
		Props: model.StringInterface{
			"username": user.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postChannelPrivacyMessage", "api.channel.post_channel_privacy_message.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) RestoreChannel(channel *model.Channel, userId string) (*model.Channel, *model.AppError) {
	if channel.DeleteAt == 0 {
		return nil, model.NewAppError("restoreChannel", "api.channel.restore_channel.restored.app_error", nil, "", http.StatusBadRequest)
	}

	if err := a.Srv().Store.Channel().Restore(channel.Id, model.GetMillis()); err != nil {
		return nil, model.NewAppError("RestoreChannel", "app.channel.restore.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	channel.DeleteAt = 0
	a.invalidateCacheForChannel(channel)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_RESTORED, channel.TeamId, "", "", nil)
	message.Add("channel_id", channel.Id)
	a.Publish(message)

	user, nErr := a.Srv().Store.User().Get(userId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("RestoreChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("RestoreChannel", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if user != nil {
		T := utils.GetUserTranslations(user.Locale)

		post := &model.Post{
			ChannelId: channel.Id,
			Message:   T("api.channel.restore_channel.unarchived", map[string]interface{}{"Username": user.Username}),
			Type:      model.POST_CHANNEL_RESTORED,
			UserId:    userId,
			Props: model.StringInterface{
				"username": user.Username,
			},
		}

		if _, err := a.CreatePost(post, channel, false, true); err != nil {
			mlog.Error("Failed to post unarchive message", mlog.Err(err))
		}
	}

	return channel, nil
}

func (a *App) PatchChannel(channel *model.Channel, patch *model.ChannelPatch, userId string) (*model.Channel, *model.AppError) {
	oldChannelDisplayName := channel.DisplayName
	oldChannelHeader := channel.Header
	oldChannelPurpose := channel.Purpose

	channel.Patch(patch)
	channel, err := a.UpdateChannel(channel)
	if err != nil {
		return nil, err
	}

	if oldChannelDisplayName != channel.DisplayName {
		if err = a.PostUpdateChannelDisplayNameMessage(userId, channel, oldChannelDisplayName, channel.DisplayName); err != nil {
			mlog.Error(err.Error())
		}
	}

	if channel.Header != oldChannelHeader {
		if err = a.PostUpdateChannelHeaderMessage(userId, channel, oldChannelHeader, channel.Header); err != nil {
			mlog.Error(err.Error())
		}
	}

	if channel.Purpose != oldChannelPurpose {
		if err = a.PostUpdateChannelPurposeMessage(userId, channel, oldChannelPurpose, channel.Purpose); err != nil {
			mlog.Error(err.Error())
		}
	}

	return channel, nil
}

// GetSchemeRolesForChannel Checks if a channel or its team has an override scheme for channel roles and returns the scheme roles or default channel roles.
func (a *App) GetSchemeRolesForChannel(channelId string) (guestRoleName, userRoleName, adminRoleName string, err *model.AppError) {
	channel, err := a.GetChannel(channelId)
	if err != nil {
		return
	}

	if channel.SchemeId != nil && len(*channel.SchemeId) != 0 {
		var scheme *model.Scheme
		scheme, err = a.GetScheme(*channel.SchemeId)
		if err != nil {
			return
		}

		guestRoleName = scheme.DefaultChannelGuestRole
		userRoleName = scheme.DefaultChannelUserRole
		adminRoleName = scheme.DefaultChannelAdminRole

		return
	}

	return a.GetTeamSchemeChannelRoles(channel.TeamId)
}

// GetTeamSchemeChannelRoles Checks if a team has an override scheme and returns the scheme channel role names or default channel role names.
func (a *App) GetTeamSchemeChannelRoles(teamId string) (guestRoleName, userRoleName, adminRoleName string, err *model.AppError) {
	team, err := a.GetTeam(teamId)
	if err != nil {
		return
	}

	if team.SchemeId != nil && len(*team.SchemeId) != 0 {
		var scheme *model.Scheme
		scheme, err = a.GetScheme(*team.SchemeId)
		if err != nil {
			return
		}

		guestRoleName = scheme.DefaultChannelGuestRole
		userRoleName = scheme.DefaultChannelUserRole
		adminRoleName = scheme.DefaultChannelAdminRole
	} else {
		guestRoleName = model.CHANNEL_GUEST_ROLE_ID
		userRoleName = model.CHANNEL_USER_ROLE_ID
		adminRoleName = model.CHANNEL_ADMIN_ROLE_ID
	}

	return
}

// GetChannelModerationsForChannel Gets a channels ChannelModerations from either the higherScoped roles or from the channel scheme roles.
func (a *App) GetChannelModerationsForChannel(channel *model.Channel) ([]*model.ChannelModeration, *model.AppError) {
	guestRoleName, memberRoleName, _, err := a.GetSchemeRolesForChannel(channel.Id)
	if err != nil {
		return nil, err
	}

	memberRole, err := a.GetRoleByName(memberRoleName)
	if err != nil {
		return nil, err
	}

	var guestRole *model.Role
	if len(guestRoleName) > 0 {
		guestRole, err = a.GetRoleByName(guestRoleName)
		if err != nil {
			return nil, err
		}
	}

	higherScopedGuestRoleName, higherScopedMemberRoleName, _, err := a.GetTeamSchemeChannelRoles(channel.TeamId)
	if err != nil {
		return nil, err
	}
	higherScopedMemberRole, err := a.GetRoleByName(higherScopedMemberRoleName)
	if err != nil {
		return nil, err
	}

	var higherScopedGuestRole *model.Role
	if len(higherScopedGuestRoleName) > 0 {
		higherScopedGuestRole, err = a.GetRoleByName(higherScopedGuestRoleName)
		if err != nil {
			return nil, err
		}
	}

	return buildChannelModerations(channel.Type, memberRole, guestRole, higherScopedMemberRole, higherScopedGuestRole), nil
}

// PatchChannelModerationsForChannel Updates a channels scheme roles based on a given ChannelModerationPatch, if the permissions match the higher scoped role the scheme is deleted.
func (a *App) PatchChannelModerationsForChannel(channel *model.Channel, channelModerationsPatch []*model.ChannelModerationPatch) ([]*model.ChannelModeration, *model.AppError) {
	higherScopedGuestRoleName, higherScopedMemberRoleName, _, err := a.GetTeamSchemeChannelRoles(channel.TeamId)
	if err != nil {
		return nil, err
	}

	higherScopedMemberRole, err := a.GetRoleByName(higherScopedMemberRoleName)
	if err != nil {
		return nil, err
	}

	var higherScopedGuestRole *model.Role
	if len(higherScopedGuestRoleName) > 0 {
		higherScopedGuestRole, err = a.GetRoleByName(higherScopedGuestRoleName)
		if err != nil {
			return nil, err
		}
	}

	higherScopedMemberPermissions := higherScopedMemberRole.GetChannelModeratedPermissions(channel.Type)

	var higherScopedGuestPermissions map[string]bool
	if higherScopedGuestRole != nil {
		higherScopedGuestPermissions = higherScopedGuestRole.GetChannelModeratedPermissions(channel.Type)
	}

	for _, moderationPatch := range channelModerationsPatch {
		if moderationPatch.Roles.Members != nil && *moderationPatch.Roles.Members && !higherScopedMemberPermissions[*moderationPatch.Name] {
			return nil, &model.AppError{Message: "Cannot add a permission that is restricted by the team or system permission scheme"}
		}
		if moderationPatch.Roles.Guests != nil && *moderationPatch.Roles.Guests && !higherScopedGuestPermissions[*moderationPatch.Name] {
			return nil, &model.AppError{Message: "Cannot add a permission that is restricted by the team or system permission scheme"}
		}
	}

	var scheme *model.Scheme
	// Channel has no scheme so create one
	if channel.SchemeId == nil || len(*channel.SchemeId) == 0 {
		scheme, err = a.CreateChannelScheme(channel)
		if err != nil {
			return nil, err
		}

		// Send a websocket event about this new role. The other new roles—member and guest—get emitted when they're updated.
		var adminRole *model.Role
		adminRole, err = a.GetRoleByName(scheme.DefaultChannelAdminRole)
		if err != nil {
			return nil, err
		}
		a.sendUpdatedRoleEvent(adminRole)

		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_SCHEME_UPDATED, "", channel.Id, "", nil)
		a.Publish(message)
		mlog.Info("Permission scheme created.", mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
	} else {
		scheme, err = a.GetScheme(*channel.SchemeId)
		if err != nil {
			return nil, err
		}
	}

	guestRoleName := scheme.DefaultChannelGuestRole
	memberRoleName := scheme.DefaultChannelUserRole
	memberRole, err := a.GetRoleByName(memberRoleName)
	if err != nil {
		return nil, err
	}

	var guestRole *model.Role
	if len(guestRoleName) > 0 {
		guestRole, err = a.GetRoleByName(guestRoleName)
		if err != nil {
			return nil, err
		}
	}

	memberRolePatch := memberRole.RolePatchFromChannelModerationsPatch(channelModerationsPatch, "members")
	var guestRolePatch *model.RolePatch
	if guestRole != nil {
		guestRolePatch = guestRole.RolePatchFromChannelModerationsPatch(channelModerationsPatch, "guests")
	}

	for _, channelModerationPatch := range channelModerationsPatch {
		permissionModified := *channelModerationPatch.Name
		if channelModerationPatch.Roles.Guests != nil && utils.StringInSlice(permissionModified, model.ChannelModeratedPermissionsChangedByPatch(guestRole, guestRolePatch)) {
			if *channelModerationPatch.Roles.Guests {
				mlog.Info("Permission enabled for guests.", mlog.String("permission", permissionModified), mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
			} else {
				mlog.Info("Permission disabled for guests.", mlog.String("permission", permissionModified), mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
			}
		}

		if channelModerationPatch.Roles.Members != nil && utils.StringInSlice(permissionModified, model.ChannelModeratedPermissionsChangedByPatch(memberRole, memberRolePatch)) {
			if *channelModerationPatch.Roles.Members {
				mlog.Info("Permission enabled for members.", mlog.String("permission", permissionModified), mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
			} else {
				mlog.Info("Permission disabled for members.", mlog.String("permission", permissionModified), mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
			}
		}
	}

	memberRolePermissionsUnmodified := len(model.ChannelModeratedPermissionsChangedByPatch(higherScopedMemberRole, memberRolePatch)) == 0
	guestRolePermissionsUnmodified := len(model.ChannelModeratedPermissionsChangedByPatch(higherScopedGuestRole, guestRolePatch)) == 0
	if memberRolePermissionsUnmodified && guestRolePermissionsUnmodified {
		// The channel scheme matches the permissions of its higherScoped scheme so delete the scheme
		if _, err = a.DeleteChannelScheme(channel); err != nil {
			return nil, err
		}

		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_SCHEME_UPDATED, "", channel.Id, "", nil)
		a.Publish(message)

		memberRole = higherScopedMemberRole
		guestRole = higherScopedGuestRole
		mlog.Info("Permission scheme deleted.", mlog.String("channel_id", channel.Id), mlog.String("channel_name", channel.Name))
	} else {
		memberRole, err = a.PatchRole(memberRole, memberRolePatch)
		if err != nil {
			return nil, err
		}
		guestRole, err = a.PatchRole(guestRole, guestRolePatch)
		if err != nil {
			return nil, err
		}
	}

	return buildChannelModerations(channel.Type, memberRole, guestRole, higherScopedMemberRole, higherScopedGuestRole), nil
}

func buildChannelModerations(channelType string, memberRole *model.Role, guestRole *model.Role, higherScopedMemberRole *model.Role, higherScopedGuestRole *model.Role) []*model.ChannelModeration {
	var memberPermissions, guestPermissions, higherScopedMemberPermissions, higherScopedGuestPermissions map[string]bool
	if memberRole != nil {
		memberPermissions = memberRole.GetChannelModeratedPermissions(channelType)
	}
	if guestRole != nil {
		guestPermissions = guestRole.GetChannelModeratedPermissions(channelType)
	}
	if higherScopedMemberRole != nil {
		higherScopedMemberPermissions = higherScopedMemberRole.GetChannelModeratedPermissions(channelType)
	}
	if higherScopedGuestRole != nil {
		higherScopedGuestPermissions = higherScopedGuestRole.GetChannelModeratedPermissions(channelType)
	}

	var channelModerations []*model.ChannelModeration
	for _, permissionKey := range model.ChannelModeratedPermissions {
		roles := &model.ChannelModeratedRoles{}

		roles.Members = &model.ChannelModeratedRole{
			Value:   memberPermissions[permissionKey],
			Enabled: higherScopedMemberPermissions[permissionKey],
		}

		if permissionKey == "manage_members" {
			roles.Guests = nil
		} else {
			roles.Guests = &model.ChannelModeratedRole{
				Value:   guestPermissions[permissionKey],
				Enabled: higherScopedGuestPermissions[permissionKey],
			}
		}

		moderation := &model.ChannelModeration{
			Name:  permissionKey,
			Roles: roles,
		}

		channelModerations = append(channelModerations, moderation)
	}

	return channelModerations
}

func (a *App) UpdateChannelMemberRoles(channelId string, userId string, newRoles string) (*model.ChannelMember, *model.AppError) {
	var member *model.ChannelMember
	var err *model.AppError
	if member, err = a.GetChannelMember(channelId, userId); err != nil {
		return nil, err
	}

	schemeGuestRole, schemeUserRole, schemeAdminRole, err := a.GetSchemeRolesForChannel(channelId)
	if err != nil {
		return nil, err
	}

	prevSchemeGuestValue := member.SchemeGuest

	var newExplicitRoles []string
	member.SchemeGuest = false
	member.SchemeUser = false
	member.SchemeAdmin = false

	for _, roleName := range strings.Fields(newRoles) {
		var role *model.Role
		role, err = a.GetRoleByName(roleName)
		if err != nil {
			err.StatusCode = http.StatusBadRequest
			return nil, err
		}

		if !role.SchemeManaged {
			// The role is not scheme-managed, so it's OK to apply it to the explicit roles field.
			newExplicitRoles = append(newExplicitRoles, roleName)
		} else {
			// The role is scheme-managed, so need to check if it is part of the scheme for this channel or not.
			switch roleName {
			case schemeAdminRole:
				member.SchemeAdmin = true
			case schemeUserRole:
				member.SchemeUser = true
			case schemeGuestRole:
				member.SchemeGuest = true
			default:
				// If not part of the scheme for this channel, then it is not allowed to apply it as an explicit role.
				return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.scheme_role.app_error", nil, "role_name="+roleName, http.StatusBadRequest)
			}
		}
	}

	if member.SchemeUser && member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	if prevSchemeGuestValue != member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.changing_guest_role.app_error", nil, "", http.StatusBadRequest)
	}

	member.ExplicitRoles = strings.Join(newExplicitRoles, " ")

	return a.updateChannelMember(member)
}

func (a *App) UpdateChannelMemberSchemeRoles(channelId string, userId string, isSchemeGuest bool, isSchemeUser bool, isSchemeAdmin bool) (*model.ChannelMember, *model.AppError) {
	member, err := a.GetChannelMember(channelId, userId)
	if err != nil {
		return nil, err
	}

	member.SchemeAdmin = isSchemeAdmin
	member.SchemeUser = isSchemeUser
	member.SchemeGuest = isSchemeGuest

	if member.SchemeUser && member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberSchemeRoles", "api.channel.update_channel_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	// If the migration is not completed, we also need to check the default channel_admin/channel_user roles are not present in the roles field.
	if err = a.IsPhase2MigrationCompleted(); err != nil {
		member.ExplicitRoles = RemoveRoles([]string{model.CHANNEL_GUEST_ROLE_ID, model.CHANNEL_USER_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID}, member.ExplicitRoles)
	}

	return a.updateChannelMember(member)
}

func (a *App) UpdateChannelMemberNotifyProps(data map[string]string, channelId string, userId string) (*model.ChannelMember, *model.AppError) {
	var member *model.ChannelMember
	var err *model.AppError
	if member, err = a.GetChannelMember(channelId, userId); err != nil {
		return nil, err
	}

	// update whichever notify properties have been provided, but don't change the others
	if markUnread, exists := data[model.MARK_UNREAD_NOTIFY_PROP]; exists {
		member.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP] = markUnread
	}

	if desktop, exists := data[model.DESKTOP_NOTIFY_PROP]; exists {
		member.NotifyProps[model.DESKTOP_NOTIFY_PROP] = desktop
	}

	if email, exists := data[model.EMAIL_NOTIFY_PROP]; exists {
		member.NotifyProps[model.EMAIL_NOTIFY_PROP] = email
	}

	if push, exists := data[model.PUSH_NOTIFY_PROP]; exists {
		member.NotifyProps[model.PUSH_NOTIFY_PROP] = push
	}

	if ignoreChannelMentions, exists := data[model.IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP]; exists {
		member.NotifyProps[model.IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP] = ignoreChannelMentions
	}

	member, err = a.updateChannelMember(member)
	if err != nil {
		return nil, err
	}

	a.invalidateCacheForChannelMembersNotifyProps(member.ChannelId)

	return member, nil
}

func (a *App) updateChannelMember(member *model.ChannelMember) (*model.ChannelMember, *model.AppError) {
	member, nErr := a.Srv().Store.Channel().UpdateMember(member)
	if nErr != nil {
		var appErr *model.AppError
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("updateChannelMember", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("updateChannelMember", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	a.InvalidateCacheForUser(member.UserId)

	// Notify the clients that the member notify props changed
	evt := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED, "", "", member.UserId, nil)
	evt.Add("channelMember", member.ToJson())
	a.Publish(evt)

	return member, nil
}

func (a *App) DeleteChannel(channel *model.Channel, userId string) *model.AppError {
	ihc := make(chan store.StoreResult, 1)
	ohc := make(chan store.StoreResult, 1)

	go func() {
		webhooks, err := a.Srv().Store.Webhook().GetIncomingByChannel(channel.Id)
		ihc <- store.StoreResult{Data: webhooks, NErr: err}
		close(ihc)
	}()

	go func() {
		outgoingHooks, err := a.Srv().Store.Webhook().GetOutgoingByChannel(channel.Id, -1, -1)
		ohc <- store.StoreResult{Data: outgoingHooks, NErr: err}
		close(ohc)
	}()

	var user *model.User
	if userId != "" {
		var nErr error
		user, nErr = a.Srv().Store.User().Get(userId)
		if nErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(nErr, &nfErr):
				return model.NewAppError("DeleteChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
			default:
				return model.NewAppError("DeleteChannel", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
	}

	ihcresult := <-ihc
	if ihcresult.NErr != nil {
		return model.NewAppError("DeleteChannel", "app.webhooks.get_incoming_by_channel.app_error", nil, ihcresult.NErr.Error(), http.StatusInternalServerError)
	}

	ohcresult := <-ohc
	if ohcresult.NErr != nil {
		return model.NewAppError("DeleteChannel", "app.webhooks.get_outgoing_by_channel.app_error", nil, ohcresult.NErr.Error(), http.StatusInternalServerError)
	}

	incomingHooks := ihcresult.Data.([]*model.IncomingWebhook)
	outgoingHooks := ohcresult.Data.([]*model.OutgoingWebhook)

	if channel.DeleteAt > 0 {
		err := model.NewAppError("deleteChannel", "api.channel.delete_channel.deleted.app_error", nil, "", http.StatusBadRequest)
		return err
	}

	if channel.Name == model.DEFAULT_CHANNEL {
		err := model.NewAppError("deleteChannel", "api.channel.delete_channel.cannot.app_error", map[string]interface{}{"Channel": model.DEFAULT_CHANNEL}, "", http.StatusBadRequest)
		return err
	}

	if user != nil {
		T := utils.GetUserTranslations(user.Locale)

		post := &model.Post{
			ChannelId: channel.Id,
			Message:   fmt.Sprintf(T("api.channel.delete_channel.archived"), user.Username),
			Type:      model.POST_CHANNEL_DELETED,
			UserId:    userId,
			Props: model.StringInterface{
				"username": user.Username,
			},
		}

		if _, err := a.CreatePost(post, channel, false, true); err != nil {
			mlog.Error("Failed to post archive message", mlog.Err(err))
		}
	}

	now := model.GetMillis()
	for _, hook := range incomingHooks {
		if err := a.Srv().Store.Webhook().DeleteIncoming(hook.Id, now); err != nil {
			mlog.Error("Encountered error deleting incoming webhook", mlog.String("hook_id", hook.Id), mlog.Err(err))
		}
		a.invalidateCacheForWebhook(hook.Id)
	}

	for _, hook := range outgoingHooks {
		if err := a.Srv().Store.Webhook().DeleteOutgoing(hook.Id, now); err != nil {
			mlog.Error("Encountered error deleting outgoing webhook", mlog.String("hook_id", hook.Id), mlog.Err(err))
		}
	}

	deleteAt := model.GetMillis()

	if err := a.Srv().Store.Channel().Delete(channel.Id, deleteAt); err != nil {
		return model.NewAppError("DeleteChannel", "app.channel.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	a.invalidateCacheForChannel(channel)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_DELETED, channel.TeamId, "", "", nil)
	message.Add("channel_id", channel.Id)
	message.Add("delete_at", deleteAt)
	a.Publish(message)

	return nil
}

func (a *App) addUserToChannel(user *model.User, channel *model.Channel, teamMember *model.TeamMember) (*model.ChannelMember, *model.AppError) {
	if channel.Type != model.CHANNEL_OPEN && channel.Type != model.CHANNEL_PRIVATE {
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user_to_channel.type.app_error", nil, "", http.StatusBadRequest)
	}

	channelMember, nErr := a.Srv().Store.Channel().GetMember(channel.Id, user.Id)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		if !errors.As(nErr, &nfErr) {
			return nil, model.NewAppError("AddUserToChannel", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	} else {
		return channelMember, nil
	}

	if channel.IsGroupConstrained() {
		nonMembers, err := a.FilterNonGroupChannelMembers([]string{user.Id}, channel)
		if err != nil {
			return nil, model.NewAppError("addUserToChannel", "api.channel.add_user_to_channel.type.app_error", nil, "", http.StatusInternalServerError)
		}
		if len(nonMembers) > 0 {
			return nil, model.NewAppError("addUserToChannel", "api.channel.add_members.user_denied", map[string]interface{}{"UserIDs": nonMembers}, "", http.StatusBadRequest)
		}
	}

	newMember := &model.ChannelMember{
		ChannelId:   channel.Id,
		UserId:      user.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
		SchemeGuest: user.IsGuest(),
		SchemeUser:  !user.IsGuest(),
	}

	if !user.IsGuest() {
		var userShouldBeAdmin bool
		userShouldBeAdmin, appErr := a.UserIsInAdminRoleGroup(user.Id, channel.Id, model.GroupSyncableTypeChannel)
		if appErr != nil {
			return nil, appErr
		}
		newMember.SchemeAdmin = userShouldBeAdmin
	}

	newMember, nErr = a.Srv().Store.Channel().SaveMember(newMember)
	if nErr != nil {
		mlog.Error("Failed to add member", mlog.String("user_id", user.Id), mlog.String("channel_id", channel.Id), mlog.Err(nErr))
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user.to.channel.failed.app_error", nil, "", http.StatusInternalServerError)
	}
	a.WaitForChannelMembership(channel.Id, user.Id)

	if nErr := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); nErr != nil {
		mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(nErr))
		return nil, model.NewAppError("AddUserToChannel", "app.channel_member_history.log_join_event.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	a.InvalidateCacheForUser(user.Id)
	a.invalidateCacheForChannelMembers(channel.Id)

	return newMember, nil
}

func (a *App) AddUserToChannel(user *model.User, channel *model.Channel) (*model.ChannelMember, *model.AppError) {
	teamMember, nErr := a.Srv().Store.Team().GetMember(channel.TeamId, user.Id)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("AddUserToChannel", "app.team.get_member.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("AddUserToChannel", "app.team.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if teamMember.DeleteAt > 0 {
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user.to.channel.failed.deleted.app_error", nil, "", http.StatusBadRequest)
	}

	newMember, err := a.addUserToChannel(user, channel, teamMember)
	if err != nil {
		return nil, err
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ADDED, "", channel.Id, "", nil)
	message.Add("user_id", user.Id)
	message.Add("team_id", channel.TeamId)
	a.Publish(message)

	return newMember, nil
}

func (a *App) AddChannelMember(userId string, channel *model.Channel, userRequestorId string, postRootId string) (*model.ChannelMember, *model.AppError) {
	if member, err := a.Srv().Store.Channel().GetMember(channel.Id, userId); err != nil {
		var nfErr *store.ErrNotFound
		if !errors.As(err, &nfErr) {
			return nil, model.NewAppError("AddChannelMember", "app.channel.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	} else {
		return member, nil
	}

	var user *model.User
	var err *model.AppError

	if user, err = a.GetUser(userId); err != nil {
		return nil, err
	}

	var userRequestor *model.User
	if userRequestorId != "" {
		if userRequestor, err = a.GetUser(userRequestorId); err != nil {
			return nil, err
		}
	}

	cm, err := a.AddUserToChannel(user, channel)
	if err != nil {
		return nil, err
	}

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.UserHasJoinedChannel(pluginContext, cm, userRequestor)
				return true
			}, plugin.UserHasJoinedChannelId)
		})
	}

	if userRequestorId == "" || userId == userRequestorId {
		a.postJoinChannelMessage(user, channel)
	} else {
		a.Srv().Go(func() {
			a.PostAddToChannelMessage(userRequestor, user, channel, postRootId)
		})
	}

	return cm, nil
}

func (a *App) AddDirectChannels(teamId string, user *model.User) *model.AppError {
	var profiles []*model.User
	options := &model.UserGetOptions{InTeamId: teamId, Page: 0, PerPage: 100}
	profiles, err := a.Srv().Store.User().GetProfiles(options)
	if err != nil {
		return model.NewAppError("AddDirectChannels", "api.user.add_direct_channels_and_forget.failed.error", map[string]interface{}{"UserId": user.Id, "TeamId": teamId, "Error": err.Error()}, "", http.StatusInternalServerError)
	}

	var preferences model.Preferences

	for _, profile := range profiles {
		if profile.Id == user.Id {
			continue
		}

		preference := model.Preference{
			UserId:   user.Id,
			Category: model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW,
			Name:     profile.Id,
			Value:    "true",
		}

		preferences = append(preferences, preference)

		if len(preferences) >= 10 {
			break
		}
	}

	if err := a.Srv().Store.Preference().Save(&preferences); err != nil {
		return model.NewAppError("AddDirectChannels", "api.user.add_direct_channels_and_forget.failed.error", map[string]interface{}{"UserId": user.Id, "TeamId": teamId, "Error": err.Error()}, "", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) PostUpdateChannelHeaderMessage(userId string, channel *model.Channel, oldChannelHeader, newChannelHeader string) *model.AppError {
	user, err := a.Srv().Store.User().Get(userId)
	if err != nil {
		return model.NewAppError("PostUpdateChannelHeaderMessage", "api.channel.post_update_channel_header_message_and_forget.retrieve_user.error", nil, err.Error(), http.StatusBadRequest)
	}

	var message string
	if oldChannelHeader == "" {
		message = fmt.Sprintf(utils.T("api.channel.post_update_channel_header_message_and_forget.updated_to"), user.Username, newChannelHeader)
	} else if newChannelHeader == "" {
		message = fmt.Sprintf(utils.T("api.channel.post_update_channel_header_message_and_forget.removed"), user.Username, oldChannelHeader)
	} else {
		message = fmt.Sprintf(utils.T("api.channel.post_update_channel_header_message_and_forget.updated_from"), user.Username, oldChannelHeader, newChannelHeader)
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      model.POST_HEADER_CHANGE,
		UserId:    userId,
		Props: model.StringInterface{
			"username":   user.Username,
			"old_header": oldChannelHeader,
			"new_header": newChannelHeader,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("", "api.channel.post_update_channel_header_message_and_forget.post.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) PostUpdateChannelPurposeMessage(userId string, channel *model.Channel, oldChannelPurpose string, newChannelPurpose string) *model.AppError {
	user, err := a.Srv().Store.User().Get(userId)
	if err != nil {
		return model.NewAppError("PostUpdateChannelPurposeMessage", "app.channel.post_update_channel_purpose_message.retrieve_user.error", nil, err.Error(), http.StatusBadRequest)
	}

	var message string
	if oldChannelPurpose == "" {
		message = fmt.Sprintf(utils.T("app.channel.post_update_channel_purpose_message.updated_to"), user.Username, newChannelPurpose)
	} else if newChannelPurpose == "" {
		message = fmt.Sprintf(utils.T("app.channel.post_update_channel_purpose_message.removed"), user.Username, oldChannelPurpose)
	} else {
		message = fmt.Sprintf(utils.T("app.channel.post_update_channel_purpose_message.updated_from"), user.Username, oldChannelPurpose, newChannelPurpose)
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      model.POST_PURPOSE_CHANGE,
		UserId:    userId,
		Props: model.StringInterface{
			"username":    user.Username,
			"old_purpose": oldChannelPurpose,
			"new_purpose": newChannelPurpose,
		},
	}
	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("", "app.channel.post_update_channel_purpose_message.post.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) PostUpdateChannelDisplayNameMessage(userId string, channel *model.Channel, oldChannelDisplayName, newChannelDisplayName string) *model.AppError {
	user, err := a.Srv().Store.User().Get(userId)
	if err != nil {
		return model.NewAppError("PostUpdateChannelDisplayNameMessage", "api.channel.post_update_channel_displayname_message_and_forget.retrieve_user.error", nil, err.Error(), http.StatusBadRequest)
	}

	message := fmt.Sprintf(utils.T("api.channel.post_update_channel_displayname_message_and_forget.updated_from"), user.Username, oldChannelDisplayName, newChannelDisplayName)

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      model.POST_DISPLAYNAME_CHANGE,
		UserId:    userId,
		Props: model.StringInterface{
			"username":        user.Username,
			"old_displayname": oldChannelDisplayName,
			"new_displayname": newChannelDisplayName,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("PostUpdateChannelDisplayNameMessage", "api.channel.post_update_channel_displayname_message_and_forget.create_post.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetChannel(channelId string) (*model.Channel, *model.AppError) {
	channel, err := a.Srv().Store.Channel().Get(channelId, true)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannel", "app.channel.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannel", "app.channel.get.find.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return channel, nil
}

func (a *App) GetChannelByName(channelName, teamId string, includeDeleted bool) (*model.Channel, *model.AppError) {
	var channel *model.Channel
	var err error

	if includeDeleted {
		channel, err = a.Srv().Store.Channel().GetByNameIncludeDeleted(teamId, channelName, false)
	} else {
		channel, err = a.Srv().Store.Channel().GetByName(teamId, channelName, false)
	}

	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelByName", "app.channel.get_by_name.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelByName", "app.channel.get_by_name.existing.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return channel, nil
}

func (a *App) GetChannelsByNames(channelNames []string, teamId string) ([]*model.Channel, *model.AppError) {
	channels, err := a.Srv().Store.Channel().GetByNames(teamId, channelNames, true)
	if err != nil {
		return nil, model.NewAppError("GetChannelsByNames", "app.channel.get_by_name.existing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return channels, nil
}

func (a *App) GetChannelByNameForTeamName(channelName, teamName string, includeDeleted bool) (*model.Channel, *model.AppError) {
	var team *model.Team

	team, err := a.Srv().Store.Team().GetByName(teamName)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelByNameForTeamName", "app.team.get_by_name.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelByNameForTeamName", "app.team.get_by_name.app_error", nil, err.Error(), http.StatusNotFound)
		}
	}

	var result *model.Channel

	var nErr error
	if includeDeleted {
		result, nErr = a.Srv().Store.Channel().GetByNameIncludeDeleted(team.Id, channelName, false)
	} else {
		result, nErr = a.Srv().Store.Channel().GetByName(team.Id, channelName, false)
	}

	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("GetChannelByNameForTeamName", "app.channel.get_by_name.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelByNameForTeamName", "app.channel.get_by_name.existing.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	return result, nil
}

func (a *App) GetChannelsForUser(teamId string, userId string, includeDeleted bool, lastDeleteAt int) (*model.ChannelList, *model.AppError) {
	list, err := a.Srv().Store.Channel().GetChannels(teamId, userId, includeDeleted, lastDeleteAt)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelsForUser", "app.channel.get_channels.not_found.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelsForUser", "app.channel.get_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return list, nil
}

func (a *App) GetAllChannels(page, perPage int, opts model.ChannelSearchOpts) (*model.ChannelListWithTeamData, *model.AppError) {
	if opts.ExcludeDefaultChannels {
		opts.ExcludeChannelNames = a.DefaultChannelNames()
	}
	storeOpts := store.ChannelSearchOpts{
		ExcludeChannelNames:  opts.ExcludeChannelNames,
		NotAssociatedToGroup: opts.NotAssociatedToGroup,
		IncludeDeleted:       opts.IncludeDeleted,
	}
	channels, err := a.Srv().Store.Channel().GetAllChannels(page*perPage, perPage, storeOpts)
	if err != nil {
		return nil, model.NewAppError("GetAllChannels", "app.channel.get_all_channels.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channels, nil
}

func (a *App) GetAllChannelsCount(opts model.ChannelSearchOpts) (int64, *model.AppError) {
	if opts.ExcludeDefaultChannels {
		opts.ExcludeChannelNames = a.DefaultChannelNames()
	}
	storeOpts := store.ChannelSearchOpts{
		ExcludeChannelNames:  opts.ExcludeChannelNames,
		NotAssociatedToGroup: opts.NotAssociatedToGroup,
		IncludeDeleted:       opts.IncludeDeleted,
	}
	count, err := a.Srv().Store.Channel().GetAllChannelsCount(storeOpts)
	if err != nil {
		return 0, model.NewAppError("GetAllChannelsCount", "app.channel.get_all_channels_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (a *App) GetDeletedChannels(teamId string, offset int, limit int, userId string) (*model.ChannelList, *model.AppError) {
	list, err := a.Srv().Store.Channel().GetDeleted(teamId, offset, limit, userId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetDeletedChannels", "app.channel.get_deleted.missing.app_error", nil, err.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetDeletedChannels", "app.channel.get_deleted.existing.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return list, nil
}

func (a *App) GetChannelsUserNotIn(teamId string, userId string, offset int, limit int) (*model.ChannelList, *model.AppError) {
	channels, err := a.Srv().Store.Channel().GetMoreChannels(teamId, userId, offset, limit)
	if err != nil {
		return nil, model.NewAppError("GetChannelsUserNotIn", "app.channel.get_more_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return channels, nil
}

func (a *App) GetPublicChannelsByIdsForTeam(teamId string, channelIds []string) (*model.ChannelList, *model.AppError) {
	list, err := a.Srv().Store.Channel().GetPublicChannelsByIdsForTeam(teamId, channelIds)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetPublicChannelsByIdsForTeam", "app.channel.get_channels_by_ids.not_found.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetPublicChannelsByIdsForTeam", "app.channel.get_channels_by_ids.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return list, nil
}

func (a *App) GetPublicChannelsForTeam(teamId string, offset int, limit int) (*model.ChannelList, *model.AppError) {
	list, err := a.Srv().Store.Channel().GetPublicChannelsForTeam(teamId, offset, limit)
	if err != nil {
		return nil, model.NewAppError("GetPublicChannelsForTeam", "app.channel.get_public_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return list, nil
}

func (a *App) GetPrivateChannelsForTeam(teamId string, offset int, limit int) (*model.ChannelList, *model.AppError) {
	list, err := a.Srv().Store.Channel().GetPrivateChannelsForTeam(teamId, offset, limit)
	if err != nil {
		return nil, model.NewAppError("GetPrivateChannelsForTeam", "app.channel.get_private_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return list, nil
}

func (a *App) GetChannelMember(channelId string, userId string) (*model.ChannelMember, *model.AppError) {
	channelMember, err := a.Srv().Store.Channel().GetMember(channelId, userId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelMember", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelMember", "app.channel.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return channelMember, nil
}

func (a *App) GetChannelMembersPage(channelId string, page, perPage int) (*model.ChannelMembers, *model.AppError) {
	channelMembers, err := a.Srv().Store.Channel().GetMembers(channelId, page*perPage, perPage)
	if err != nil {
		return nil, model.NewAppError("GetChannelMembersPage", "app.channel.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelMembers, nil
}

func (a *App) GetChannelMembersTimezones(channelId string) ([]string, *model.AppError) {
	membersTimezones, err := a.Srv().Store.Channel().GetChannelMembersTimezones(channelId)
	if err != nil {
		return nil, model.NewAppError("GetChannelMembersTimezones", "app.channel.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var timezones []string
	for _, membersTimezone := range membersTimezones {
		if membersTimezone["automaticTimezone"] == "" && membersTimezone["manualTimezone"] == "" {
			continue
		}
		timezones = append(timezones, model.GetPreferredTimezone(membersTimezone))
	}

	return model.RemoveDuplicateStrings(timezones), nil
}

func (a *App) GetChannelMembersByIds(channelId string, userIds []string) (*model.ChannelMembers, *model.AppError) {
	members, err := a.Srv().Store.Channel().GetMembersByIds(channelId, userIds)
	if err != nil {
		return nil, model.NewAppError("GetChannelMembersByIds", "app.channel.get_members_by_ids.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return members, nil
}

func (a *App) GetChannelMembersForUser(teamId string, userId string) (*model.ChannelMembers, *model.AppError) {
	channelMembers, err := a.Srv().Store.Channel().GetMembersForUser(teamId, userId)
	if err != nil {
		return nil, model.NewAppError("GetChannelMembersForUser", "app.channel.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelMembers, nil
}

func (a *App) GetChannelMembersForUserWithPagination(teamId, userId string, page, perPage int) ([]*model.ChannelMember, *model.AppError) {
	m, err := a.Srv().Store.Channel().GetMembersForUserWithPagination(teamId, userId, page, perPage)
	if err != nil {
		return nil, model.NewAppError("GetChannelMembersForUserWithPagination", "app.channel.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	members := make([]*model.ChannelMember, 0)
	if m != nil {
		for _, member := range *m {
			member := member
			members = append(members, &member)
		}
	}
	return members, nil
}

func (a *App) GetChannelMemberCount(channelId string) (int64, *model.AppError) {
	count, err := a.Srv().Store.Channel().GetMemberCount(channelId, true)
	if err != nil {
		return 0, model.NewAppError("GetChannelMemberCount", "app.channel.get_member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (a *App) GetChannelGuestCount(channelId string) (int64, *model.AppError) {
	count, err := a.Srv().Store.Channel().GetGuestCount(channelId, true)
	if err != nil {
		return 0, model.NewAppError("SqlChannelStore.GetGuestCount", "app.channel.get_member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (a *App) GetChannelPinnedPostCount(channelId string) (int64, *model.AppError) {
	count, err := a.Srv().Store.Channel().GetPinnedPostCount(channelId, true)
	if err != nil {
		return 0, model.NewAppError("GetChannelPinnedPostCount", "app.channel.get_pinnedpost_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (a *App) GetChannelCounts(teamId string, userId string) (*model.ChannelCounts, *model.AppError) {
	counts, err := a.Srv().Store.Channel().GetChannelCounts(teamId, userId)
	if err != nil {
		return nil, model.NewAppError("SqlChannelStore.GetChannelCounts", "app.channel.get_channel_counts.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return counts, nil
}

func (a *App) GetChannelUnread(channelId, userId string) (*model.ChannelUnread, *model.AppError) {
	channelUnread, err := a.Srv().Store.Channel().GetChannelUnread(channelId, userId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelUnread", "app.channel.get_unread.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelUnread", "app.channel.get_unread.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if channelUnread.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP] == model.CHANNEL_MARK_UNREAD_MENTION {
		channelUnread.MsgCount = 0
	}

	return channelUnread, nil
}

func (a *App) JoinChannel(channel *model.Channel, userId string) *model.AppError {
	userChan := make(chan store.StoreResult, 1)
	memberChan := make(chan store.StoreResult, 1)
	go func() {
		user, err := a.Srv().Store.User().Get(userId)
		userChan <- store.StoreResult{Data: user, NErr: err}
		close(userChan)
	}()
	go func() {
		member, err := a.Srv().Store.Channel().GetMember(channel.Id, userId)
		memberChan <- store.StoreResult{Data: member, NErr: err}
		close(memberChan)
	}()

	uresult := <-userChan
	if uresult.NErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(uresult.NErr, &nfErr):
			return model.NewAppError("CreateChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("CreateChannel", "app.user.get.app_error", nil, uresult.NErr.Error(), http.StatusInternalServerError)
		}
	}

	mresult := <-memberChan
	if mresult.NErr == nil && mresult.Data != nil {
		// user is already in the channel
		return nil
	}

	user := uresult.Data.(*model.User)

	if channel.Type != model.CHANNEL_OPEN {
		return model.NewAppError("JoinChannel", "api.channel.join_channel.permissions.app_error", nil, "", http.StatusBadRequest)
	}

	cm, err := a.AddUserToChannel(user, channel)
	if err != nil {
		return err
	}

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.UserHasJoinedChannel(pluginContext, cm, nil)
				return true
			}, plugin.UserHasJoinedChannelId)
		})
	}

	if err := a.postJoinChannelMessage(user, channel); err != nil {
		return err
	}

	return nil
}

func (a *App) postJoinChannelMessage(user *model.User, channel *model.Channel) *model.AppError {
	message := fmt.Sprintf(utils.T("api.channel.join_channel.post_and_forget"), user.Username)
	postType := model.POST_JOIN_CHANNEL

	if user.IsGuest() {
		message = fmt.Sprintf(utils.T("api.channel.guest_join_channel.post_and_forget"), user.Username)
		postType = model.POST_GUEST_JOIN_CHANNEL
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      postType,
		UserId:    user.Id,
		Props: model.StringInterface{
			"username": user.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postJoinChannelMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) postJoinTeamMessage(user *model.User, channel *model.Channel) *model.AppError {
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(utils.T("api.team.join_team.post_and_forget"), user.Username),
		Type:      model.POST_JOIN_TEAM,
		UserId:    user.Id,
		Props: model.StringInterface{
			"username": user.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postJoinTeamMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) LeaveChannel(channelId string, userId string) *model.AppError {
	sc := make(chan store.StoreResult, 1)
	go func() {
		channel, err := a.Srv().Store.Channel().Get(channelId, true)
		sc <- store.StoreResult{Data: channel, NErr: err}
		close(sc)
	}()

	uc := make(chan store.StoreResult, 1)
	go func() {
		user, err := a.Srv().Store.User().Get(userId)
		uc <- store.StoreResult{Data: user, NErr: err}
		close(uc)
	}()

	mcc := make(chan store.StoreResult, 1)
	go func() {
		count, err := a.Srv().Store.Channel().GetMemberCount(channelId, false)
		mcc <- store.StoreResult{Data: count, NErr: err}
		close(mcc)
	}()

	cresult := <-sc
	if cresult.NErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(cresult.NErr, &nfErr):
			return model.NewAppError("LeaveChannel", "app.channel.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("LeaveChannel", "app.channel.get.find.app_error", nil, cresult.NErr.Error(), http.StatusInternalServerError)
		}
	}
	uresult := <-uc
	if uresult.NErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(uresult.NErr, &nfErr):
			return model.NewAppError("LeaveChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("LeaveChannel", "app.user.get.app_error", nil, uresult.NErr.Error(), http.StatusInternalServerError)
		}
	}
	ccresult := <-mcc
	if ccresult.NErr != nil {
		return model.NewAppError("LeaveChannel", "app.channel.get_member_count.app_error", nil, ccresult.NErr.Error(), http.StatusInternalServerError)
	}

	channel := cresult.Data.(*model.Channel)
	user := uresult.Data.(*model.User)
	membersCount := ccresult.Data.(int64)

	if channel.IsGroupOrDirect() {
		err := model.NewAppError("LeaveChannel", "api.channel.leave.direct.app_error", nil, "", http.StatusBadRequest)
		return err
	}

	if channel.Type == model.CHANNEL_PRIVATE && membersCount == 1 {
		err := model.NewAppError("LeaveChannel", "api.channel.leave.last_member.app_error", nil, "userId="+user.Id, http.StatusBadRequest)
		return err
	}

	if err := a.removeUserFromChannel(userId, userId, channel); err != nil {
		return err
	}

	if channel.Name == model.DEFAULT_CHANNEL && !*a.Config().ServiceSettings.ExperimentalEnableDefaultChannelLeaveJoinMessages {
		return nil
	}

	a.Srv().Go(func() {
		a.postLeaveChannelMessage(user, channel)
	})

	return nil
}

func (a *App) postLeaveChannelMessage(user *model.User, channel *model.Channel) *model.AppError {
	post := &model.Post{
		ChannelId: channel.Id,
		// Message here embeds `@username`, not just `username`, to ensure that mentions
		// treat this as a username mention even though the user has now left the channel.
		// The client renders its own system message, ignoring this value altogether.
		Message: fmt.Sprintf(utils.T("api.channel.leave.left"), fmt.Sprintf("@%s", user.Username)),
		Type:    model.POST_LEAVE_CHANNEL,
		UserId:  user.Id,
		Props: model.StringInterface{
			"username": user.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postLeaveChannelMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) PostAddToChannelMessage(user *model.User, addedUser *model.User, channel *model.Channel, postRootId string) *model.AppError {
	message := fmt.Sprintf(utils.T("api.channel.add_member.added"), addedUser.Username, user.Username)
	postType := model.POST_ADD_TO_CHANNEL

	if addedUser.IsGuest() {
		message = fmt.Sprintf(utils.T("api.channel.add_guest.added"), addedUser.Username, user.Username)
		postType = model.POST_ADD_GUEST_TO_CHANNEL
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      postType,
		UserId:    user.Id,
		RootId:    postRootId,
		Props: model.StringInterface{
			"userId":                       user.Id,
			"username":                     user.Username,
			model.POST_PROPS_ADDED_USER_ID: addedUser.Id,
			"addedUsername":                addedUser.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postAddToChannelMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) postAddToTeamMessage(user *model.User, addedUser *model.User, channel *model.Channel, postRootId string) *model.AppError {
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(utils.T("api.team.add_user_to_team.added"), addedUser.Username, user.Username),
		Type:      model.POST_ADD_TO_TEAM,
		UserId:    user.Id,
		RootId:    postRootId,
		Props: model.StringInterface{
			"userId":                       user.Id,
			"username":                     user.Username,
			model.POST_PROPS_ADDED_USER_ID: addedUser.Id,
			"addedUsername":                addedUser.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postAddToTeamMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) postRemoveFromChannelMessage(removerUserId string, removedUser *model.User, channel *model.Channel) *model.AppError {
	post := &model.Post{
		ChannelId: channel.Id,
		// Message here embeds `@username`, not just `username`, to ensure that mentions
		// treat this as a username mention even though the user has now left the channel.
		// The client renders its own system message, ignoring this value altogether.
		Message: fmt.Sprintf(utils.T("api.channel.remove_member.removed"), fmt.Sprintf("@%s", removedUser.Username)),
		Type:    model.POST_REMOVE_FROM_CHANNEL,
		UserId:  removerUserId,
		Props: model.StringInterface{
			"removedUserId":   removedUser.Id,
			"removedUsername": removedUser.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postRemoveFromChannelMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) removeUserFromChannel(userIdToRemove string, removerUserId string, channel *model.Channel) *model.AppError {
	user, nErr := a.Srv().Store.User().Get(userIdToRemove)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return model.NewAppError("removeUserFromChannel", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("removeUserFromChannel", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}
	isGuest := user.IsGuest()

	if channel.Name == model.DEFAULT_CHANNEL {
		if !isGuest {
			return model.NewAppError("RemoveUserFromChannel", "api.channel.remove.default.app_error", map[string]interface{}{"Channel": model.DEFAULT_CHANNEL}, "", http.StatusBadRequest)
		}
	}

	if channel.IsGroupConstrained() && userIdToRemove != removerUserId && !user.IsBot {
		nonMembers, err := a.FilterNonGroupChannelMembers([]string{userIdToRemove}, channel)
		if err != nil {
			return model.NewAppError("removeUserFromChannel", "api.channel.remove_user_from_channel.app_error", nil, "", http.StatusInternalServerError)
		}
		if len(nonMembers) == 0 {
			return model.NewAppError("removeUserFromChannel", "api.channel.remove_members.denied", map[string]interface{}{"UserIDs": nonMembers}, "", http.StatusBadRequest)
		}
	}

	cm, err := a.GetChannelMember(channel.Id, userIdToRemove)
	if err != nil {
		return err
	}

	if err := a.Srv().Store.Channel().RemoveMember(channel.Id, userIdToRemove); err != nil {
		return model.NewAppError("removeUserFromChannel", "app.channel.remove_member.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	if err := a.Srv().Store.ChannelMemberHistory().LogLeaveEvent(userIdToRemove, channel.Id, model.GetMillis()); err != nil {
		return model.NewAppError("removeUserFromChannel", "app.channel_member_history.log_leave_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if isGuest {
		currentMembers, err := a.GetChannelMembersForUser(channel.TeamId, userIdToRemove)
		if err != nil {
			return err
		}
		if len(*currentMembers) == 0 {
			teamMember, err := a.GetTeamMember(channel.TeamId, userIdToRemove)
			if err != nil {
				return model.NewAppError("removeUserFromChannel", "api.team.remove_user_from_team.missing.app_error", nil, err.Error(), http.StatusBadRequest)
			}

			if err = a.RemoveTeamMemberFromTeam(teamMember, removerUserId); err != nil {
				return err
			}
		}
	}

	a.InvalidateCacheForUser(userIdToRemove)
	a.invalidateCacheForChannelMembers(channel.Id)

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		var actorUser *model.User
		if removerUserId != "" {
			actorUser, _ = a.GetUser(removerUserId)
		}

		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.UserHasLeftChannel(pluginContext, cm, actorUser)
				return true
			}, plugin.UserHasLeftChannelId)
		})
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_REMOVED, "", channel.Id, "", nil)
	message.Add("user_id", userIdToRemove)
	message.Add("remover_id", removerUserId)
	a.Publish(message)

	// because the removed user no longer belongs to the channel we need to send a separate websocket event
	userMsg := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_REMOVED, "", "", userIdToRemove, nil)
	userMsg.Add("channel_id", channel.Id)
	userMsg.Add("remover_id", removerUserId)
	a.Publish(userMsg)

	return nil
}

func (a *App) RemoveUserFromChannel(userIdToRemove string, removerUserId string, channel *model.Channel) *model.AppError {
	var err *model.AppError

	if err = a.removeUserFromChannel(userIdToRemove, removerUserId, channel); err != nil {
		return err
	}

	var user *model.User
	if user, err = a.GetUser(userIdToRemove); err != nil {
		return err
	}

	if userIdToRemove == removerUserId {
		a.postLeaveChannelMessage(user, channel)
	} else {
		a.Srv().Go(func() {
			a.postRemoveFromChannelMessage(removerUserId, user, channel)
		})
	}

	return nil
}

func (a *App) GetNumberOfChannelsOnTeam(teamId string) (int, *model.AppError) {
	// Get total number of channels on current team
	list, err := a.Srv().Store.Channel().GetTeamChannels(teamId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return 0, model.NewAppError("GetNumberOfChannelsOnTeam", "app.channel.get_channels.not_found.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return 0, model.NewAppError("GetNumberOfChannelsOnTeam", "app.channel.get_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return len(*list), nil
}

func (a *App) SetActiveChannel(userId string, channelId string) *model.AppError {
	status, err := a.GetStatus(userId)

	oldStatus := model.STATUS_OFFLINE

	if err != nil {
		status = &model.Status{UserId: userId, Status: model.STATUS_ONLINE, Manual: false, LastActivityAt: model.GetMillis(), ActiveChannel: channelId}
	} else {
		oldStatus = status.Status
		status.ActiveChannel = channelId
		if !status.Manual && channelId != "" {
			status.Status = model.STATUS_ONLINE
		}
		status.LastActivityAt = model.GetMillis()
	}

	a.AddStatusCache(status)

	if status.Status != oldStatus {
		a.BroadcastStatus(status)
	}

	return nil
}

func (a *App) UpdateChannelLastViewedAt(channelIds []string, userId string) *model.AppError {
	if _, err := a.Srv().Store.Channel().UpdateLastViewedAt(channelIds, userId, *a.Config().ServiceSettings.ThreadAutoFollow); err != nil {
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return model.NewAppError("UpdateChannelLastViewedAt", "app.channel.update_last_viewed_at.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return model.NewAppError("UpdateChannelLastViewedAt", "app.channel.update_last_viewed_at.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if *a.Config().ServiceSettings.EnableChannelViewedMessages {
		for _, channelId := range channelIds {
			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_VIEWED, "", "", userId, nil)
			message.Add("channel_id", channelId)
			a.Publish(message)
		}
	}

	return nil
}

// MarkChanelAsUnreadFromPost will take a post and set the channel as unread from that one.
func (a *App) MarkChannelAsUnreadFromPost(postID string, userID string) (*model.ChannelUnreadAt, *model.AppError) {
	post, err := a.GetSinglePost(postID)
	if err != nil {
		return nil, err
	}

	user, err := a.GetUser(userID)
	if err != nil {
		return nil, err
	}

	unreadMentions, err := a.countMentionsFromPost(user, post)
	if err != nil {
		return nil, err
	}

	if *a.Config().ServiceSettings.ThreadAutoFollow && post.RootId != "" {
		threadMembership, _ := a.Srv().Store.Thread().GetMembershipForUser(user.Id, post.RootId)
		if threadMembership != nil {
			channel, nErr := a.Srv().Store.Channel().Get(post.ChannelId, true)
			if nErr != nil {
				return nil, model.NewAppError("MarkChannelAsUnreadFromPost", "app.channel.update_last_viewed_at_post.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
			threadMembership.UnreadMentions, err = a.countThreadMentions(user, post, channel.TeamId, post.UpdateAt-1)
			if err != nil {
				return nil, err
			}
			_, nErr = a.Srv().Store.Thread().UpdateMembership(threadMembership)
			if nErr != nil {
				return nil, model.NewAppError("MarkChannelAsUnreadFromPost", "app.channel.update_last_viewed_at_post.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
	}

	channelUnread, nErr := a.Srv().Store.Channel().UpdateLastViewedAtPost(post, userID, unreadMentions, *a.Config().ServiceSettings.ThreadAutoFollow)
	if nErr != nil {
		return channelUnread, model.NewAppError("MarkChannelAsUnreadFromPost", "app.channel.update_last_viewed_at_post.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_POST_UNREAD, channelUnread.TeamId, channelUnread.ChannelId, channelUnread.UserId, nil)
	message.Add("msg_count", channelUnread.MsgCount)
	message.Add("mention_count", channelUnread.MentionCount)
	message.Add("last_viewed_at", channelUnread.LastViewedAt)
	message.Add("post_id", postID)
	a.Publish(message)

	a.UpdateMobileAppBadge(userID)

	return channelUnread, nil
}

func (a *App) AutocompleteChannels(teamId string, term string) (*model.ChannelList, *model.AppError) {
	includeDeleted := *a.Config().TeamSettings.ExperimentalViewArchivedChannels
	term = strings.TrimSpace(term)

	channelList, err := a.Srv().Store.Channel().AutocompleteInTeam(teamId, term, includeDeleted)
	if err != nil {
		return nil, model.NewAppError("AutocompleteChannels", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

func (a *App) AutocompleteChannelsForSearch(teamId string, userId string, term string) (*model.ChannelList, *model.AppError) {
	includeDeleted := *a.Config().TeamSettings.ExperimentalViewArchivedChannels

	term = strings.TrimSpace(term)

	channelList, err := a.Srv().Store.Channel().AutocompleteInTeamForSearch(teamId, userId, term, includeDeleted)
	if err != nil {
		return nil, model.NewAppError("AutocompleteChannelsForSearch", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

// SearchAllChannels returns a list of channels, the total count of the results of the search (if the paginate search option is true), and an error.
func (a *App) SearchAllChannels(term string, opts model.ChannelSearchOpts) (*model.ChannelListWithTeamData, int64, *model.AppError) {
	if opts.ExcludeDefaultChannels {
		opts.ExcludeChannelNames = a.DefaultChannelNames()
	}
	storeOpts := store.ChannelSearchOpts{
		ExcludeChannelNames:     opts.ExcludeChannelNames,
		NotAssociatedToGroup:    opts.NotAssociatedToGroup,
		IncludeDeleted:          opts.IncludeDeleted,
		Deleted:                 opts.Deleted,
		TeamIds:                 opts.TeamIds,
		GroupConstrained:        opts.GroupConstrained,
		ExcludeGroupConstrained: opts.ExcludeGroupConstrained,
		Public:                  opts.Public,
		Private:                 opts.Private,
		Page:                    opts.Page,
		PerPage:                 opts.PerPage,
	}

	term = strings.TrimSpace(term)

	channelList, totalCount, err := a.Srv().Store.Channel().SearchAllChannels(term, storeOpts)
	if err != nil {
		return nil, 0, model.NewAppError("SearchAllChannels", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, totalCount, nil
}

func (a *App) SearchChannels(teamId string, term string) (*model.ChannelList, *model.AppError) {
	includeDeleted := *a.Config().TeamSettings.ExperimentalViewArchivedChannels

	term = strings.TrimSpace(term)

	channelList, err := a.Srv().Store.Channel().SearchInTeam(teamId, term, includeDeleted)
	if err != nil {
		return nil, model.NewAppError("SearchChannels", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

func (a *App) SearchArchivedChannels(teamId string, term string, userId string) (*model.ChannelList, *model.AppError) {
	term = strings.TrimSpace(term)

	channelList, err := a.Srv().Store.Channel().SearchArchivedInTeam(teamId, term, userId)
	if err != nil {
		return nil, model.NewAppError("SearchArchivedChannels", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

func (a *App) SearchChannelsForUser(userId, teamId, term string) (*model.ChannelList, *model.AppError) {
	includeDeleted := *a.Config().TeamSettings.ExperimentalViewArchivedChannels

	term = strings.TrimSpace(term)

	channelList, err := a.Srv().Store.Channel().SearchForUserInTeam(userId, teamId, term, includeDeleted)
	if err != nil {
		return nil, model.NewAppError("SearchChannelsForUser", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

func (a *App) SearchGroupChannels(userId, term string) (*model.ChannelList, *model.AppError) {
	if term == "" {
		return &model.ChannelList{}, nil
	}

	channelList, err := a.Srv().Store.Channel().SearchGroupChannels(userId, term)
	if err != nil {
		return nil, model.NewAppError("SearchGroupChannels", "app.channel.search_group_channels.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return channelList, nil
}

func (a *App) SearchChannelsUserNotIn(teamId string, userId string, term string) (*model.ChannelList, *model.AppError) {
	term = strings.TrimSpace(term)
	channelList, err := a.Srv().Store.Channel().SearchMore(userId, teamId, term)
	if err != nil {
		return nil, model.NewAppError("SearchChannelsUserNotIn", "app.channel.search.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelList, nil
}

func (a *App) MarkChannelsAsViewed(channelIds []string, userId string, currentSessionId string) (map[string]int64, *model.AppError) {
	// I start looking for channels with notifications before I mark it as read, to clear the push notifications if needed
	channelsToClearPushNotifications := []string{}
	if *a.Config().EmailSettings.SendPushNotifications {
		for _, channelId := range channelIds {
			channel, errCh := a.Srv().Store.Channel().Get(channelId, true)
			if errCh != nil {
				mlog.Warn("Failed to get channel", mlog.Err(errCh))
				continue
			}

			member, err := a.Srv().Store.Channel().GetMember(channelId, userId)
			if err != nil {
				mlog.Warn("Failed to get membership", mlog.Err(err))
				continue
			}

			notify := member.NotifyProps[model.PUSH_NOTIFY_PROP]
			if notify == model.CHANNEL_NOTIFY_DEFAULT {
				user, err := a.GetUser(userId)
				if err != nil {
					mlog.Warn("Failed to get user", mlog.String("user_id", userId), mlog.Err(err))
					continue
				}
				notify = user.NotifyProps[model.PUSH_NOTIFY_PROP]
			}
			if notify == model.USER_NOTIFY_ALL {
				if count, err := a.Srv().Store.User().GetAnyUnreadPostCountForChannel(userId, channelId); err == nil {
					if count > 0 {
						channelsToClearPushNotifications = append(channelsToClearPushNotifications, channelId)
					}
				}
			} else if notify == model.USER_NOTIFY_MENTION || channel.Type == model.CHANNEL_DIRECT {
				if count, err := a.Srv().Store.User().GetUnreadCountForChannel(userId, channelId); err == nil {
					if count > 0 {
						channelsToClearPushNotifications = append(channelsToClearPushNotifications, channelId)
					}
				}
			}
		}
	}
	times, err := a.Srv().Store.Channel().UpdateLastViewedAt(channelIds, userId, *a.Config().ServiceSettings.ThreadAutoFollow)
	if err != nil {
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("MarkChannelsAsViewed", "app.channel.update_last_viewed_at.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("MarkChannelsAsViewed", "app.channel.update_last_viewed_at.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if *a.Config().ServiceSettings.EnableChannelViewedMessages {
		for _, channelId := range channelIds {
			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_VIEWED, "", "", userId, nil)
			message.Add("channel_id", channelId)
			a.Publish(message)
		}
	}
	for _, channelId := range channelsToClearPushNotifications {
		a.clearPushNotification(currentSessionId, userId, channelId)
	}
	return times, nil
}

func (a *App) ViewChannel(view *model.ChannelView, userId string, currentSessionId string) (map[string]int64, *model.AppError) {
	if err := a.SetActiveChannel(userId, view.ChannelId); err != nil {
		return nil, err
	}

	channelIds := []string{}

	if len(view.ChannelId) > 0 {
		channelIds = append(channelIds, view.ChannelId)
	}

	if len(view.PrevChannelId) > 0 {
		channelIds = append(channelIds, view.PrevChannelId)
	}

	if len(channelIds) == 0 {
		return map[string]int64{}, nil
	}

	return a.MarkChannelsAsViewed(channelIds, userId, currentSessionId)
}

func (a *App) PermanentDeleteChannel(channel *model.Channel) *model.AppError {
	if err := a.Srv().Store.Post().PermanentDeleteByChannel(channel.Id); err != nil {
		return model.NewAppError("PermanentDeleteChannel", "app.post.permanent_delete_by_channel.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := a.Srv().Store.Channel().PermanentDeleteMembersByChannel(channel.Id); err != nil {
		return model.NewAppError("PermanentDeleteChannel", "app.channel.remove_member.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := a.Srv().Store.Webhook().PermanentDeleteIncomingByChannel(channel.Id); err != nil {
		return model.NewAppError("PermanentDeleteChannel", "app.webhooks.permanent_delete_incoming_by_channel.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := a.Srv().Store.Webhook().PermanentDeleteOutgoingByChannel(channel.Id); err != nil {
		return model.NewAppError("PermanentDeleteChannel", "app.webhooks.permanent_delete_outgoing_by_channel.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	deleteAt := model.GetMillis()

	if nErr := a.Srv().Store.Channel().PermanentDelete(channel.Id); nErr != nil {
		return model.NewAppError("PermanentDeleteChannel", "app.channel.permanent_delete.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	a.invalidateCacheForChannel(channel)
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_DELETED, channel.TeamId, "", "", nil)
	message.Add("channel_id", channel.Id)
	message.Add("delete_at", deleteAt)
	a.Publish(message)

	return nil
}

func (a *App) RemoveAllDeactivatedMembersFromChannel(channel *model.Channel) *model.AppError {
	err := a.Srv().Store.Channel().RemoveAllDeactivatedMembers(channel.Id)
	if err != nil {
		return model.NewAppError("RemoveAllDeactivatedMembersFromChannel", "app.channel.remove_all_deactivated_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// MoveChannel method is prone to data races if someone joins to channel during the move process. However this
// function is only exposed to sysadmins and the possibility of this edge case is relatively small.
func (a *App) MoveChannel(team *model.Team, channel *model.Channel, user *model.User) *model.AppError {
	// Check that all channel members are in the destination team.
	channelMembers, err := a.GetChannelMembersPage(channel.Id, 0, 10000000)
	if err != nil {
		return err
	}

	channelMemberIds := []string{}
	for _, channelMember := range *channelMembers {
		channelMemberIds = append(channelMemberIds, channelMember.UserId)
	}

	if len(channelMemberIds) > 0 {
		teamMembers, err2 := a.GetTeamMembersByIds(team.Id, channelMemberIds, nil)
		if err2 != nil {
			return err2
		}

		if len(teamMembers) != len(*channelMembers) {
			teamMembersMap := make(map[string]*model.TeamMember, len(teamMembers))
			for _, teamMember := range teamMembers {
				teamMembersMap[teamMember.UserId] = teamMember
			}
			for _, channelMember := range *channelMembers {
				if _, ok := teamMembersMap[channelMember.UserId]; !ok {
					mlog.Warn("Not member of the target team", mlog.String("userId", channelMember.UserId))
				}
			}
			return model.NewAppError("MoveChannel", "app.channel.move_channel.members_do_not_match.error", nil, "", http.StatusInternalServerError)
		}
	}

	// keep instance of the previous team
	previousTeam, nErr := a.Srv().Store.Team().Get(channel.TeamId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return model.NewAppError("MoveChannel", "app.team.get.find.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("MoveChannel", "app.team.get.finding.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if nErr := a.Srv().Store.Channel().UpdateSidebarChannelCategoryOnMove(channel, team.Id); nErr != nil {
		return model.NewAppError("MoveChannel", "app.channel.sidebar_categories.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	channel.TeamId = team.Id
	if _, err := a.Srv().Store.Channel().Update(channel); err != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return model.NewAppError("MoveChannel", "app.channel.update.bad_id", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return appErr
		default:
			return model.NewAppError("MoveChannel", "app.channel.update_channel.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if incomingWebhooks, err := a.GetIncomingWebhooksForTeamPage(previousTeam.Id, 0, 10000000); err != nil {
		mlog.Warn("Failed to get incoming webhooks", mlog.Err(err))
	} else {
		for _, webhook := range incomingWebhooks {
			if webhook.ChannelId == channel.Id {
				webhook.TeamId = team.Id
				if _, err := a.Srv().Store.Webhook().UpdateIncoming(webhook); err != nil {
					mlog.Warn("Failed to move incoming webhook to new team", mlog.String("webhook id", webhook.Id))
				}
			}
		}
	}

	if outgoingWebhooks, err := a.GetOutgoingWebhooksForTeamPage(previousTeam.Id, 0, 10000000); err != nil {
		mlog.Warn("Failed to get outgoing webhooks", mlog.Err(err))
	} else {
		for _, webhook := range outgoingWebhooks {
			if webhook.ChannelId == channel.Id {
				webhook.TeamId = team.Id
				if _, err := a.Srv().Store.Webhook().UpdateOutgoing(webhook); err != nil {
					mlog.Warn("Failed to move outgoing webhook to new team.", mlog.String("webhook id", webhook.Id))
				}
			}
		}
	}

	if err := a.RemoveUsersFromChannelNotMemberOfTeam(user, channel, team); err != nil {
		mlog.Warn("error while removing non-team member users", mlog.Err(err))
	}

	if user != nil {
		if err := a.postChannelMoveMessage(user, channel, previousTeam); err != nil {
			mlog.Warn("error while posting move channel message", mlog.Err(err))
		}
	}

	return nil
}

func (a *App) postChannelMoveMessage(user *model.User, channel *model.Channel, previousTeam *model.Team) *model.AppError {

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(utils.T("api.team.move_channel.success"), previousTeam.Name),
		Type:      model.POST_MOVE_CHANNEL,
		UserId:    user.Id,
		Props: model.StringInterface{
			"username": user.Username,
		},
	}

	if _, err := a.CreatePost(post, channel, false, true); err != nil {
		return model.NewAppError("postChannelMoveMessage", "api.team.move_channel.post.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) RemoveUsersFromChannelNotMemberOfTeam(remover *model.User, channel *model.Channel, team *model.Team) *model.AppError {
	channelMembers, err := a.GetChannelMembersPage(channel.Id, 0, 10000000)
	if err != nil {
		return err
	}

	channelMemberIds := []string{}
	channelMemberMap := make(map[string]struct{})
	for _, channelMember := range *channelMembers {
		channelMemberMap[channelMember.UserId] = struct{}{}
		channelMemberIds = append(channelMemberIds, channelMember.UserId)
	}

	if len(channelMemberIds) > 0 {
		teamMembers, err := a.GetTeamMembersByIds(team.Id, channelMemberIds, nil)
		if err != nil {
			return err
		}

		if len(teamMembers) != len(*channelMembers) {
			for _, teamMember := range teamMembers {
				delete(channelMemberMap, teamMember.UserId)
			}

			var removerId string
			if remover != nil {
				removerId = remover.Id
			}
			for userId := range channelMemberMap {
				if err := a.removeUserFromChannel(userId, removerId, channel); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *App) GetPinnedPosts(channelId string) (*model.PostList, *model.AppError) {
	posts, err := a.Srv().Store.Channel().GetPinnedPosts(channelId)
	if err != nil {
		return nil, model.NewAppError("GetPinnedPosts", "app.channel.pinned_posts.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return posts, nil
}

func (a *App) ToggleMuteChannel(channelId, userId string) (*model.ChannelMember, *model.AppError) {
	member, nErr := a.Srv().Store.Channel().GetMember(channelId, userId)
	if nErr != nil {
		var appErr *model.AppError
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("ToggleMuteChannel", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("ToggleMuteChannel", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	member.SetChannelMuted(!member.IsChannelMuted())

	member, err := a.updateChannelMember(member)
	if err != nil {
		return nil, err
	}

	a.invalidateCacheForChannelMembersNotifyProps(member.ChannelId)

	return member, nil
}

func (a *App) setChannelsMuted(channelIds []string, userId string, muted bool) ([]*model.ChannelMember, *model.AppError) {
	members, nErr := a.Srv().Store.Channel().GetMembersByChannelIds(channelIds, userId)
	if nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("setChannelsMuted", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	var membersToUpdate []*model.ChannelMember
	for _, member := range *members {
		if muted == member.IsChannelMuted() {
			continue
		}

		updatedMember := member
		updatedMember.SetChannelMuted(muted)

		membersToUpdate = append(membersToUpdate, &updatedMember)
	}

	if len(membersToUpdate) == 0 {
		return nil, nil
	}

	updated, nErr := a.Srv().Store.Channel().UpdateMultipleMembers(membersToUpdate)
	if nErr != nil {
		var appErr *model.AppError
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("setChannelsMuted", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("setChannelsMuted", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	for _, member := range updated {
		a.invalidateCacheForChannelMembersNotifyProps(member.ChannelId)

		evt := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED, "", "", member.UserId, nil)
		evt.Add("channelMember", member.ToJson())
		a.Publish(evt)
	}

	return updated, nil
}

func (a *App) FillInChannelProps(channel *model.Channel) *model.AppError {
	return a.FillInChannelsProps(&model.ChannelList{channel})
}

func (a *App) FillInChannelsProps(channelList *model.ChannelList) *model.AppError {
	// Group the channels by team and call GetChannelsByNames just once per team.
	channelsByTeam := make(map[string]model.ChannelList)
	for _, channel := range *channelList {
		channelsByTeam[channel.TeamId] = append(channelsByTeam[channel.TeamId], channel)
	}

	for teamId, channelList := range channelsByTeam {
		allChannelMentions := make(map[string]bool)
		channelMentions := make(map[*model.Channel][]string, len(channelList))

		// Collect mentions across the channels so as to query just once for this team.
		for _, channel := range channelList {
			channelMentions[channel] = model.ChannelMentions(channel.Header)

			for _, channelMention := range channelMentions[channel] {
				allChannelMentions[channelMention] = true
			}
		}

		allChannelMentionNames := make([]string, 0, len(allChannelMentions))
		for channelName := range allChannelMentions {
			allChannelMentionNames = append(allChannelMentionNames, channelName)
		}

		if len(allChannelMentionNames) > 0 {
			mentionedChannels, err := a.GetChannelsByNames(allChannelMentionNames, teamId)
			if err != nil {
				return err
			}

			mentionedChannelsByName := make(map[string]*model.Channel)
			for _, channel := range mentionedChannels {
				mentionedChannelsByName[channel.Name] = channel
			}

			for _, channel := range channelList {
				channelMentionsProp := make(map[string]interface{}, len(channelMentions[channel]))
				for _, channelMention := range channelMentions[channel] {
					if mentioned, ok := mentionedChannelsByName[channelMention]; ok {
						if mentioned.Type == model.CHANNEL_OPEN {
							channelMentionsProp[mentioned.Name] = map[string]interface{}{
								"display_name": mentioned.DisplayName,
							}
						}
					}
				}

				if len(channelMentionsProp) > 0 {
					channel.AddProp("channel_mentions", channelMentionsProp)
				} else if channel.Props != nil {
					delete(channel.Props, "channel_mentions")
				}
			}
		}
	}

	return nil
}

func (a *App) ClearChannelMembersCache(channelID string) {
	perPage := 100
	page := 0

	for {
		channelMembers, err := a.Srv().Store.Channel().GetMembers(channelID, page*perPage, perPage)
		if err != nil {
			a.Log().Warn("error clearing cache for channel members", mlog.String("channel_id", channelID))
			break
		}

		for _, channelMember := range *channelMembers {
			a.ClearSessionCacheForUser(channelMember.UserId)

			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED, "", "", channelMember.UserId, nil)
			message.Add("channelMember", channelMember.ToJson())
			a.Publish(message)
		}

		length := len(*(channelMembers))
		if length < perPage {
			break
		}

		page++
	}
}

func (a *App) GetMemberCountsByGroup(channelID string, includeTimezones bool) ([]*model.ChannelMemberCountByGroup, *model.AppError) {
	channelMemberCounts, err := a.Srv().Store.Channel().GetMemberCountsByGroup(channelID, includeTimezones)
	if err != nil {
		return nil, model.NewAppError("GetMemberCountsByGroup", "app.channel.get_member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelMemberCounts, nil
}

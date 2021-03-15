// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (a *App) SaveReactionForPost(reaction *model.Reaction) (*model.Reaction, *model.AppError) {
	post, err := a.GetSinglePost(reaction.PostId)
	if err != nil {
		return nil, err
	}

	channel, err := a.GetChannel(post.ChannelId)
	if err != nil {
		return nil, err
	}

	if channel.DeleteAt > 0 {
		return nil, model.NewAppError("deleteReactionForPost", "api.reaction.save.archived_channel.app_error", nil, "", http.StatusForbidden)
	}

	if a.Srv().License() != nil && *a.Config().TeamSettings.ExperimentalTownSquareIsReadOnly && channel.Name == model.DEFAULT_CHANNEL {
		var user *model.User
		user, err = a.GetUser(reaction.UserId)
		if err != nil {
			return nil, err
		}

		if !a.RolesGrantPermission(user.GetRoles(), model.PERMISSION_MANAGE_SYSTEM.Id) {
			return nil, model.NewAppError("saveReactionForPost", "api.reaction.town_square_read_only", nil, "", http.StatusForbidden)
		}
	}

	reaction, nErr := a.Srv().Store.Reaction().Save(reaction)
	if nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("SaveReactionForPost", "app.reaction.save.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	// The post is always modified since the UpdateAt always changes
	a.invalidateCacheForChannelPosts(post.ChannelId)

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.ReactionHasBeenAdded(pluginContext, reaction)
				return true
			}, plugin.ReactionHasBeenAddedID)
		})
	}

	a.Srv().Go(func() {
		a.sendReactionEvent(model.WEBSOCKET_EVENT_REACTION_ADDED, reaction, post)
	})

	return reaction, nil
}

func (a *App) GetReactionsForPost(postID string) ([]*model.Reaction, *model.AppError) {
	reactions, err := a.Srv().Store.Reaction().GetForPost(postID, true)
	if err != nil {
		return nil, model.NewAppError("GetReactionsForPost", "app.reaction.get_for_post.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return reactions, nil
}

func (a *App) GetBulkReactionsForPosts(postIDs []string) (map[string][]*model.Reaction, *model.AppError) {
	reactions := make(map[string][]*model.Reaction)

	allReactions, err := a.Srv().Store.Reaction().BulkGetForPosts(postIDs)
	if err != nil {
		return nil, model.NewAppError("GetBulkReactionsForPosts", "app.reaction.bulk_get_for_post_ids.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, reaction := range allReactions {
		reactionsForPost := reactions[reaction.PostId]
		reactionsForPost = append(reactionsForPost, reaction)

		reactions[reaction.PostId] = reactionsForPost
	}

	reactions = populateEmptyReactions(postIDs, reactions)
	return reactions, nil
}

func populateEmptyReactions(postIDs []string, reactions map[string][]*model.Reaction) map[string][]*model.Reaction {
	for _, postID := range postIDs {
		if _, present := reactions[postID]; !present {
			reactions[postID] = []*model.Reaction{}
		}
	}
	return reactions
}

func (a *App) DeleteReactionForPost(reaction *model.Reaction) *model.AppError {
	post, err := a.GetSinglePost(reaction.PostId)
	if err != nil {
		return err
	}

	channel, err := a.GetChannel(post.ChannelId)
	if err != nil {
		return err
	}

	if channel.DeleteAt > 0 {
		return model.NewAppError("DeleteReactionForPost", "api.reaction.delete.archived_channel.app_error", nil, "", http.StatusForbidden)
	}

	if a.Srv().License() != nil && *a.Config().TeamSettings.ExperimentalTownSquareIsReadOnly && channel.Name == model.DEFAULT_CHANNEL {
		user, err := a.GetUser(reaction.UserId)
		if err != nil {
			return err
		}

		if !a.RolesGrantPermission(user.GetRoles(), model.PERMISSION_MANAGE_SYSTEM.Id) {
			return model.NewAppError("DeleteReactionForPost", "api.reaction.town_square_read_only", nil, "", http.StatusForbidden)
		}
	}

	if _, err := a.Srv().Store.Reaction().Delete(reaction); err != nil {
		return model.NewAppError("DeleteReactionForPost", "app.reaction.delete_all_with_emoji_name.get_reactions.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	// The post is always modified since the UpdateAt always changes
	a.invalidateCacheForChannelPosts(post.ChannelId)

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.ReactionHasBeenRemoved(pluginContext, reaction)
				return true
			}, plugin.ReactionHasBeenRemovedID)
		})
	}

	a.Srv().Go(func() {
		a.sendReactionEvent(model.WEBSOCKET_EVENT_REACTION_REMOVED, reaction, post)
	})

	return nil
}

func (a *App) sendReactionEvent(event string, reaction *model.Reaction, post *model.Post) {
	// send out that a reaction has been added/removed
	message := model.NewWebSocketEvent(event, "", post.ChannelId, "", nil)
	message.Add("reaction", reaction.ToJson())
	a.Publish(message)
}

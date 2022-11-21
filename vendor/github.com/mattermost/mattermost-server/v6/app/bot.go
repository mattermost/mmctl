// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/app/request"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/product"
	"github.com/mattermost/mattermost-server/v6/shared/i18n"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/mattermost/mattermost-server/v6/store"
)

const (
	internalKeyPrefix = "mmi_"
	botUserKey        = internalKeyPrefix + "botid"
)

// Ensure bot service wrapper implements `product.BotService`
var _ product.BotService = (*botServiceWrapper)(nil)

// botServiceWrapper provides an implementation of `product.BotService` for use by products.
type botServiceWrapper struct {
	app AppIface
}

func (w *botServiceWrapper) EnsureBot(c *request.Context, productID string, bot *model.Bot) (string, error) {
	return w.app.EnsureBot(c, productID, bot)
}

// EnsureBot provides similar functionality with the plugin-api BotService. It doesn't accept
// any ensureBotOptions hence it is not required for now.
// TODO: Once the focalboard migration completed, we should add this logic to the app and
// let plugin-api use the same code
func (a *App) EnsureBot(c request.CTX, productID string, bot *model.Bot) (string, error) {
	if bot == nil {
		return "", errors.New("passed a nil bot")
	}

	if bot.Username == "" {
		return "", errors.New("passed a bot with no username")
	}

	botIDBytes, err := a.GetPluginKey(productID, botUserKey)
	if err != nil {
		return "", err
	}

	// If the bot has already been created, use it
	if botIDBytes != nil {
		botID := string(botIDBytes)

		// ensure existing bot is synced with what is being created
		botPatch := &model.BotPatch{
			Username:    &bot.Username,
			DisplayName: &bot.DisplayName,
			Description: &bot.Description,
		}

		if _, err = a.PatchBot(botID, botPatch); err != nil {
			return "", fmt.Errorf("failed to patch bot: %w", err)
		}

		return botID, nil
	}

	// Check for an existing bot user with that username. If one exists, then use that.
	if user, appErr := a.GetUserByUsername(bot.Username); appErr == nil && user != nil {
		if user.IsBot {
			if appErr := a.SetPluginKey(productID, botUserKey, []byte(user.Id)); appErr != nil {
				return "", fmt.Errorf("failed to set plugin key: %w", err)
			}
		} else {
			c.Logger().Error("Product attempted to use an account that already exists. Convert user to a bot "+
				"account in the CLI by running 'mattermost user convert <username> --bot'. If the user is an "+
				"existing user account you want to preserve, change its username and restart the Mattermost server, "+
				"after which the plugin will create a bot account with that name. For more information about bot "+
				"accounts, see https://mattermost.com/pl/default-bot-accounts", mlog.String("username",
				bot.Username),
				mlog.String("user_id",
					user.Id),
			)
		}
		return user.Id, nil
	}

	createdBot, err := a.CreateBot(c, bot)
	if err != nil {
		return "", fmt.Errorf("failed to create bot: %w", err)
	}

	if appErr := a.SetPluginKey(productID, botUserKey, []byte(createdBot.UserId)); appErr != nil {
		return "", fmt.Errorf("failed to set plugin key: %w", err)
	}

	return createdBot.UserId, nil
}

// CreateBot creates the given bot and corresponding user.
func (a *App) CreateBot(c request.CTX, bot *model.Bot) (*model.Bot, *model.AppError) {
	vErr := bot.IsValidCreate()
	if vErr != nil {
		return nil, vErr
	}

	user, nErr := a.Srv().Store().User().Save(model.UserFromBot(bot))
	if nErr != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &invErr):
			code := ""
			switch invErr.Field {
			case "email":
				code = "app.user.save.email_exists.app_error"
			case "username":
				code = "app.user.save.username_exists.app_error"
			default:
				code = "app.user.save.existing.app_error"
			}
			return nil, model.NewAppError("CreateBot", code, nil, "", http.StatusBadRequest).Wrap(nErr)
		default:
			return nil, model.NewAppError("CreateBot", "app.user.save.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}
	bot.UserId = user.Id

	savedBot, nErr := a.Srv().Store().Bot().Save(bot)
	if nErr != nil {
		a.Srv().Store().User().PermanentDelete(bot.UserId)
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateBot", "app.bot.createbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	// Get the owner of the bot, if one exists. If not, don't send a message
	ownerUser, err := a.Srv().Store().User().Get(context.Background(), bot.OwnerId)
	var nfErr *store.ErrNotFound
	if err != nil && !errors.As(err, &nfErr) {
		return nil, model.NewAppError("CreateBot", "app.user.get.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	} else if ownerUser != nil {
		// Send a message to the bot's creator to inform them that the bot needs to be added
		// to a team and channel after it's created
		botOwner, err := a.GetUser(bot.OwnerId)
		if err != nil {
			return nil, err
		}
		channel, err := a.getOrCreateDirectChannelWithUser(c, user, botOwner)
		if err != nil {
			return nil, err
		}

		T := i18n.GetUserTranslations(ownerUser.Locale)
		botAddPost := &model.Post{
			Type:      model.PostTypeAddBotTeamsChannels,
			UserId:    savedBot.UserId,
			ChannelId: channel.Id,
			Message:   T("api.bot.teams_channels.add_message_mobile"),
		}

		if _, err := a.CreatePostAsUser(c, botAddPost, c.Session().Id, true); err != nil {
			return nil, err
		}
	}

	return savedBot, nil
}

func (a *App) GetWarnMetricsBot() (*model.Bot, *model.AppError) {
	perPage := 1
	userOptions := &model.UserGetOptions{
		Page:     0,
		PerPage:  perPage,
		Role:     model.SystemAdminRoleId,
		Inactive: false,
	}

	sysAdminList, err := a.GetUsersFromProfiles(userOptions)
	if err != nil {
		return nil, err
	}

	if len(sysAdminList) == 0 {
		return nil, model.NewAppError("GetWarnMetricsBot", "app.bot.get_warn_metrics_bot.empty_admin_list.app_error", nil, "", http.StatusInternalServerError)
	}

	T := i18n.GetUserTranslations(sysAdminList[0].Locale)
	warnMetricsBot := &model.Bot{
		Username:    model.BotWarnMetricBotUsername,
		DisplayName: T("app.system.warn_metric.bot_displayname"),
		Description: "",
		OwnerId:     sysAdminList[0].Id,
	}

	return a.getOrCreateBot(warnMetricsBot)
}

func (a *App) GetSystemBot() (*model.Bot, *model.AppError) {
	perPage := 1
	userOptions := &model.UserGetOptions{
		Page:     0,
		PerPage:  perPage,
		Role:     model.SystemAdminRoleId,
		Inactive: false,
	}

	sysAdminList, err := a.GetUsersFromProfiles(userOptions)
	if err != nil {
		return nil, err
	}

	if len(sysAdminList) == 0 {
		return nil, model.NewAppError("GetSystemBot", "app.bot.get_system_bot.empty_admin_list.app_error", nil, "", http.StatusInternalServerError)
	}

	T := i18n.GetUserTranslations(sysAdminList[0].Locale)
	systemBot := &model.Bot{
		Username:    model.BotSystemBotUsername,
		DisplayName: T("app.system.system_bot.bot_displayname"),
		Description: "",
		OwnerId:     sysAdminList[0].Id,
	}

	return a.getOrCreateBot(systemBot)
}

func (a *App) getOrCreateBot(botDef *model.Bot) (*model.Bot, *model.AppError) {
	botUser, appErr := a.GetUserByUsername(botDef.Username)
	if appErr != nil {
		if appErr.StatusCode != http.StatusNotFound {
			return nil, appErr
		}

		// cannot find this bot user, save the user
		user, nErr := a.Srv().Store().User().Save(model.UserFromBot(botDef))
		if nErr != nil {
			var appError *model.AppError
			var invErr *store.ErrInvalidInput
			switch {
			case errors.As(nErr, &appError):
				return nil, appError
			case errors.As(nErr, &invErr):
				code := ""
				switch invErr.Field {
				case "email":
					code = "app.user.save.email_exists.app_error"
				case "username":
					code = "app.user.save.username_exists.app_error"
				default:
					code = "app.user.save.existing.app_error"
				}
				return nil, model.NewAppError("getOrCreateBot", code, nil, "", http.StatusBadRequest).Wrap(nErr)
			default:
				return nil, model.NewAppError("getOrCreateBot", "app.user.save.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
			}
		}
		botDef.UserId = user.Id

		//save the bot
		savedBot, nErr := a.Srv().Store().Bot().Save(botDef)
		if nErr != nil {
			a.Srv().Store().User().PermanentDelete(savedBot.UserId)
			var nAppErr *model.AppError
			switch {
			case errors.As(nErr, &nAppErr): // in case we haven't converted to plain error.
				return nil, nAppErr
			default: // last fallback in case it doesn't map to an existing app error.
				return nil, model.NewAppError("getOrCreateBot", "app.bot.createbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
			}
		}
		return savedBot, nil
	}

	if botUser == nil {
		return nil, model.NewAppError("getOrCreateBot", "app.bot.createbot.internal_error", nil, "", http.StatusInternalServerError)
	}

	//return the bot for this user
	savedBot, appErr := a.GetBot(botUser.Id, false)
	if appErr != nil {
		return nil, appErr
	}

	return savedBot, nil
}

// PatchBot applies the given patch to the bot and corresponding user.
func (a *App) PatchBot(botUserId string, botPatch *model.BotPatch) (*model.Bot, *model.AppError) {
	bot, err := a.GetBot(botUserId, true)
	if err != nil {
		return nil, err
	}

	if !bot.WouldPatch(botPatch) {
		return bot, nil
	}

	bot.Patch(botPatch)

	user, nErr := a.Srv().Store().User().Get(context.Background(), botUserId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("PatchBot", MissingAccountError, nil, "", http.StatusNotFound).Wrap(nErr)
		default:
			return nil, model.NewAppError("PatchBot", "app.user.get.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	patchedUser := model.UserFromBot(bot)
	user.Id = patchedUser.Id
	user.Username = patchedUser.Username
	user.Email = patchedUser.Email
	user.FirstName = patchedUser.FirstName

	userUpdate, nErr := a.Srv().Store().User().Update(user, true)
	if nErr != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		var conErr *store.ErrConflict
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &invErr):
			return nil, model.NewAppError("PatchBot", "app.user.update.find.app_error", nil, "", http.StatusBadRequest).Wrap(nErr)
		case errors.As(nErr, &conErr):
			if conErr.Resource == "Username" {
				return nil, model.NewAppError("PatchBot", "app.user.save.username_exists.app_error", nil, "", http.StatusBadRequest).Wrap(nErr)
			}
			return nil, model.NewAppError("PatchBot", "app.user.save.email_exists.app_error", nil, "", http.StatusBadRequest).Wrap(nErr)
		default:
			return nil, model.NewAppError("PatchBot", "app.user.update.finding.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}
	a.InvalidateCacheForUser(user.Id)

	ruser := userUpdate.New
	a.sendUpdatedUserEvent(*ruser)

	bot, nErr = a.Srv().Store().Bot().Update(bot)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(nErr)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("PatchBot", "app.bot.patchbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}
	return bot, nil
}

// GetBot returns the given bot.
func (a *App) GetBot(botUserId string, includeDeleted bool) (*model.Bot, *model.AppError) {
	bot, err := a.Srv().Store().Bot().Get(botUserId, includeDeleted)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(err)
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("GetBot", "app.bot.getbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}
	return bot, nil
}

// GetBots returns the requested page of bots.
func (a *App) GetBots(options *model.BotGetOptions) (model.BotList, *model.AppError) {
	bots, err := a.Srv().Store().Bot().GetAll(options)
	if err != nil {
		return nil, model.NewAppError("GetBots", "app.bot.getbots.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return bots, nil
}

// UpdateBotActive marks a bot as active or inactive, along with its corresponding user.
func (a *App) UpdateBotActive(c request.CTX, botUserId string, active bool) (*model.Bot, *model.AppError) {
	user, nErr := a.Srv().Store().User().Get(context.Background(), botUserId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("PatchBot", MissingAccountError, nil, "", http.StatusNotFound).Wrap(nErr)
		default:
			return nil, model.NewAppError("PatchBot", "app.user.get.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	if _, err := a.UpdateActive(c, user, active); err != nil {
		return nil, err
	}

	bot, nErr := a.Srv().Store().Bot().Get(botUserId, true)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(nErr)
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("UpdateBotActive", "app.bot.getbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	changed := true
	if active && bot.DeleteAt != 0 {
		bot.DeleteAt = 0
	} else if !active && bot.DeleteAt == 0 {
		bot.DeleteAt = model.GetMillis()
	} else {
		changed = false
	}

	if changed {
		bot, nErr = a.Srv().Store().Bot().Update(bot)
		if nErr != nil {
			var nfErr *store.ErrNotFound
			var appErr *model.AppError
			switch {
			case errors.As(nErr, &nfErr):
				return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(nErr)
			case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
				return nil, appErr
			default: // last fallback in case it doesn't map to an existing app error.
				return nil, model.NewAppError("PatchBot", "app.bot.patchbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
			}
		}
	}

	return bot, nil
}

// PermanentDeleteBot permanently deletes a bot and its corresponding user.
func (a *App) PermanentDeleteBot(botUserId string) *model.AppError {
	if err := a.Srv().Store().Bot().PermanentDelete(botUserId); err != nil {
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return model.NewAppError("PermanentDeleteBot", "app.bot.permenent_delete.bad_id", map[string]any{"user_id": invErr.Value}, "", http.StatusBadRequest).Wrap(err)
		default: // last fallback in case it doesn't map to an existing app error.
			return model.NewAppError("PatchBot", "app.bot.permanent_delete.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}

	if err := a.Srv().Store().User().PermanentDelete(botUserId); err != nil {
		return model.NewAppError("PermanentDeleteBot", "app.user.permanent_delete.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	return nil
}

// UpdateBotOwner changes a bot's owner to the given value.
func (a *App) UpdateBotOwner(botUserId, newOwnerId string) (*model.Bot, *model.AppError) {
	bot, err := a.Srv().Store().Bot().Get(botUserId, true)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(err)
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("UpdateBotOwner", "app.bot.getbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}

	bot.OwnerId = newOwnerId

	bot, err = a.Srv().Store().Bot().Update(bot)
	if err != nil {
		var nfErr *store.ErrNotFound
		var appErr *model.AppError
		switch {
		case errors.As(err, &nfErr):
			return nil, model.MakeBotNotFoundError(nfErr.ID).Wrap(err)
		case errors.As(err, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("PatchBot", "app.bot.patchbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}

	return bot, nil
}

// disableUserBots disables all bots owned by the given user.
func (a *App) disableUserBots(c request.CTX, userID string) *model.AppError {
	perPage := 20
	for {
		options := &model.BotGetOptions{
			OwnerId:        userID,
			IncludeDeleted: false,
			OnlyOrphaned:   false,
			Page:           0,
			PerPage:        perPage,
		}
		userBots, err := a.GetBots(options)
		if err != nil {
			return err
		}

		for _, bot := range userBots {
			_, err := a.UpdateBotActive(c, bot.UserId, false)
			if err != nil {
				c.Logger().Warn("Unable to deactivate bot.", mlog.String("bot_user_id", bot.UserId), mlog.Err(err))
			}
		}

		// Get next set of bots if we got the max number of bots
		if len(userBots) == perPage {
			options.Page += 1
			continue
		}
		break
	}

	return nil
}

func (a *App) notifySysadminsBotOwnerDeactivated(c request.CTX, userID string) *model.AppError {
	perPage := 25
	botOptions := &model.BotGetOptions{
		OwnerId:        userID,
		IncludeDeleted: false,
		OnlyOrphaned:   false,
		Page:           0,
		PerPage:        perPage,
	}
	// get owner bots
	var userBots []*model.Bot
	for {
		bots, err := a.GetBots(botOptions)
		if err != nil {
			return err
		}

		userBots = append(userBots, bots...)

		if len(bots) < perPage {
			break
		}

		botOptions.Page += 1
	}

	// user does not own bots
	if len(userBots) == 0 {
		return nil
	}

	userOptions := &model.UserGetOptions{
		Page:     0,
		PerPage:  perPage,
		Role:     model.SystemAdminRoleId,
		Inactive: false,
	}
	// get sysadmins
	var sysAdmins []*model.User
	for {
		sysAdminsList, err := a.GetUsersFromProfiles(userOptions)
		if err != nil {
			return err
		}

		sysAdmins = append(sysAdmins, sysAdminsList...)

		if len(sysAdminsList) < perPage {
			break
		}

		userOptions.Page += 1
	}

	// user being disabled
	user, err := a.GetUser(userID)
	if err != nil {
		return err
	}

	// for each sysadmin, notify user that owns bots was disabled
	for _, sysAdmin := range sysAdmins {
		channel, appErr := a.GetOrCreateDirectChannel(c, sysAdmin.Id, sysAdmin.Id)
		if appErr != nil {
			return appErr
		}

		post := &model.Post{
			UserId:    sysAdmin.Id,
			ChannelId: channel.Id,
			Message:   a.getDisableBotSysadminMessage(user, userBots),
			Type:      model.PostTypeSystemGeneric,
		}

		_, appErr = a.CreatePost(c, post, channel, false, true)
		if appErr != nil {
			return appErr
		}
	}
	return nil
}

func (a *App) getDisableBotSysadminMessage(user *model.User, userBots model.BotList) string {
	disableBotsSetting := *a.Config().ServiceSettings.DisableBotsWhenOwnerIsDeactivated

	var printAllBots = true
	numBotsToPrint := len(userBots)

	if numBotsToPrint > 10 {
		numBotsToPrint = 10
		printAllBots = false
	}

	var message, botList string
	for _, bot := range userBots[:numBotsToPrint] {
		botList += fmt.Sprintf("* %v\n", bot.Username)
	}

	T := i18n.GetUserTranslations(user.Locale)
	message = T("app.bot.get_disable_bot_sysadmin_message",
		map[string]any{
			"UserName":           user.Username,
			"NumBots":            len(userBots),
			"BotNames":           botList,
			"disableBotsSetting": disableBotsSetting,
			"printAllBots":       printAllBots,
		})

	return message
}

// ConvertUserToBot converts a user to bot.
func (a *App) ConvertUserToBot(user *model.User) (*model.Bot, *model.AppError) {
	bot, err := a.Srv().Store().Bot().Save(model.BotFromUser(user))
	if err != nil {
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateBot", "app.bot.createbot.internal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}
	return bot, nil
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"
)

type ensureBotOptions struct {
	ProfileImagePath string
	IconImagePath    string
}

type EnsureBotOption func(*ensureBotOptions)

func ProfileImagePath(path string) EnsureBotOption {
	return func(args *ensureBotOptions) {
		args.ProfileImagePath = path
	}
}

func IconImagePath(path string) EnsureBotOption {
	return func(args *ensureBotOptions) {
		args.IconImagePath = path
	}
}

// EnsureBot implements Helpers.EnsureBot
func (p *HelpersImpl) EnsureBot(bot *model.Bot, options ...EnsureBotOption) (retBotID string, retErr error) {
	err := p.ensureServerVersion("5.10.0")
	if err != nil {
		return "", errors.Wrap(err, "failed to ensure bot")
	}

	// Default options
	o := &ensureBotOptions{
		ProfileImagePath: "",
		IconImagePath:    "",
	}

	for _, setter := range options {
		setter(o)
	}

	botID, err := p.ensureBot(bot)
	if err != nil {
		return "", err
	}

	err = p.setBotImages(botID, o.ProfileImagePath, o.IconImagePath)
	if err != nil {
		return "", err
	}
	return botID, nil
}

type ShouldProcessMessageOption func(*shouldProcessMessageOptions)

type shouldProcessMessageOptions struct {
	AllowSystemMessages bool
	AllowBots           bool
	AllowWebhook        bool
	FilterChannelIDs    []string
	FilterUserIDs       []string
	OnlyBotDMs          bool
}

// AllowSystemMessages configures a call to ShouldProcessMessage to return true for system messages.
//
// As it is typically desirable only to consume messages from users of the system, ShouldProcessMessage ignores system messages by default.
func AllowSystemMessages() ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.AllowSystemMessages = true
	}
}

// AllowBots configures a call to ShouldProcessMessage to return true for bot posts.
//
// As it is typically desirable only to consume messages from human users of the system, ShouldProcessMessage ignores bot messages by default. When allowing bots, take care to avoid a loop where two plugins respond to each others posts repeatedly.
func AllowBots() ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.AllowBots = true
	}
}

// AllowWebhook configures a call to ShouldProcessMessage to return true for posts from webhook.
//
// As it is typically desirable only to consume messages from human users of the system, ShouldProcessMessage ignores webhook messages by default.
func AllowWebhook() ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.AllowWebhook = true
	}
}

// FilterChannelIDs configures a call to ShouldProcessMessage to return true only for the given channels.
//
// By default, posts from all channels are allowed to be processed.
func FilterChannelIDs(filterChannelIDs []string) ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.FilterChannelIDs = filterChannelIDs
	}
}

// FilterUserIDs configures a call to ShouldProcessMessage to return true only for the given users.
//
// By default, posts from all non-bot users are allowed.
func FilterUserIDs(filterUserIDs []string) ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.FilterUserIDs = filterUserIDs
	}
}

// OnlyBotDMs configures a call to ShouldProcessMessage to return true only for direct messages sent to the bot created by EnsureBot.
//
// By default, posts from all channels are allowed.
func OnlyBotDMs() ShouldProcessMessageOption {
	return func(options *shouldProcessMessageOptions) {
		options.OnlyBotDMs = true
	}
}

// ShouldProcessMessage implements Helpers.ShouldProcessMessage
func (p *HelpersImpl) ShouldProcessMessage(post *model.Post, options ...ShouldProcessMessageOption) (bool, error) {
	messageProcessOptions := &shouldProcessMessageOptions{}
	for _, option := range options {
		option(messageProcessOptions)
	}

	botIDBytes, kvGetErr := p.API.KVGet(BOT_USER_KEY)
	if kvGetErr != nil {
		return false, errors.Wrap(kvGetErr, "failed to get bot")
	}

	if botIDBytes != nil {
		if post.UserId == string(botIDBytes) {
			return false, nil
		}
	}

	if post.IsSystemMessage() && !messageProcessOptions.AllowSystemMessages {
		return false, nil
	}

	if !messageProcessOptions.AllowWebhook && post.GetProp("from_webhook") == "true" {
		return false, nil
	}

	if !messageProcessOptions.AllowBots {
		user, appErr := p.API.GetUser(post.UserId)
		if appErr != nil {
			return false, errors.Wrap(appErr, "unable to get user")
		}

		if user.IsBot {
			return false, nil
		}
	}

	if len(messageProcessOptions.FilterChannelIDs) != 0 && !utils.StringInSlice(post.ChannelId, messageProcessOptions.FilterChannelIDs) {
		return false, nil
	}

	if len(messageProcessOptions.FilterUserIDs) != 0 && !utils.StringInSlice(post.UserId, messageProcessOptions.FilterUserIDs) {
		return false, nil
	}

	if botIDBytes != nil && messageProcessOptions.OnlyBotDMs {
		channel, appErr := p.API.GetChannel(post.ChannelId)
		if appErr != nil {
			return false, errors.Wrap(appErr, "unable to get channel")
		}

		if !model.IsBotDMChannel(channel, string(botIDBytes)) {
			return false, nil
		}
	}

	return true, nil
}

func (p *HelpersImpl) readFile(path string) ([]byte, error) {
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bundle path")
	}

	imageBytes, err := ioutil.ReadFile(filepath.Join(bundlePath, path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read image")
	}
	return imageBytes, nil
}

func (p *HelpersImpl) ensureBot(bot *model.Bot) (retBotID string, retErr error) {
	// Must provide a bot with a username
	if bot == nil || len(bot.Username) < 1 {
		return "", errors.New("passed a bad bot, nil or no username")
	}

	// If we fail for any reason, this could be a race between creation of bot and
	// retrieval from another EnsureBot. Just try the basic retrieve existing again.
	defer func() {
		if retBotID == "" || retErr != nil {
			var err error
			var botIDBytes []byte

			err = utils.ProgressiveRetry(func() error {
				botIDBytes, err = p.API.KVGet(BOT_USER_KEY)
				if err != nil {
					return err
				}
				return nil
			})

			if err == nil && botIDBytes != nil {
				retBotID = string(botIDBytes)
				retErr = nil
			}
		}
	}()

	botIDBytes, kvGetErr := p.API.KVGet(BOT_USER_KEY)
	if kvGetErr != nil {
		return "", errors.Wrap(kvGetErr, "failed to get bot")
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

		if _, err := p.API.PatchBot(botID, botPatch); err != nil {
			return "", errors.Wrap(err, "failed to patch bot")
		}

		return botID, nil
	}

	// Check for an existing bot user with that username. If one exists, then use that.
	if user, userGetErr := p.API.GetUserByUsername(bot.Username); userGetErr == nil && user != nil {
		if user.IsBot {
			if kvSetErr := p.API.KVSet(BOT_USER_KEY, []byte(user.Id)); kvSetErr != nil {
				p.API.LogWarn("Failed to set claimed bot user id.", "userid", user.Id, "err", kvSetErr)
			}
		} else {
			p.API.LogError("Plugin attempted to use an account that already exists. Convert user to a bot account in the CLI by running 'mattermost user convert <username> --bot'. If the user is an existing user account you want to preserve, change its username and restart the Mattermost server, after which the plugin will create a bot account with that name. For more information about bot accounts, see https://mattermost.com/pl/default-bot-accounts", "username", bot.Username, "user_id", user.Id)
		}
		return user.Id, nil
	}

	// Create a new bot user for the plugin
	createdBot, createBotErr := p.API.CreateBot(bot)
	if createBotErr != nil {
		return "", errors.Wrap(createBotErr, "failed to create bot")
	}

	if kvSetErr := p.API.KVSet(BOT_USER_KEY, []byte(createdBot.UserId)); kvSetErr != nil {
		p.API.LogWarn("Failed to set created bot user id.", "userid", createdBot.UserId, "err", kvSetErr)
	}

	return createdBot.UserId, nil
}

func (p *HelpersImpl) setBotImages(botID, profileImagePath, iconImagePath string) error {
	if profileImagePath != "" {
		imageBytes, err := p.readFile(profileImagePath)
		if err != nil {
			return errors.Wrap(err, "failed to read profile image")
		}
		appErr := p.API.SetProfileImage(botID, imageBytes)
		if appErr != nil {
			return errors.Wrap(appErr, "failed to set profile image")
		}
	}
	if iconImagePath != "" {
		imageBytes, err := p.readFile(iconImagePath)
		if err != nil {
			return errors.Wrap(err, "failed to read icon image")
		}
		appErr := p.API.SetBotIconImage(botID, imageBytes)
		if appErr != nil {
			return errors.Wrap(appErr, "failed to set icon image")
		}
	}
	return nil
}

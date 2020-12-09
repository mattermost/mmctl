// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"

	"github.com/pkg/errors"
)

// bot is a subset of the model.Bot type, omitting the model.User fields.
type bot struct {
	UserId         string `json:"user_id"`
	Description    string `json:"description"`
	OwnerId        string `json:"owner_id"`
	LastIconUpdate int64  `json:"last_icon_update"`
	CreateAt       int64  `json:"create_at"`
	UpdateAt       int64  `json:"update_at"`
	DeleteAt       int64  `json:"delete_at"`
}

func botFromModel(b *model.Bot) *bot {
	return &bot{
		UserId:         b.UserId,
		Description:    b.Description,
		OwnerId:        b.OwnerId,
		LastIconUpdate: b.LastIconUpdate,
		CreateAt:       b.CreateAt,
		UpdateAt:       b.UpdateAt,
		DeleteAt:       b.DeleteAt,
	}
}

// SqlBotStore is a store for managing bots in the database.
// Bots are otherwise normal users with extra metadata record in the Bots table. The primary key
// for a bot matches the primary key value for corresponding User record.
type SqlBotStore struct {
	*SqlStore
	metrics einterfaces.MetricsInterface
}

// newSqlBotStore creates an instance of SqlBotStore, registering the table schema in question.
func newSqlBotStore(sqlStore *SqlStore, metrics einterfaces.MetricsInterface) store.BotStore {
	us := &SqlBotStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(bot{}, "Bots").SetKeys(false, "UserId")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Description").SetMaxSize(1024)
		table.ColMap("OwnerId").SetMaxSize(model.BOT_CREATOR_ID_MAX_RUNES)
	}

	return us
}

func (us SqlBotStore) createIndexesIfNotExists() {
}

// Get fetches the given bot in the database.
func (us SqlBotStore) Get(botUserId string, includeDeleted bool) (*model.Bot, error) {
	var excludeDeletedSql = "AND b.DeleteAt = 0"
	if includeDeleted {
		excludeDeletedSql = ""
	}

	query := `
		SELECT
			b.UserId,
			u.Username,
			u.FirstName AS DisplayName,
			b.Description,
			b.OwnerId,
			COALESCE(b.LastIconUpdate, 0) AS LastIconUpdate,
			b.CreateAt,
			b.UpdateAt,
			b.DeleteAt
		FROM
			Bots b
		JOIN
			Users u ON (u.Id = b.UserId)
		WHERE
			b.UserId = :user_id
			` + excludeDeletedSql + `
	`

	var bot *model.Bot
	if err := us.GetReplica().SelectOne(&bot, query, map[string]interface{}{"user_id": botUserId}); err == sql.ErrNoRows {
		return nil, store.NewErrNotFound("Bot", botUserId)
	} else if err != nil {
		return nil, errors.Wrapf(err, "selectone: user_id=%s", botUserId)
	}

	return bot, nil
}

// GetAll fetches from all bots in the database.
func (us SqlBotStore) GetAll(options *model.BotGetOptions) ([]*model.Bot, error) {
	params := map[string]interface{}{
		"offset": options.Page * options.PerPage,
		"limit":  options.PerPage,
	}

	var conditions []string
	var conditionsSql string
	var additionalJoin string

	if !options.IncludeDeleted {
		conditions = append(conditions, "b.DeleteAt = 0")
	}
	if options.OwnerId != "" {
		conditions = append(conditions, "b.OwnerId = :creator_id")
		params["creator_id"] = options.OwnerId
	}
	if options.OnlyOrphaned {
		additionalJoin = "JOIN Users o ON (o.Id = b.OwnerId)"
		conditions = append(conditions, "o.DeleteAt != 0")
	}

	if len(conditions) > 0 {
		conditionsSql = "WHERE " + strings.Join(conditions, " AND ")
	}

	sql := `
			SELECT
			    b.UserId,
			    u.Username,
			    u.FirstName AS DisplayName,
			    b.Description,
			    b.OwnerId,
			    COALESCE(b.LastIconUpdate, 0) AS LastIconUpdate,
			    b.CreateAt,
			    b.UpdateAt,
			    b.DeleteAt
			FROM
			    Bots b
			JOIN
			    Users u ON (u.Id = b.UserId)
			` + additionalJoin + `
			` + conditionsSql + `
			ORDER BY
			    b.CreateAt ASC,
			    u.Username ASC
			LIMIT
			    :limit
			OFFSET
			    :offset
		`

	var bots []*model.Bot
	if _, err := us.GetReplica().Select(&bots, sql, params); err != nil {
		return nil, errors.Wrap(err, "select")
	}

	return bots, nil
}

// Save persists a new bot to the database.
// It assumes the corresponding user was saved via the user store.
func (us SqlBotStore) Save(bot *model.Bot) (*model.Bot, error) {
	bot = bot.Clone()
	bot.PreSave()

	if err := bot.IsValid(); err != nil { // TODO: change to return error in v6.
		return nil, err
	}

	if err := us.GetMaster().Insert(botFromModel(bot)); err != nil {
		return nil, errors.Wrapf(err, "insert: user_id=%s", bot.UserId)
	}

	return bot, nil
}

// Update persists an updated bot to the database.
// It assumes the corresponding user was updated via the user store.
func (us SqlBotStore) Update(bot *model.Bot) (*model.Bot, error) {
	bot = bot.Clone()

	bot.PreUpdate()
	if err := bot.IsValid(); err != nil { // TODO: needs to return error in v6
		return nil, err
	}

	oldBot, err := us.Get(bot.UserId, true)
	if err != nil {
		return nil, err
	}

	oldBot.Description = bot.Description
	oldBot.OwnerId = bot.OwnerId
	oldBot.LastIconUpdate = bot.LastIconUpdate
	oldBot.UpdateAt = bot.UpdateAt
	oldBot.DeleteAt = bot.DeleteAt
	bot = oldBot

	if count, err := us.GetMaster().Update(botFromModel(bot)); err != nil {
		return nil, errors.Wrapf(err, "update: user_id=%s", bot.UserId)
	} else if count > 1 {
		return nil, fmt.Errorf("unexpected count while updating bot: count=%d, userId=%s", count, bot.UserId)
	}

	return bot, nil
}

// PermanentDelete removes the bot from the database altogether.
// If the corresponding user is to be deleted, it must be done via the user store.
func (us SqlBotStore) PermanentDelete(botUserId string) error {
	query := "DELETE FROM Bots WHERE UserId = :user_id"
	if _, err := us.GetMaster().Exec(query, map[string]interface{}{"user_id": botUserId}); err != nil {
		return store.NewErrInvalidInput("Bot", "UserId", botUserId)
	}
	return nil
}

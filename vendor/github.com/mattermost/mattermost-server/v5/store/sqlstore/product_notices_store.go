// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlProductNoticesStore struct {
	*SqlStore
}

func newSqlProductNoticesStore(sqlStore *SqlStore) store.ProductNoticesStore {
	s := SqlProductNoticesStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ProductNoticeViewState{}, "ProductNoticeViewState").SetKeys(false, "UserId", "NoticeId")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("NoticeId").SetMaxSize(26)
	}

	return s
}

func (s SqlProductNoticesStore) createIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_notice_views_timestamp", "ProductNoticeViewState", "Timestamp")
	s.CreateIndexIfNotExists("idx_notice_views_user_id", "ProductNoticeViewState", "UserId")
	s.CreateIndexIfNotExists("idx_notice_views_notice_id", "ProductNoticeViewState", "NoticeId")

	s.CreateCompositeIndexIfNotExists("idx_notice_views_user_notice", "ProductNoticeViewState", []string{"UserId", "NoticeId"})

}

func (s SqlProductNoticesStore) Clear(notices []string) error {
	sql, args, _ := s.getQueryBuilder().Delete("ProductNoticeViewState").Where(sq.Eq{"NoticeId": notices}).ToSql()
	if _, err := s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete records from ProductNoticeViewState")
	}
	return nil
}

func (s SqlProductNoticesStore) ClearOldNotices(currentNotices *model.ProductNotices) error {
	var notices []string
	for _, currentNotice := range *currentNotices {
		notices = append(notices, currentNotice.ID)
	}
	sql, args, _ := s.getQueryBuilder().Delete("ProductNoticeViewState").Where(sq.NotEq{"NoticeId": notices}).ToSql()
	if _, err := s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete records from ProductNoticeViewState")
	}
	return nil
}

func (s SqlProductNoticesStore) View(userId string, notices []string) error {
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return errors.Wrap(err, "begin_transaction")
	}
	defer finalizeTransaction(transaction)

	var noticeStates []model.ProductNoticeViewState
	sql, args, _ := s.getQueryBuilder().
		Select("*").
		From("ProductNoticeViewState").
		Where(sq.And{sq.Eq{"UserId": userId}, sq.Eq{"NoticeId": notices}}).
		ToSql()
	if _, err := transaction.Select(&noticeStates, sql, args...); err != nil {
		return errors.Wrapf(err, "failed to get ProductNoticeViewState with userId=%s", userId)
	}

	now := time.Now().UTC().Unix()

	// update existing records
	for i := range noticeStates {
		noticeStates[i].Viewed += 1
		noticeStates[i].Timestamp = now
		if _, err := transaction.Update(&noticeStates[i]); err != nil {
			return errors.Wrapf(err, "failed to update ProductNoticeViewState")
		}
	}

	// add new ones
	haveNoticeState := func(n string) bool {
		for _, ns := range noticeStates {
			if ns.NoticeId == n {
				return true
			}
		}
		return false
	}

	for _, noticeId := range notices {
		if !haveNoticeState(noticeId) {
			if err := transaction.Insert(&model.ProductNoticeViewState{
				UserId:    userId,
				NoticeId:  noticeId,
				Viewed:    1,
				Timestamp: now,
			}); err != nil {
				return errors.Wrapf(err, "failed to insert ProductNoticeViewState")
			}
		}
	}

	if err := transaction.Commit(); err != nil {
		return errors.Wrap(err, "commit_transaction")
	}

	return nil
}

func (s SqlProductNoticesStore) GetViews(userId string) ([]model.ProductNoticeViewState, error) {
	var noticeStates []model.ProductNoticeViewState
	sql, args, _ := s.getQueryBuilder().Select("*").From("ProductNoticeViewState").Where(sq.Eq{"UserId": userId}).ToSql()
	if _, err := s.GetReplica().Select(&noticeStates, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to get ProductNoticeViewState with userId=%s", userId)
	}
	return noticeStates, nil
}

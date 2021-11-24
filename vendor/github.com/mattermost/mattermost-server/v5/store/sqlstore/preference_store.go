// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/pkg/errors"
)

type SqlPreferenceStore struct {
	*SqlStore
}

func newSqlPreferenceStore(sqlStore *SqlStore) store.PreferenceStore {
	s := &SqlPreferenceStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Preference{}, "Preferences").SetKeys(false, "UserId", "Category", "Name")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Category").SetMaxSize(32)
		table.ColMap("Name").SetMaxSize(32)
		table.ColMap("Value").SetMaxSize(2000)
	}

	return s
}

func (s SqlPreferenceStore) createIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_preferences_category", "Preferences", "Category")
	s.CreateIndexIfNotExists("idx_preferences_name", "Preferences", "Name")
}

func (s SqlPreferenceStore) deleteUnusedFeatures() {
	mlog.Debug("Deleting any unused pre-release features")
	sql, args, err := s.getQueryBuilder().
		Delete("Preferences").
		Where(sq.Eq{"Category": model.PREFERENCE_CATEGORY_ADVANCED_SETTINGS}).
		Where(sq.Eq{"Value": "false"}).
		Where(sq.Like{"Name": store.FeatureTogglePrefix + "%"}).ToSql()
	if err != nil {
		mlog.Warn(errors.Wrap(err, "could not build sql query to delete unused features!").Error())
	}
	if _, err = s.GetMaster().Exec(sql, args...); err != nil {
		mlog.Warn("Failed to delete unused features", mlog.Err(err))
	}
}

func (s SqlPreferenceStore) Save(preferences *model.Preferences) error {
	// wrap in a transaction so that if one fails, everything fails
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return errors.Wrap(err, "begin_transaction")
	}

	defer finalizeTransaction(transaction)
	for _, preference := range *preferences {
		preference := preference
		if upsertErr := s.save(transaction, &preference); upsertErr != nil {
			return upsertErr
		}
	}

	if err := transaction.Commit(); err != nil {
		// don't need to rollback here since the transaction is already closed
		return errors.Wrap(err, "commit_transaction")
	}
	return nil
}

func (s SqlPreferenceStore) save(transaction *gorp.Transaction, preference *model.Preference) error {
	preference.PreUpdate()

	if err := preference.IsValid(); err != nil {
		return err
	}

	query := s.getQueryBuilder().
		Insert("Preferences").
		Columns("UserId", "Category", "Name", "Value").
		Values(preference.UserId, preference.Category, preference.Name, preference.Value)

	if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		query = query.SuffixExpr(sq.Expr("ON DUPLICATE KEY UPDATE Value = ?", preference.Value))
	} else if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		query = query.SuffixExpr(sq.Expr("ON CONFLICT (userid, category, name) DO UPDATE SET Value = ?", preference.Value))
	} else {
		return store.NewErrNotImplemented("failed to update preference because of missing driver")
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to generate sqlquery")
	}

	if _, err = transaction.Exec(queryString, args...); err != nil {
		return errors.Wrap(err, "failed to save Preference")
	}
	return nil
}

func (s SqlPreferenceStore) Get(userId string, category string, name string) (*model.Preference, error) {
	var preference *model.Preference
	query, args, err := s.getQueryBuilder().
		Select("*").
		From("Preferences").
		Where(sq.Eq{"UserId": userId}).
		Where(sq.Eq{"Category": category}).
		Where(sq.Eq{"Name": name}).
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err, "could not build sql query to get preference")
	}
	if err = s.GetReplica().SelectOne(&preference, query, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to find Preference with userId=%s, category=%s, name=%s", userId, category, name)
	}

	return preference, nil
}

func (s SqlPreferenceStore) GetCategory(userId string, category string) (model.Preferences, error) {
	var preferences model.Preferences
	query, args, err := s.getQueryBuilder().
		Select("*").
		From("Preferences").
		Where(sq.Eq{"UserId": userId}).
		Where(sq.Eq{"Category": category}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not build sql query to get preference")
	}
	if _, err = s.GetReplica().Select(&preferences, query, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to find Preference with userId=%s, category=%s", userId, category)
	}
	return preferences, nil

}

func (s SqlPreferenceStore) GetAll(userId string) (model.Preferences, error) {
	var preferences model.Preferences
	query, args, err := s.getQueryBuilder().
		Select("*").
		From("Preferences").
		Where(sq.Eq{"UserId": userId}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not build sql query to get preference")
	}
	if _, err = s.GetReplica().Select(&preferences, query, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to find Preference with userId=%s", userId)
	}
	return preferences, nil
}

func (s SqlPreferenceStore) PermanentDeleteByUser(userId string) error {
	sql, args, err := s.getQueryBuilder().
		Delete("Preferences").
		Where(sq.Eq{"UserId": userId}).ToSql()
	if err != nil {
		return errors.Wrap(err, "could not build sql query to get delete preference by user")
	}
	if _, err := s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete Preference with userId=%s", userId)
	}
	return nil
}

func (s SqlPreferenceStore) Delete(userId, category, name string) error {

	sql, args, err := s.getQueryBuilder().
		Delete("Preferences").
		Where(sq.Eq{"UserId": userId}).
		Where(sq.Eq{"Category": category}).
		Where(sq.Eq{"Name": name}).ToSql()

	if err != nil {
		return errors.Wrap(err, "could not build sql query to get delete preference")
	}

	if _, err = s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete Preference with userId=%s, category=%s and name=%s", userId, category, name)
	}

	return nil
}

func (s SqlPreferenceStore) DeleteCategory(userId string, category string) error {

	sql, args, err := s.getQueryBuilder().
		Delete("Preferences").
		Where(sq.Eq{"UserId": userId}).
		Where(sq.Eq{"Category": category}).ToSql()

	if err != nil {
		return errors.Wrap(err, "could not build sql query to get delete preference by category")
	}

	if _, err = s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete Preference with userId=%s and category=%s", userId, category)
	}

	return nil
}

func (s SqlPreferenceStore) DeleteCategoryAndName(category string, name string) error {
	sql, args, err := s.getQueryBuilder().
		Delete("Preferences").
		Where(sq.Eq{"Name": name}).
		Where(sq.Eq{"Category": category}).ToSql()

	if err != nil {
		return errors.Wrap(err, "could not build sql query to get delete preference by category and name")
	}

	if _, err = s.GetMaster().Exec(sql, args...); err != nil {
		return errors.Wrapf(err, "failed to delete Preference with category=%s and name=%s", category, name)
	}

	return nil
}

// DeleteOrphanedRows removes entries from Preferences (flagged post) when a
// corresponding post no longer exists.
func (s *SqlPreferenceStore) DeleteOrphanedRows(limit int) (deleted int64, err error) {
	// We need the extra level of nesting to deal with MySQL's locking
	const query = `
	DELETE FROM Preferences WHERE Name IN (
		SELECT * FROM (
			SELECT Preferences.Name FROM Preferences
			LEFT JOIN Posts ON Preferences.Name = Posts.Id
			WHERE Posts.Id IS NULL AND Category = :Category
			LIMIT :Limit
		) AS A
	)`
	props := map[string]interface{}{"Limit": limit, "Category": model.PREFERENCE_CATEGORY_FLAGGED_POST}
	result, err := s.GetMaster().Exec(query, props)
	if err != nil {
		return
	}
	deleted, err = result.RowsAffected()
	return
}

func (s SqlPreferenceStore) CleanupFlagsBatch(limit int64) (int64, error) {
	if limit < 0 {
		// uint64 does not throw an error, it overflows if it is negative.
		// it is better to manually check here, or change the function type to uint64
		return int64(0), errors.Errorf("Received a negative limit")
	}
	nameInQ, nameInArgs, err := sq.Select("*").
		FromSelect(
			sq.Select("Preferences.Name").
				From("Preferences").
				LeftJoin("Posts ON Preferences.Name = Posts.Id").
				Where(sq.Eq{"Preferences.Category": model.PREFERENCE_CATEGORY_FLAGGED_POST}).
				Where(sq.Eq{"Posts.Id": nil}).
				Limit(uint64(limit)),
			"t").
		ToSql()
	if err != nil {
		return int64(0), errors.Wrap(err, "could not build nested sql query to delete preference")
	}
	query, args, err := s.getQueryBuilder().Delete("Preferences").
		Where(sq.Eq{"Category": model.PREFERENCE_CATEGORY_FLAGGED_POST}).
		Where(sq.Expr("name IN ("+nameInQ+")", nameInArgs...)).
		ToSql()

	if err != nil {
		return int64(0), errors.Wrap(err, "could not build sql query to delete preference")
	}

	sqlResult, err := s.GetMaster().Exec(query, args...)
	if err != nil {
		return int64(0), errors.Wrap(err, "failed to delete Preference")
	}
	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return int64(0), errors.Wrap(err, "unable to get rows affected")
	}

	return rowsAffected, nil
}

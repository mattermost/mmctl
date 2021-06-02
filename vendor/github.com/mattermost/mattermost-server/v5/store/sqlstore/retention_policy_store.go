// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/pkg/errors"
)

type SqlRetentionPolicyStore struct {
	*SqlStore
	metrics einterfaces.MetricsInterface
}

func newSqlRetentionPolicyStore(sqlStore *SqlStore, metrics einterfaces.MetricsInterface) store.RetentionPolicyStore {
	s := &SqlRetentionPolicyStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.RetentionPolicy{}, "RetentionPolicies")
		table.SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("DisplayName").SetMaxSize(64)

		tableC := db.AddTableWithName(model.RetentionPolicyChannel{}, "RetentionPoliciesChannels")
		tableC.SetKeys(false, "ChannelId")
		tableC.ColMap("PolicyId").SetMaxSize(26)
		tableC.ColMap("ChannelId").SetMaxSize(26)

		tableT := db.AddTableWithName(model.RetentionPolicyTeam{}, "RetentionPoliciesTeams")
		tableT.SetKeys(false, "TeamId")
		tableT.ColMap("PolicyId").SetMaxSize(26)
		tableT.ColMap("TeamId").SetMaxSize(26)
	}

	return s
}

func (s *SqlRetentionPolicyStore) createIndexesIfNotExists() {
	s.CreateCompositeIndexIfNotExists("IDX_RetentionPolicies_DisplayName_Id", "RetentionPolicies",
		[]string{"DisplayName", "Id"})
	s.CreateIndexIfNotExists("IDX_RetentionPoliciesChannels_PolicyId", "RetentionPoliciesChannels", "PolicyId")
	s.CreateIndexIfNotExists("IDX_RetentionPoliciesTeams_PolicyId", "RetentionPoliciesTeams", "PolicyId")
	s.CreateForeignKeyIfNotExists("RetentionPoliciesChannels", "PolicyId", "RetentionPolicies", "Id", true)
	s.CreateForeignKeyIfNotExists("RetentionPoliciesTeams", "PolicyId", "RetentionPolicies", "Id", true)
}

// executePossiblyEmptyQuery only executes the query if it is non-empty. This helps avoid
// having to check for MySQL, which, unlike Postgres, does not allow empty queries.
func executePossiblyEmptyQuery(txn *gorp.Transaction, query string, args ...interface{}) (sql.Result, error) {
	if query == "" {
		return nil, nil
	}
	return txn.Exec(query, args...)
}

func (s *SqlRetentionPolicyStore) Save(policy *model.RetentionPolicyWithTeamAndChannelIDs) (*model.RetentionPolicyWithTeamAndChannelCounts, error) {
	// Strategy:
	// 1. Insert new policy
	// 2. Insert new channels into policy
	// 3. Insert new teams into policy

	if err := s.checkTeamsExist(policy.TeamIDs); err != nil {
		return nil, err
	}
	if err := s.checkChannelsExist(policy.ChannelIDs); err != nil {
		return nil, err
	}

	policy.ID = model.NewId()

	policyInsertQuery, policyInsertArgs, err := s.getQueryBuilder().
		Insert("RetentionPolicies").
		Columns("Id", "DisplayName", "PostDuration").
		Values(policy.ID, policy.DisplayName, policy.PostDuration).
		ToSql()
	if err != nil {
		return nil, err
	}

	channelsInsertQuery, channelsInsertArgs, err := s.buildInsertRetentionPoliciesChannelsQuery(policy.ID, policy.ChannelIDs)
	if err != nil {
		return nil, err
	}

	teamsInsertQuery, teamsInsertArgs, err := s.buildInsertRetentionPoliciesTeamsQuery(policy.ID, policy.TeamIDs)
	if err != nil {
		return nil, err
	}

	policySelectQuery, policySelectProps := s.buildGetPolicyQuery(policy.ID)

	txn, err := s.GetMaster().Begin()
	if err != nil {
		return nil, err
	}
	defer finalizeTransaction(txn)
	// Create a new policy in RetentionPolicies
	if _, err = txn.Exec(policyInsertQuery, policyInsertArgs...); err != nil {
		return nil, err
	}
	// Insert the channel IDs into RetentionPoliciesChannels
	if _, err = executePossiblyEmptyQuery(txn, channelsInsertQuery, channelsInsertArgs...); err != nil {
		return nil, err
	}
	// Insert the team IDs into RetentionPoliciesTeams
	if _, err = executePossiblyEmptyQuery(txn, teamsInsertQuery, teamsInsertArgs...); err != nil {
		return nil, err
	}
	// Select the new policy (with team/channel counts) which we just created
	var newPolicy model.RetentionPolicyWithTeamAndChannelCounts
	if err = txn.SelectOne(&newPolicy, policySelectQuery, policySelectProps); err != nil {
		return nil, err
	}
	if err = txn.Commit(); err != nil {
		return nil, err
	}
	return &newPolicy, nil
}

func (s *SqlRetentionPolicyStore) checkTeamsExist(teamIDs []string) error {
	if len(teamIDs) > 0 {
		teamsSelectQuery, teamsSelectArgs, err := s.getQueryBuilder().
			Select("Id").
			From("Teams").
			Where(sq.Eq{"Id": teamIDs}).
			ToSql()
		if err != nil {
			return err
		}
		var rows []*string
		_, err = s.GetReplica().Select(&rows, teamsSelectQuery, teamsSelectArgs...)
		if err != nil {
			return err
		}
		if len(rows) == len(teamIDs) {
			return nil
		}
		retrievedIDs := make(map[string]bool)
		for _, teamID := range rows {
			retrievedIDs[*teamID] = true
		}
		for _, teamID := range teamIDs {
			if _, ok := retrievedIDs[teamID]; !ok {
				return store.NewErrNotFound("Team", teamID)
			}
		}
	}
	return nil
}

func (s *SqlRetentionPolicyStore) checkChannelsExist(channelIDs []string) error {
	if len(channelIDs) > 0 {
		channelsSelectQuery, channelsSelectArgs, err := s.getQueryBuilder().
			Select("Id").
			From("Channels").
			Where(sq.Eq{"Id": channelIDs}).
			ToSql()
		if err != nil {
			return err
		}
		var rows []*string
		_, err = s.GetReplica().Select(&rows, channelsSelectQuery, channelsSelectArgs...)
		if err != nil {
			return err
		}
		if len(rows) == len(channelIDs) {
			return nil
		}
		retrievedIDs := make(map[string]bool)
		for _, channelID := range rows {
			retrievedIDs[*channelID] = true
		}
		for _, channelID := range channelIDs {
			if _, ok := retrievedIDs[channelID]; !ok {
				return store.NewErrNotFound("Channel", channelID)
			}
		}
	}
	return nil
}

func (s *SqlRetentionPolicyStore) buildInsertRetentionPoliciesChannelsQuery(policyID string, channelIDs []string) (query string, args []interface{}, err error) {
	if len(channelIDs) > 0 {
		builder := s.getQueryBuilder().
			Insert("RetentionPoliciesChannels").
			Columns("PolicyId", "ChannelId")
		for _, channelID := range channelIDs {
			builder = builder.Values(policyID, channelID)
		}
		query, args, err = builder.ToSql()
	}
	return
}

func (s *SqlRetentionPolicyStore) buildInsertRetentionPoliciesTeamsQuery(policyID string, teamIDs []string) (query string, args []interface{}, err error) {
	if len(teamIDs) > 0 {
		builder := s.getQueryBuilder().
			Insert("RetentionPoliciesTeams").
			Columns("PolicyId", "TeamId")
		for _, teamID := range teamIDs {
			builder = builder.Values(policyID, teamID)
		}
		query, args, err = builder.ToSql()
	}
	return
}

func (s *SqlRetentionPolicyStore) Patch(patch *model.RetentionPolicyWithTeamAndChannelIDs) (*model.RetentionPolicyWithTeamAndChannelCounts, error) {
	// Strategy:
	// 1. Update policy attributes
	// 2. Delete existing channels from policy
	// 3. Insert new channels into policy
	// 4. Delete existing teams from policy
	// 5. Insert new teams into policy
	// 6. Read new policy

	var err error
	if err = s.checkTeamsExist(patch.TeamIDs); err != nil {
		return nil, err
	}
	if err = s.checkChannelsExist(patch.ChannelIDs); err != nil {
		return nil, err
	}

	policyUpdateQuery := ""
	policyUpdateArgs := []interface{}{}
	if patch.DisplayName != "" || patch.PostDuration != nil {
		builder := s.getQueryBuilder().Update("RetentionPolicies")
		if patch.DisplayName != "" {
			builder = builder.Set("DisplayName", patch.DisplayName)
		}
		if patch.PostDuration != nil {
			builder = builder.Set("PostDuration", *patch.PostDuration)
		}
		policyUpdateQuery, policyUpdateArgs, err = builder.
			Where(sq.Eq{"Id": patch.ID}).
			ToSql()
		if err != nil {
			return nil, err
		}
	}

	channelsDeleteQuery := ""
	channelsDeleteArgs := []interface{}{}
	channelsInsertQuery := ""
	channelsInsertArgs := []interface{}{}
	if patch.ChannelIDs != nil {
		channelsDeleteQuery, channelsDeleteArgs, err = s.getQueryBuilder().
			Delete("RetentionPoliciesChannels").
			Where(sq.Eq{"PolicyId": patch.ID}).
			ToSql()
		if err != nil {
			return nil, err
		}

		channelsInsertQuery, channelsInsertArgs, err = s.buildInsertRetentionPoliciesChannelsQuery(patch.ID, patch.ChannelIDs)
		if err != nil {
			return nil, err
		}
	}

	teamsDeleteQuery := ""
	teamsDeleteArgs := []interface{}{}
	teamsInsertQuery := ""
	teamsInsertArgs := []interface{}{}
	if patch.TeamIDs != nil {
		teamsDeleteQuery, teamsDeleteArgs, err = s.getQueryBuilder().
			Delete("RetentionPoliciesTeams").
			Where(sq.Eq{"PolicyId": patch.ID}).
			ToSql()
		if err != nil {
			return nil, err
		}

		teamsInsertQuery, teamsInsertArgs, err = s.buildInsertRetentionPoliciesTeamsQuery(patch.ID, patch.TeamIDs)
		if err != nil {
			return nil, err
		}
	}

	policySelectQuery, policySelectProps := s.buildGetPolicyQuery(patch.ID)

	txn, err := s.GetMaster().Begin()
	if err != nil {
		return nil, err
	}
	defer finalizeTransaction(txn)
	// Update the fields of the policy in RetentionPolicies
	if _, err = executePossiblyEmptyQuery(txn, policyUpdateQuery, policyUpdateArgs...); err != nil {
		return nil, err
	}
	// Remove all channels from the policy in RetentionPoliciesChannels
	if _, err = executePossiblyEmptyQuery(txn, channelsDeleteQuery, channelsDeleteArgs...); err != nil {
		return nil, err
	}
	// Insert the new channels for the policy in RetentionPoliciesChannels
	if _, err = executePossiblyEmptyQuery(txn, channelsInsertQuery, channelsInsertArgs...); err != nil {
		return nil, err
	}
	// Remove all teams from the policy in RetentionPoliciesTeams
	if _, err = executePossiblyEmptyQuery(txn, teamsDeleteQuery, teamsDeleteArgs...); err != nil {
		return nil, err
	}
	// Insert the new teams for the policy in RetentionPoliciesTeams
	if _, err = executePossiblyEmptyQuery(txn, teamsInsertQuery, teamsInsertArgs...); err != nil {
		return nil, err
	}
	// Select the policy which we just updated
	var newPolicy model.RetentionPolicyWithTeamAndChannelCounts
	if err = txn.SelectOne(&newPolicy, policySelectQuery, policySelectProps); err != nil {
		return nil, err
	}
	if err = txn.Commit(); err != nil {
		return nil, err
	}
	return &newPolicy, nil
}

func (s *SqlRetentionPolicyStore) buildGetPolicyQuery(id string) (query string, props map[string]interface{}) {
	return s.buildGetPoliciesQuery(id, 0, 1)
}

// buildGetPoliciesQuery builds a query to select information for the policy with the specified
// ID, or, if `id` is the empty string, from all policies. The results returned will be sorted by
// policy display name and ID.
func (s *SqlRetentionPolicyStore) buildGetPoliciesQuery(id string, offset, limit int) (query string, props map[string]interface{}) {
	props = map[string]interface{}{"Offset": offset, "Limit": limit}
	whereIdEqualsPolicyId := ""
	if id != "" {
		whereIdEqualsPolicyId = "WHERE RetentionPolicies.Id = :PolicyId"
		props["PolicyId"] = id
	}
	query = `
	SELECT RetentionPolicies.Id,
	       RetentionPolicies.DisplayName,
	       RetentionPolicies.PostDuration,
	       A.Count AS ChannelCount,
	       B.Count AS TeamCount
	FROM RetentionPolicies
	INNER JOIN (
		SELECT RetentionPolicies.Id,
		       COUNT(RetentionPoliciesChannels.ChannelId) AS Count
		FROM RetentionPolicies
		LEFT JOIN RetentionPoliciesChannels ON RetentionPolicies.Id = RetentionPoliciesChannels.PolicyId
		` + whereIdEqualsPolicyId + `
		GROUP BY RetentionPolicies.Id
		ORDER BY RetentionPolicies.DisplayName, RetentionPolicies.Id
		LIMIT :Limit
		OFFSET :Offset
	) AS A ON RetentionPolicies.Id = A.Id
	INNER JOIN (
		SELECT RetentionPolicies.Id,
		       COUNT(RetentionPoliciesTeams.TeamId) AS Count
		FROM RetentionPolicies
		LEFT JOIN RetentionPoliciesTeams ON RetentionPolicies.Id = RetentionPoliciesTeams.PolicyId
		` + whereIdEqualsPolicyId + `
		GROUP BY RetentionPolicies.Id
		ORDER BY RetentionPolicies.DisplayName, RetentionPolicies.Id
		LIMIT :Limit
		OFFSET :Offset
	) AS B ON RetentionPolicies.Id = B.Id
	ORDER BY RetentionPolicies.DisplayName, RetentionPolicies.Id`
	return
}

func (s *SqlRetentionPolicyStore) Get(id string) (*model.RetentionPolicyWithTeamAndChannelCounts, error) {
	query, props := s.buildGetPolicyQuery(id)
	var policy model.RetentionPolicyWithTeamAndChannelCounts
	if err := s.GetReplica().SelectOne(&policy, query, props); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (s *SqlRetentionPolicyStore) GetAll(offset, limit int) (policies []*model.RetentionPolicyWithTeamAndChannelCounts, err error) {
	query, props := s.buildGetPoliciesQuery("", offset, limit)
	_, err = s.GetReplica().Select(&policies, query, props)
	return
}

func (s *SqlRetentionPolicyStore) GetCount() (int64, error) {
	return s.GetReplica().SelectInt("SELECT COUNT(*) FROM RetentionPolicies")
}

func (s *SqlRetentionPolicyStore) Delete(id string) error {
	builder := s.getQueryBuilder().
		Delete("RetentionPolicies").
		Where(sq.Eq{"Id": id})
	result, err := builder.RunWith(s.GetMaster()).Exec()
	if err != nil {
		return err
	}
	numRowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	} else if numRowsAffected == 0 {
		return errors.New("policy not found")
	}
	return nil
}

func (s *SqlRetentionPolicyStore) GetChannels(policyId string, offset, limit int) (channels model.ChannelListWithTeamData, err error) {
	const query = `
	SELECT Channels.*,
	       Teams.DisplayName AS TeamDisplayName,
	       Teams.Name AS TeamName,
	       Teams.UpdateAt AS TeamUpdateAt
	FROM RetentionPoliciesChannels
	INNER JOIN Channels ON RetentionPoliciesChannels.ChannelId = Channels.Id
	INNER JOIN Teams ON Channels.TeamId = Teams.Id
	WHERE RetentionPoliciesChannels.PolicyId = :PolicyId
	ORDER BY Channels.DisplayName, Channels.Id
	LIMIT :Limit
	OFFSET :Offset`
	props := map[string]interface{}{"PolicyId": policyId, "Limit": limit, "Offset": offset}
	_, err = s.GetReplica().Select(&channels, query, props)
	for _, channel := range channels {
		channel.PolicyID = model.NewString(policyId)
	}
	return
}

func (s *SqlRetentionPolicyStore) GetChannelsCount(policyId string) (int64, error) {
	const query = `
	SELECT COUNT(*)
	FROM RetentionPolicies
	INNER JOIN RetentionPoliciesChannels ON RetentionPolicies.Id = RetentionPoliciesChannels.PolicyId
	WHERE RetentionPolicies.Id = :PolicyId`
	props := map[string]interface{}{"PolicyId": policyId}
	return s.GetReplica().SelectInt(query, props)
}

func (s *SqlRetentionPolicyStore) AddChannels(policyId string, channelIds []string) error {
	if len(channelIds) == 0 {
		return nil
	}
	if err := s.checkChannelsExist(channelIds); err != nil {
		return err
	}
	builder := s.getQueryBuilder().
		Insert("RetentionPoliciesChannels").
		Columns("policyId", "channelId")
	for _, channelId := range channelIds {
		builder = builder.Values(policyId, channelId)
	}
	_, err := builder.RunWith(s.GetMaster()).Exec()
	if err != nil {
		switch dbErr := err.(type) {
		case *pq.Error:
			if dbErr.Code == PGForeignKeyViolationErrorCode {
				return store.NewErrNotFound("RetentionPolicy", policyId)
			}
		case *mysql.MySQLError:
			if dbErr.Number == MySQLForeignKeyViolationErrorCode {
				return store.NewErrNotFound("RetentionPolicy", policyId)
			}
		}
	}
	return err
}

func (s *SqlRetentionPolicyStore) RemoveChannels(policyId string, channelIds []string) error {
	if len(channelIds) == 0 {
		return nil
	}
	builder := s.getQueryBuilder().
		Delete("RetentionPoliciesChannels").
		Where(sq.And{
			sq.Eq{"PolicyId": policyId},
			sq.Eq{"ChannelId": channelIds},
		})
	_, err := builder.RunWith(s.GetMaster()).Exec()
	return err
}

func (s *SqlRetentionPolicyStore) GetTeams(policyId string, offset, limit int) (teams []*model.Team, err error) {
	const query = `
	SELECT Teams.* FROM RetentionPoliciesTeams
	INNER JOIN Teams ON RetentionPoliciesTeams.TeamId = Teams.Id
	WHERE RetentionPoliciesTeams.PolicyId = :PolicyId
	ORDER BY Teams.DisplayName, Teams.Id
	LIMIT :Limit
	OFFSET :Offset`
	props := map[string]interface{}{"PolicyId": policyId, "Limit": limit, "Offset": offset}
	_, err = s.GetReplica().Select(&teams, query, props)
	for _, team := range teams {
		team.PolicyID = &policyId
	}
	return
}

func (s *SqlRetentionPolicyStore) GetTeamsCount(policyId string) (int64, error) {
	const query = `
	SELECT COUNT(*)
	FROM RetentionPolicies
	INNER JOIN RetentionPoliciesTeams ON RetentionPolicies.Id = RetentionPoliciesTeams.PolicyId
	WHERE RetentionPolicies.Id = :PolicyId`
	props := map[string]interface{}{"PolicyId": policyId}
	return s.GetReplica().SelectInt(query, props)
}

func (s *SqlRetentionPolicyStore) AddTeams(policyId string, teamIds []string) error {
	if len(teamIds) == 0 {
		return nil
	}
	if err := s.checkTeamsExist(teamIds); err != nil {
		return err
	}
	builder := s.getQueryBuilder().
		Insert("RetentionPoliciesTeams").
		Columns("PolicyId", "TeamId")
	for _, teamId := range teamIds {
		builder = builder.Values(policyId, teamId)
	}
	_, err := builder.RunWith(s.GetMaster()).Exec()
	return err
}

func (s *SqlRetentionPolicyStore) RemoveTeams(policyId string, teamIds []string) error {
	if len(teamIds) == 0 {
		return nil
	}
	builder := s.getQueryBuilder().
		Delete("RetentionPoliciesTeams").
		Where(sq.And{
			sq.Eq{"PolicyId": policyId},
			sq.Eq{"TeamId": teamIds},
		})
	_, err := builder.RunWith(s.GetMaster()).Exec()
	return err
}

func (s *SqlRetentionPolicyStore) GetTeamPoliciesForUser(userID string, offset, limit int) (policies []*model.RetentionPolicyForTeam, err error) {
	const query = `
	SELECT Teams.Id, RetentionPolicies.PostDuration
	FROM Users
	INNER JOIN TeamMembers ON Users.Id = TeamMembers.UserId
	INNER JOIN Teams ON TeamMembers.TeamId = Teams.Id
	INNER JOIN RetentionPoliciesTeams ON Teams.Id = RetentionPoliciesTeams.TeamId
	INNER JOIN RetentionPolicies ON RetentionPoliciesTeams.PolicyId = RetentionPolicies.Id
	WHERE Users.Id = :UserId
		AND TeamMembers.DeleteAt = 0
		AND Teams.DeleteAt = 0
	ORDER BY Teams.Id
	LIMIT :Limit
	OFFSET :Offset`
	props := map[string]interface{}{"UserId": userID, "Limit": limit, "Offset": offset}
	_, err = s.GetReplica().Select(&policies, query, props)
	return
}

func (s *SqlRetentionPolicyStore) GetTeamPoliciesCountForUser(userID string) (int64, error) {
	const query = `
	SELECT COUNT(*)
	FROM Users
	INNER JOIN TeamMembers ON Users.Id = TeamMembers.UserId
	INNER JOIN Teams ON TeamMembers.TeamId = Teams.Id
	INNER JOIN RetentionPoliciesTeams ON Teams.Id = RetentionPoliciesTeams.TeamId
	INNER JOIN RetentionPolicies ON RetentionPoliciesTeams.PolicyId = RetentionPolicies.Id
	WHERE Users.Id = :UserId
		AND TeamMembers.DeleteAt = 0
		AND Teams.DeleteAt = 0`
	props := map[string]interface{}{"UserId": userID}
	return s.GetReplica().SelectInt(query, props)
}

func (s *SqlRetentionPolicyStore) GetChannelPoliciesForUser(userID string, offset, limit int) (policies []*model.RetentionPolicyForChannel, err error) {
	const query = `
	SELECT Channels.Id, RetentionPolicies.PostDuration
	FROM Users
	INNER JOIN ChannelMembers ON Users.Id = ChannelMembers.UserId
	INNER JOIN Channels ON ChannelMembers.ChannelId = Channels.Id
	INNER JOIN RetentionPoliciesChannels ON Channels.Id = RetentionPoliciesChannels.ChannelId
	INNER JOIN RetentionPolicies ON RetentionPoliciesChannels.PolicyId = RetentionPolicies.Id
	WHERE Users.Id = :UserId
		AND Channels.DeleteAt = 0
	ORDER BY Channels.Id
	LIMIT :Limit
	OFFSET :Offset`
	props := map[string]interface{}{"UserId": userID, "Limit": limit, "Offset": offset}
	_, err = s.GetReplica().Select(&policies, query, props)
	return
}

func (s *SqlRetentionPolicyStore) GetChannelPoliciesCountForUser(userID string) (int64, error) {
	const query = `
	SELECT COUNT(*)
	FROM Users
	INNER JOIN ChannelMembers ON Users.Id = ChannelMembers.UserId
	INNER JOIN Channels ON ChannelMembers.ChannelId = Channels.Id
	INNER JOIN RetentionPoliciesChannels ON Channels.Id = RetentionPoliciesChannels.ChannelId
	INNER JOIN RetentionPolicies ON RetentionPoliciesChannels.PolicyId = RetentionPolicies.Id
	WHERE Users.Id = :UserId
		AND Channels.DeleteAt = 0`
	props := map[string]interface{}{"UserId": userID}
	return s.GetReplica().SelectInt(query, props)
}

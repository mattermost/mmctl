// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mattermost/gorp"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlSchemeStore struct {
	*SqlStore
}

func newSqlSchemeStore(sqlStore *SqlStore) store.SchemeStore {
	s := &SqlSchemeStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Scheme{}, "Schemes").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(model.SCHEME_NAME_MAX_LENGTH).SetUnique(true)
		table.ColMap("DisplayName").SetMaxSize(model.SCHEME_DISPLAY_NAME_MAX_LENGTH)
		table.ColMap("Description").SetMaxSize(model.SCHEME_DESCRIPTION_MAX_LENGTH)
		table.ColMap("Scope").SetMaxSize(32)
		table.ColMap("DefaultTeamAdminRole").SetMaxSize(64)
		table.ColMap("DefaultTeamUserRole").SetMaxSize(64)
		table.ColMap("DefaultTeamGuestRole").SetMaxSize(64)
		table.ColMap("DefaultChannelAdminRole").SetMaxSize(64)
		table.ColMap("DefaultChannelUserRole").SetMaxSize(64)
		table.ColMap("DefaultChannelGuestRole").SetMaxSize(64)
	}

	return s
}

func (s SqlSchemeStore) createIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_schemes_channel_guest_role", "Schemes", "DefaultChannelGuestRole")
	s.CreateIndexIfNotExists("idx_schemes_channel_user_role", "Schemes", "DefaultChannelUserRole")
	s.CreateIndexIfNotExists("idx_schemes_channel_admin_role", "Schemes", "DefaultChannelAdminRole")
}

func (s *SqlSchemeStore) Save(scheme *model.Scheme) (*model.Scheme, error) {
	if scheme.Id == "" {
		transaction, err := s.GetMaster().Begin()
		if err != nil {
			return nil, errors.Wrap(err, "begin_transaction")
		}
		defer finalizeTransaction(transaction)

		newScheme, err := s.createScheme(scheme, transaction)
		if err != nil {
			return nil, err
		}
		if err := transaction.Commit(); err != nil {
			return nil, errors.Wrap(err, "commit_transaction")
		}
		return newScheme, nil
	}

	if !scheme.IsValid() {
		return nil, store.NewErrInvalidInput("Scheme", "<any>", fmt.Sprintf("%v", scheme))
	}

	scheme.UpdateAt = model.GetMillis()

	rowsChanged, err := s.GetMaster().Update(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update Scheme")
	}
	if rowsChanged != 1 {
		return nil, errors.New("no record to update")
	}

	return scheme, nil
}

func (s *SqlSchemeStore) createScheme(scheme *model.Scheme, transaction *gorp.Transaction) (*model.Scheme, error) {
	// Fetch the default system scheme roles to populate default permissions.
	defaultRoleNames := []string{model.TEAM_ADMIN_ROLE_ID, model.TEAM_USER_ROLE_ID, model.TEAM_GUEST_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID, model.CHANNEL_USER_ROLE_ID, model.CHANNEL_GUEST_ROLE_ID}
	defaultRoles := make(map[string]*model.Role)
	roles, err := s.SqlStore.Role().GetByNames(defaultRoleNames)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		switch role.Name {
		case model.TEAM_ADMIN_ROLE_ID:
			defaultRoles[model.TEAM_ADMIN_ROLE_ID] = role
		case model.TEAM_USER_ROLE_ID:
			defaultRoles[model.TEAM_USER_ROLE_ID] = role
		case model.TEAM_GUEST_ROLE_ID:
			defaultRoles[model.TEAM_GUEST_ROLE_ID] = role
		case model.CHANNEL_ADMIN_ROLE_ID:
			defaultRoles[model.CHANNEL_ADMIN_ROLE_ID] = role
		case model.CHANNEL_USER_ROLE_ID:
			defaultRoles[model.CHANNEL_USER_ROLE_ID] = role
		case model.CHANNEL_GUEST_ROLE_ID:
			defaultRoles[model.CHANNEL_GUEST_ROLE_ID] = role
		}
	}

	if len(defaultRoles) != 6 {
		return nil, errors.New("createScheme: unable to retrieve default scheme roles")
	}

	// Create the appropriate default roles for the scheme.
	if scheme.Scope == model.SCHEME_SCOPE_TEAM {
		// Team Admin Role
		teamAdminRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team Admin Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_ADMIN_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err := s.SqlStore.Role().(*SqlRoleStore).createRole(teamAdminRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamAdminRole = savedRole.Name

		// Team User Role
		teamUserRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team User Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_USER_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(teamUserRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamUserRole = savedRole.Name

		// Team Guest Role
		teamGuestRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team Guest Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_GUEST_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(teamGuestRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamGuestRole = savedRole.Name
	}

	if scheme.Scope == model.SCHEME_SCOPE_TEAM || scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
		// Channel Admin Role
		channelAdminRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel Admin Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_ADMIN_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelAdminRole.Permissions = []string{}
		}

		savedRole, err := s.SqlStore.Role().(*SqlRoleStore).createRole(channelAdminRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelAdminRole = savedRole.Name

		// Channel User Role
		channelUserRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel User Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_USER_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelUserRole.Permissions = filterModerated(channelUserRole.Permissions)
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(channelUserRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelUserRole = savedRole.Name

		// Channel Guest Role
		channelGuestRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel Guest Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_GUEST_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelGuestRole.Permissions = filterModerated(channelGuestRole.Permissions)
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(channelGuestRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelGuestRole = savedRole.Name
	}

	scheme.Id = model.NewId()
	if scheme.Name == "" {
		scheme.Name = model.NewId()
	}
	scheme.CreateAt = model.GetMillis()
	scheme.UpdateAt = scheme.CreateAt

	// Validate the scheme
	if !scheme.IsValidForCreate() {
		return nil, store.NewErrInvalidInput("Scheme", "<any>", fmt.Sprintf("%v", scheme))
	}

	if err := transaction.Insert(scheme); err != nil {
		return nil, errors.Wrap(err, "failed to save Scheme")
	}

	return scheme, nil
}

func filterModerated(permissions []string) []string {
	filteredPermissions := []string{}
	for _, perm := range permissions {
		if _, ok := model.ChannelModeratedPermissionsMap[perm]; ok {
			filteredPermissions = append(filteredPermissions, perm)
		}
	}
	return filteredPermissions
}

func (s *SqlSchemeStore) Get(schemeId string) (*model.Scheme, error) {
	var scheme model.Scheme
	if err := s.GetReplica().SelectOne(&scheme, "SELECT * from Schemes WHERE Id = :Id", map[string]interface{}{"Id": schemeId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Scheme", fmt.Sprintf("schemeId=%s", schemeId))
		}
		return nil, errors.Wrapf(err, "failed to get Scheme with schemeId=%s", schemeId)
	}

	return &scheme, nil
}

func (s *SqlSchemeStore) GetByName(schemeName string) (*model.Scheme, error) {
	var scheme model.Scheme

	if err := s.GetReplica().SelectOne(&scheme, "SELECT * from Schemes WHERE Name = :Name", map[string]interface{}{"Name": schemeName}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Scheme", fmt.Sprintf("schemeName=%s", schemeName))
		}
		return nil, errors.Wrapf(err, "failed to get Scheme with schemeName=%s", schemeName)
	}

	return &scheme, nil
}

func (s *SqlSchemeStore) Delete(schemeId string) (*model.Scheme, error) {
	// Get the scheme
	var scheme model.Scheme
	if err := s.GetReplica().SelectOne(&scheme, "SELECT * from Schemes WHERE Id = :Id", map[string]interface{}{"Id": schemeId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Scheme", fmt.Sprintf("schemeId=%s", schemeId))
		}
		return nil, errors.Wrapf(err, "failed to get Scheme with schemeId=%s", schemeId)
	}

	// Update any teams or channels using this scheme to the default scheme.
	if scheme.Scope == model.SCHEME_SCOPE_TEAM {
		if _, err := s.GetMaster().Exec("UPDATE Teams SET SchemeId = '' WHERE SchemeId = :SchemeId", map[string]interface{}{"SchemeId": schemeId}); err != nil {
			return nil, errors.Wrapf(err, "failed to update Teams with schemeId=%s", schemeId)
		}

		s.Team().ClearCaches()
	} else if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
		if _, err := s.GetMaster().Exec("UPDATE Channels SET SchemeId = '' WHERE SchemeId = :SchemeId", map[string]interface{}{"SchemeId": schemeId}); err != nil {
			return nil, errors.Wrapf(err, "failed to update Channels with schemeId=%s", schemeId)
		}
	}

	// Blow away the channel caches.
	s.Channel().ClearCaches()

	// Delete the roles belonging to the scheme.
	roleNames := []string{scheme.DefaultChannelGuestRole, scheme.DefaultChannelUserRole, scheme.DefaultChannelAdminRole}
	if scheme.Scope == model.SCHEME_SCOPE_TEAM {
		roleNames = append(roleNames, scheme.DefaultTeamGuestRole, scheme.DefaultTeamUserRole, scheme.DefaultTeamAdminRole)
	}

	var inQueryList []string
	queryArgs := make(map[string]interface{})
	for i, roleId := range roleNames {
		inQueryList = append(inQueryList, fmt.Sprintf(":RoleName%v", i))
		queryArgs[fmt.Sprintf("RoleName%v", i)] = roleId
	}
	inQuery := strings.Join(inQueryList, ", ")

	time := model.GetMillis()
	queryArgs["UpdateAt"] = time
	queryArgs["DeleteAt"] = time

	if _, err := s.GetMaster().Exec("UPDATE Roles SET UpdateAt = :UpdateAt, DeleteAt = :DeleteAt WHERE Name IN ("+inQuery+")", queryArgs); err != nil {
		return nil, errors.Wrapf(err, "failed to update Roles with name in (%s)", inQuery)
	}

	// Delete the scheme itself.
	scheme.UpdateAt = time
	scheme.DeleteAt = time

	rowsChanged, err := s.GetMaster().Update(&scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update Scheme with schemeId=%s", schemeId)
	}
	if rowsChanged != 1 {
		return nil, errors.New("no record to update")
	}
	return &scheme, nil
}

func (s *SqlSchemeStore) GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error) {
	var schemes []*model.Scheme

	scopeClause := ""
	if scope != "" {
		scopeClause = " AND Scope=:Scope "
	}

	if _, err := s.GetReplica().Select(&schemes, "SELECT * from Schemes WHERE DeleteAt = 0 "+scopeClause+" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Limit": limit, "Offset": offset, "Scope": scope}); err != nil {
		return nil, errors.Wrapf(err, "failed to get Schemes")
	}

	return schemes, nil
}

func (s *SqlSchemeStore) PermanentDeleteAll() error {
	if _, err := s.GetMaster().Exec("DELETE from Schemes"); err != nil {
		return errors.Wrap(err, "failed to delete Schemes")
	}

	return nil
}

func (s *SqlSchemeStore) CountByScope(scope string) (int64, error) {
	count, err := s.GetReplica().SelectInt("SELECT count(*) FROM Schemes WHERE Scope = :Scope AND DeleteAt = 0", map[string]interface{}{"Scope": scope})
	if err != nil {
		return int64(0), errors.Wrap(err, "failed to count Schemes by scope")
	}
	return count, nil
}

func (s *SqlSchemeStore) CountWithoutPermission(schemeScope, permissionID string, roleScope model.RoleScope, roleType model.RoleType) (int64, error) {
	joinCol := fmt.Sprintf("Default%s%sRole", roleScope, roleType)
	query := fmt.Sprintf(`
		SELECT
			count(*)
		FROM Schemes
			JOIN Roles ON Roles.Name = Schemes.%s
		WHERE
			Schemes.DeleteAt = 0 AND
			Schemes.Scope = '%s' AND
			Roles.Permissions NOT LIKE '%%%s%%'
	`, joinCol, schemeScope, permissionID)
	count, err := s.GetReplica().SelectInt(query)
	if err != nil {
		return int64(0), errors.Wrap(err, "failed to count Schemes without permission")
	}
	return count, nil
}

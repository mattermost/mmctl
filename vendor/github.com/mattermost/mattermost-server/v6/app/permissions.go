// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/app/request"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/product"
)

const permissionsExportBatchSize = 100
const systemSchemeName = "00000000-0000-0000-0000-000000000000" // Prevents collisions with user-created schemes.

// Ensure permissions service wrapper implements `product.PermissionService`
var _ product.PermissionService = (*permissionsServiceWrapper)(nil)

// permissionsServiceWrapper provides an implementation of `product.PermissionService` for use by products.
type permissionsServiceWrapper struct {
	app AppIface
}

func (s *permissionsServiceWrapper) HasPermissionTo(userID string, permission *model.Permission) bool {
	return s.app.HasPermissionTo(userID, permission)
}

func (s *permissionsServiceWrapper) HasPermissionToTeam(userID string, teamID string, permission *model.Permission) bool {
	return s.app.HasPermissionToTeam(userID, teamID, permission)
}

func (s *permissionsServiceWrapper) HasPermissionToChannel(askingUserID string, channelID string, permission *model.Permission) bool {
	return s.app.HasPermissionToChannel(request.EmptyContext(s.app.Log()), askingUserID, channelID, permission)
}

func (a *App) ResetPermissionsSystem() *model.AppError {
	// Reset all Teams to not have a scheme.
	if err := a.Srv().Store().Team().ResetAllTeamSchemes(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.team.reset_all_team_schemes.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Reset all Channels to not have a scheme.
	if err := a.Srv().Store().Channel().ResetAllChannelSchemes(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.channel.reset_all_channel_schemes.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Reset all Custom Role assignments to Users.
	if err := a.Srv().Store().User().ClearAllCustomRoleAssignments(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.user.clear_all_custom_role_assignments.select.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Reset all Custom Role assignments to TeamMembers.
	if err := a.Srv().Store().Team().ClearAllCustomRoleAssignments(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.team.clear_all_custom_role_assignments.select.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Reset all Custom Role assignments to ChannelMembers.
	if err := a.Srv().Store().Channel().ClearAllCustomRoleAssignments(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.channel.clear_all_custom_role_assignments.select.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Purge all schemes from the database.
	if err := a.Srv().Store().Scheme().PermanentDeleteAll(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.scheme.permanent_delete_all.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Purge all roles from the database.
	if err := a.Srv().Store().Role().PermanentDeleteAll(); err != nil {
		return model.NewAppError("ResetPermissionsSystem", "app.role.permanent_delete_all.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Remove the "System" table entry that marks the advanced permissions migration as done.
	if _, err := a.Srv().Store().System().PermanentDeleteByName(model.AdvancedPermissionsMigrationKey); err != nil {
		return model.NewAppError("ResetPermissionSystem", "app.system.permanent_delete_by_name.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Remove the "System" table entry that marks the emoji permissions migration as done.
	if _, err := a.Srv().Store().System().PermanentDeleteByName(EmojisPermissionsMigrationKey); err != nil {
		return model.NewAppError("ResetPermissionSystem", "app.system.permanent_delete_by_name.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Remove the "System" table entry that marks the guest roles permissions migration as done.
	if _, err := a.Srv().Store().System().PermanentDeleteByName(GuestRolesCreationMigrationKey); err != nil {
		return model.NewAppError("ResetPermissionSystem", "app.system.permanent_delete_by_name.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	// Now that the permissions system has been reset, re-run the migration to reinitialise it.
	a.DoAppMigrations()

	return nil
}

func (a *App) ExportPermissions(w io.Writer) error {

	next := a.SchemesIterator("", permissionsExportBatchSize)
	var schemeBatch []*model.Scheme

	for schemeBatch = next(); len(schemeBatch) > 0; schemeBatch = next() {

		for _, scheme := range schemeBatch {

			roleNames := []string{
				scheme.DefaultTeamAdminRole,
				scheme.DefaultTeamUserRole,
				scheme.DefaultTeamGuestRole,
				scheme.DefaultChannelAdminRole,
				scheme.DefaultChannelUserRole,
				scheme.DefaultChannelGuestRole,
			}

			roles := []*model.Role{}
			for _, roleName := range roleNames {
				if roleName == "" {
					continue
				}
				role, err := a.GetRoleByName(context.Background(), roleName)
				if err != nil {
					return err
				}
				roles = append(roles, role)
			}

			schemeExport, err := json.Marshal(&model.SchemeConveyor{
				Name:         scheme.Name,
				DisplayName:  scheme.DisplayName,
				Description:  scheme.Description,
				Scope:        scheme.Scope,
				TeamAdmin:    scheme.DefaultTeamAdminRole,
				TeamUser:     scheme.DefaultTeamUserRole,
				TeamGuest:    scheme.DefaultTeamGuestRole,
				ChannelAdmin: scheme.DefaultChannelAdminRole,
				ChannelUser:  scheme.DefaultChannelUserRole,
				ChannelGuest: scheme.DefaultChannelGuestRole,
				Roles:        roles,
			})
			if err != nil {
				return err
			}

			schemeExport = append(schemeExport, []byte("\n")...)

			_, err = w.Write(schemeExport)
			if err != nil {
				return err
			}
		}

	}

	defaultRoleNames := []string{}
	for _, dr := range model.MakeDefaultRoles() {
		defaultRoleNames = append(defaultRoleNames, dr.Name)
	}

	roles, appErr := a.GetRolesByNames(defaultRoleNames)
	if appErr != nil {
		return errors.New(appErr.Message)
	}

	schemeExport, err := json.Marshal(&model.SchemeConveyor{
		Name:  systemSchemeName,
		Roles: roles,
	})
	if err != nil {
		return err
	}

	schemeExport = append(schemeExport, []byte("\n")...)

	_, err = w.Write(schemeExport)
	return err
}

func (a *App) ImportPermissions(jsonl io.Reader) error {
	createdSchemeIDs := []string{}

	scanner := bufio.NewScanner(jsonl)

	for scanner.Scan() {
		var schemeConveyor *model.SchemeConveyor
		err := json.Unmarshal(scanner.Bytes(), &schemeConveyor)
		if err != nil {
			rollback(a, createdSchemeIDs)
			return err
		}

		if schemeConveyor.Name == systemSchemeName {
			for _, roleIn := range schemeConveyor.Roles {
				dbRole, err := a.GetRoleByName(context.Background(), roleIn.Name)
				if err != nil {
					rollback(a, createdSchemeIDs)
					return errors.New(err.Message)
				}
				_, err = a.PatchRole(dbRole, &model.RolePatch{
					Permissions: &roleIn.Permissions,
				})
				if err != nil {
					rollback(a, createdSchemeIDs)
					return err
				}
			}
			continue
		}

		// Create the new Scheme. The new Roles are created automatically.
		var appErr *model.AppError
		schemeCreated, appErr := a.CreateScheme(schemeConveyor.Scheme())
		if appErr != nil {
			rollback(a, createdSchemeIDs)
			return errors.New(appErr.Message)
		}
		createdSchemeIDs = append(createdSchemeIDs, schemeCreated.Id)

		schemeIn := schemeConveyor.Scheme()
		roleNameTuples := [][]string{
			{schemeCreated.DefaultTeamAdminRole, schemeIn.DefaultTeamAdminRole},
			{schemeCreated.DefaultTeamUserRole, schemeIn.DefaultTeamUserRole},
			{schemeCreated.DefaultTeamGuestRole, schemeIn.DefaultTeamGuestRole},
			{schemeCreated.DefaultChannelAdminRole, schemeIn.DefaultChannelAdminRole},
			{schemeCreated.DefaultChannelUserRole, schemeIn.DefaultChannelUserRole},
			{schemeCreated.DefaultChannelGuestRole, schemeIn.DefaultChannelGuestRole},
		}
		for _, roleNameTuple := range roleNameTuples {
			if roleNameTuple[0] == "" || roleNameTuple[1] == "" {
				continue
			}

			err = updateRole(a, schemeConveyor, roleNameTuple[0], roleNameTuple[1])
			if err != nil {
				// Delete the new Schemes. The new Roles are deleted automatically.
				rollback(a, createdSchemeIDs)
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		rollback(a, createdSchemeIDs)
		return err
	}

	return nil
}

func rollback(a *App, createdSchemeIDs []string) {
	for _, schemeID := range createdSchemeIDs {
		a.DeleteScheme(schemeID)
	}
}

func updateRole(a *App, sc *model.SchemeConveyor, roleCreatedName, defaultRoleName string) error {
	var err *model.AppError

	roleCreated, err := a.GetRoleByName(context.Background(), roleCreatedName)
	if err != nil {
		return errors.New(err.Message)
	}

	var roleIn *model.Role
	for _, role := range sc.Roles {
		if role.Name == defaultRoleName {
			roleIn = role
			break
		}
	}

	roleCreated.DisplayName = roleIn.DisplayName
	roleCreated.Description = roleIn.Description
	roleCreated.Permissions = roleIn.Permissions

	_, err = a.UpdateRole(roleCreated)
	if err != nil {
		return errors.New(fmt.Sprintf("%v: %v\n", err.Message, err.DetailedError))
	}

	return nil
}

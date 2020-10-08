// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/mattermost/mattermost-server/v5/audit"
	"github.com/mattermost/mattermost-server/v5/config"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"
)

var writeFilter func(c *Context, structField reflect.StructField) bool
var readFilter func(c *Context, structField reflect.StructField) bool
var permissionMap map[string]*model.Permission

type filterType string

const filterTypeWrite filterType = "write"
const filterTypeRead filterType = "read"

func (api *API) InitConfig() {
	api.BaseRoutes.ApiRoot.Handle("/config", api.ApiSessionRequired(getConfig)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/config", api.ApiSessionRequired(updateConfig)).Methods("PUT")
	api.BaseRoutes.ApiRoot.Handle("/config/patch", api.ApiSessionRequired(patchConfig)).Methods("PUT")
	api.BaseRoutes.ApiRoot.Handle("/config/reload", api.ApiSessionRequired(configReload)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/config/client", api.ApiHandler(getClientConfig)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/config/environment", api.ApiSessionRequired(getEnvironmentConfig)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/config/migrate", api.ApiSessionRequired(migrateConfig)).Methods("POST")
}

func init() {
	writeFilter = makeFilterConfigByPermission(filterTypeWrite)
	readFilter = makeFilterConfigByPermission(filterTypeRead)
	permissionMap = map[string]*model.Permission{}
	for _, p := range model.AllPermissions {
		permissionMap[p.Id] = p
	}
}

func getConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleReadPermissions) {
		c.SetPermissionError(model.SysconsoleReadPermissions...)
		return
	}

	auditRec := c.MakeAuditRecord("getConfig", audit.Fail)
	defer c.LogAuditRec(auditRec)

	cfg, err := config.Merge(&model.Config{}, c.App.GetSanitizedConfig(), &utils.MergeConfig{
		StructFieldFilter: func(structField reflect.StructField, base, patch reflect.Value) bool {
			return readFilter(c, structField)
		},
	})
	if err != nil {
		c.Err = model.NewAppError("getConfig", "api.config.get_config.restricted_merge.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	auditRec.Success()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if c.App.Srv().License() != nil && *c.App.Srv().License().Features.Cloud && *cfg.ExperimentalSettings.RestrictSystemAdmin {
		w.Write([]byte(cfg.ToJsonFiltered(model.ConfigAccessTagType, model.ConfigAccessTagCloudRestrictable)))
	} else {
		w.Write([]byte(cfg.ToJson()))
	}
}

func configReload(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("configReload", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleReadPermissions) {
		c.SetPermissionError(model.SysconsoleReadPermissions...)
		return
	}

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("configReload", "api.restricted_system_admin", nil, "", http.StatusBadRequest)
		return
	}

	c.App.ReloadConfig()

	auditRec.Success()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ReturnStatusOK(w)
}

func updateConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	cfg := model.ConfigFromJson(r.Body)
	if cfg == nil {
		c.SetInvalidParam("config")
		return
	}

	auditRec := c.MakeAuditRecord("updateConfig", audit.Fail)
	defer c.LogAuditRec(auditRec)

	cfg.SetDefaults()

	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleWritePermissions) {
		c.SetPermissionError(model.SysconsoleWritePermissions...)
		return
	}

	appCfg := c.App.Config()
	if *appCfg.ServiceSettings.SiteURL != "" && *cfg.ServiceSettings.SiteURL == "" {
		c.Err = model.NewAppError("updateConfig", "api.config.update_config.clear_siteurl.app_error", nil, "", http.StatusBadRequest)
		return
	}

	var err1 error
	cfg, err1 = config.Merge(appCfg, cfg, &utils.MergeConfig{
		StructFieldFilter: func(structField reflect.StructField, base, patch reflect.Value) bool {
			return writeFilter(c, structField)
		},
	})
	if err1 != nil {
		c.Err = model.NewAppError("updateConfig", "api.config.update_config.restricted_merge.app_error", nil, err1.Error(), http.StatusInternalServerError)
	}

	// Do not allow plugin uploads to be toggled through the API
	cfg.PluginSettings.EnableUploads = appCfg.PluginSettings.EnableUploads

	// Do not allow certificates to be changed through the API
	cfg.PluginSettings.SignaturePublicKeyFiles = appCfg.PluginSettings.SignaturePublicKeyFiles

	c.App.HandleMessageExportConfig(cfg, appCfg)

	err := cfg.IsValid()
	if err != nil {
		c.Err = err
		return
	}

	err = c.App.SaveConfig(cfg, true)
	if err != nil {
		c.Err = err
		return
	}

	cfg, mergeErr := config.Merge(&model.Config{}, c.App.GetSanitizedConfig(), &utils.MergeConfig{
		StructFieldFilter: func(structField reflect.StructField, base, patch reflect.Value) bool {
			return readFilter(c, structField)
		},
	})
	if mergeErr != nil {
		c.Err = model.NewAppError("getConfig", "api.config.update_config.restricted_merge.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	auditRec.Success()
	c.LogAudit("updateConfig")

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if c.App.Srv().License() != nil && *c.App.Srv().License().Features.Cloud && *cfg.ExperimentalSettings.RestrictSystemAdmin {
		w.Write([]byte(cfg.ToJsonFiltered(model.ConfigAccessTagType, model.ConfigAccessTagCloudRestrictable)))
	} else {
		w.Write([]byte(cfg.ToJson()))
	}
}

func getClientConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	if format == "" {
		c.Err = model.NewAppError("getClientConfig", "api.config.client.old_format.app_error", nil, "", http.StatusNotImplemented)
		return
	}

	if format != "old" {
		c.SetInvalidParam("format")
		return
	}

	var config map[string]string
	if len(c.App.Session().UserId) == 0 {
		config = c.App.LimitedClientConfigWithComputed()
	} else {
		config = c.App.ClientConfigWithComputed()
	}

	w.Write([]byte(model.MapToJson(config)))
}

func getEnvironmentConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	envConfig := c.App.GetEnvironmentConfig()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(model.StringInterfaceToJson(envConfig)))
}

func patchConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	cfg := model.ConfigFromJson(r.Body)
	if cfg == nil {
		c.SetInvalidParam("config")
		return
	}

	auditRec := c.MakeAuditRecord("patchConfig", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleWritePermissions) {
		c.SetPermissionError(model.SysconsoleWritePermissions...)
		return
	}

	appCfg := c.App.Config()
	if *appCfg.ServiceSettings.SiteURL != "" && cfg.ServiceSettings.SiteURL != nil && *cfg.ServiceSettings.SiteURL == "" {
		c.Err = model.NewAppError("patchConfig", "api.config.update_config.clear_siteurl.app_error", nil, "", http.StatusBadRequest)
		return
	}

	filterFn := func(structField reflect.StructField, base, patch reflect.Value) bool {
		return writeFilter(c, structField)
	}

	// Do not allow plugin uploads to be toggled through the API
	cfg.PluginSettings.EnableUploads = appCfg.PluginSettings.EnableUploads

	if cfg.MessageExportSettings.EnableExport != nil {
		c.App.HandleMessageExportConfig(cfg, appCfg)
	}

	updatedCfg, mergeErr := config.Merge(appCfg, cfg, &utils.MergeConfig{
		StructFieldFilter: filterFn,
	})

	if mergeErr != nil {
		c.Err = model.NewAppError("patchConfig", "api.config.update_config.restricted_merge.app_error", nil, mergeErr.Error(), http.StatusInternalServerError)
		return
	}

	err := updatedCfg.IsValid()
	if err != nil {
		c.Err = err
		return
	}

	err = c.App.SaveConfig(updatedCfg, true)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()

	cfg, mergeErr = config.Merge(&model.Config{}, c.App.GetSanitizedConfig(), &utils.MergeConfig{
		StructFieldFilter: func(structField reflect.StructField, base, patch reflect.Value) bool {
			return readFilter(c, structField)
		},
	})
	if mergeErr != nil {
		c.Err = model.NewAppError("getConfig", "api.config.patch_config.restricted_merge.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if c.App.Srv().License() != nil && *c.App.Srv().License().Features.Cloud && *cfg.ExperimentalSettings.RestrictSystemAdmin {
		w.Write([]byte(cfg.ToJsonFiltered(model.ConfigAccessTagType, model.ConfigAccessTagCloudRestrictable)))
	} else {
		w.Write([]byte(cfg.ToJson()))
	}
}

func makeFilterConfigByPermission(accessType filterType) func(c *Context, structField reflect.StructField) bool {
	return func(c *Context, structField reflect.StructField) bool {
		if structField.Type.Kind() == reflect.Struct {
			return true
		}

		tagPermissions := strings.Split(structField.Tag.Get("access"), ",")

		// If there are no access tag values and the role has manage_system, no need to continue
		// checking permissions.
		if len(tagPermissions) == 0 {
			if c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
				return true
			}
		}

		// one iteration for write_restrictable value, it could be anywhere in the order of values
		for _, val := range tagPermissions {
			tagValue := strings.TrimSpace(val)
			if tagValue == "" {
				continue
			}
			// ConfigAccessTagWriteRestrictable trumps all other permissions
			if tagValue == model.ConfigAccessTagWriteRestrictable || tagValue == model.ConfigAccessTagCloudRestrictable {
				if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin && accessType == filterTypeWrite {
					return false
				}
				continue
			}
		}

		// another iteration for permissions checks of other tag values
		for _, val := range tagPermissions {
			tagValue := strings.TrimSpace(val)
			if tagValue == "" {
				continue
			}
			if tagValue == model.ConfigAccessTagWriteRestrictable {
				continue
			}
			if tagValue == model.ConfigAccessTagCloudRestrictable {
				continue
			}
			permissionID := fmt.Sprintf("sysconsole_%s_%s", accessType, tagValue)
			if permission, ok := permissionMap[permissionID]; ok {
				if c.App.SessionHasPermissionTo(*c.App.Session(), permission) {
					return true
				}
			} else {
				mlog.Warn("Unrecognized config permissions tag value.", mlog.String("tag_value", permissionID))
			}
		}

		// with manage_system, default to allow, otherwise default not-allow
		return c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM)
	}
}

func migrateConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.StringInterfaceFromJson(r.Body)
	from, ok := props["from"].(string)
	if !ok {
		c.SetInvalidParam("from")
		return
	}
	to, ok := props["to"].(string)
	if !ok {
		c.SetInvalidParam("to")
		return
	}

	auditRec := c.MakeAuditRecord("migrateConfig", audit.Fail)
	auditRec.AddMeta("from", from)
	auditRec.AddMeta("to", to)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	err := config.Migrate(from, to)
	if err != nil {
		c.Err = model.NewAppError("migrateConfig", "api.config.migrate_config.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/audit"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/cache"
	"github.com/mattermost/mattermost-server/v5/services/filesstore"
	"github.com/mattermost/mattermost-server/v5/services/upgrader"
)

const (
	REDIRECT_LOCATION_CACHE_SIZE = 10000
	DEFAULT_SERVER_BUSY_SECONDS  = 3600
	MAX_SERVER_BUSY_SECONDS      = 86400
)

var redirectLocationDataCache = cache.NewLRU(&cache.LRUOptions{
	Size: REDIRECT_LOCATION_CACHE_SIZE,
})

func (api *API) InitSystem() {
	api.BaseRoutes.System.Handle("/ping", api.ApiHandler(getSystemPing)).Methods("GET")

	api.BaseRoutes.System.Handle("/timezones", api.ApiSessionRequired(getSupportedTimezones)).Methods("GET")

	api.BaseRoutes.ApiRoot.Handle("/audits", api.ApiSessionRequired(getAudits)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/email/test", api.ApiSessionRequired(testEmail)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/site_url/test", api.ApiSessionRequired(testSiteURL)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/file/s3_test", api.ApiSessionRequired(testS3)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/database/recycle", api.ApiSessionRequired(databaseRecycle)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/caches/invalidate", api.ApiSessionRequired(invalidateCaches)).Methods("POST")

	api.BaseRoutes.ApiRoot.Handle("/logs", api.ApiSessionRequired(getLogs)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/logs", api.ApiHandler(postLog)).Methods("POST")

	api.BaseRoutes.ApiRoot.Handle("/analytics/old", api.ApiSessionRequired(getAnalytics)).Methods("GET")

	api.BaseRoutes.ApiRoot.Handle("/redirect_location", api.ApiSessionRequiredTrustRequester(getRedirectLocation)).Methods("GET")

	api.BaseRoutes.ApiRoot.Handle("/notifications/ack", api.ApiSessionRequired(pushNotificationAck)).Methods("POST")

	api.BaseRoutes.ApiRoot.Handle("/server_busy", api.ApiSessionRequired(setServerBusy)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/server_busy", api.ApiSessionRequired(getServerBusyExpires)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/server_busy", api.ApiSessionRequired(clearServerBusy)).Methods("DELETE")
	api.BaseRoutes.ApiRoot.Handle("/upgrade_to_enterprise", api.ApiSessionRequired(upgradeToEnterprise)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/upgrade_to_enterprise/status", api.ApiSessionRequired(upgradeToEnterpriseStatus)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/restart", api.ApiSessionRequired(restart)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/warn_metrics/status", api.ApiSessionRequired(getWarnMetricsStatus)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/warn_metrics/ack/{warn_metric_id:[A-Za-z0-9-_]+}", api.ApiHandler(sendWarnMetricAckEmail)).Methods("POST")
	api.BaseRoutes.ApiRoot.Handle("/warn_metrics/trial-license-ack/{warn_metric_id:[A-Za-z0-9-_]+}", api.ApiHandler(requestTrialLicenseAndAckWarnMetric)).Methods("POST")
	api.BaseRoutes.System.Handle("/notices/{team_id:[A-Za-z0-9]+}", api.ApiSessionRequired(getProductNotices)).Methods("GET")
	api.BaseRoutes.System.Handle("/notices/view", api.ApiSessionRequired(updateViewedProductNotices)).Methods("PUT")
}

func getSystemPing(c *Context, w http.ResponseWriter, r *http.Request) {
	reqs := c.App.Config().ClientRequirements

	s := make(map[string]string)
	s[model.STATUS] = model.STATUS_OK
	s["AndroidLatestVersion"] = reqs.AndroidLatestVersion
	s["AndroidMinVersion"] = reqs.AndroidMinVersion
	s["DesktopLatestVersion"] = reqs.DesktopLatestVersion
	s["DesktopMinVersion"] = reqs.DesktopMinVersion
	s["IosLatestVersion"] = reqs.IosLatestVersion
	s["IosMinVersion"] = reqs.IosMinVersion

	testflag := c.App.Config().FeatureFlags.TestFeature
	if testflag != "off" {
		s["TestFeatureFlag"] = testflag
	}

	actualGoroutines := runtime.NumGoroutine()
	if *c.App.Config().ServiceSettings.GoroutineHealthThreshold > 0 && actualGoroutines >= *c.App.Config().ServiceSettings.GoroutineHealthThreshold {
		mlog.Warn("The number of running goroutines is over the health threshold", mlog.Int("goroutines", actualGoroutines), mlog.Int("health_threshold", *c.App.Config().ServiceSettings.GoroutineHealthThreshold))
		s[model.STATUS] = model.STATUS_UNHEALTHY
	}

	// Enhanced ping health check:
	// If an extra form value is provided then perform extra health checks for
	// database and file storage backends.
	if r.FormValue("get_server_status") != "" {
		dbStatusKey := "database_status"
		s[dbStatusKey] = model.STATUS_OK

		writeErr := c.App.DBHealthCheckWrite()
		if writeErr != nil {
			mlog.Warn("Unable to write to database.", mlog.Err(writeErr))
			s[dbStatusKey] = model.STATUS_UNHEALTHY
			s[model.STATUS] = model.STATUS_UNHEALTHY
		}

		writeErr = c.App.DBHealthCheckDelete()
		if writeErr != nil {
			mlog.Warn("Unable to remove ping health check value from database.", mlog.Err(writeErr))
			s[dbStatusKey] = model.STATUS_UNHEALTHY
			s[model.STATUS] = model.STATUS_UNHEALTHY
		}

		if s[dbStatusKey] == model.STATUS_OK {
			mlog.Debug("Able to write to database.")
		}

		filestoreStatusKey := "filestore_status"
		s[filestoreStatusKey] = model.STATUS_OK
		license := c.App.Srv().License()
		backend, appErr := filesstore.NewFileBackend(&c.App.Config().FileSettings, license != nil && *license.Features.Compliance)
		if appErr == nil {
			appErr = backend.TestConnection()
			if appErr != nil {
				s[filestoreStatusKey] = model.STATUS_UNHEALTHY
				s[model.STATUS] = model.STATUS_UNHEALTHY
			}
		} else {
			mlog.Debug("Unable to get filestore for ping status.", mlog.Err(appErr))
			s[filestoreStatusKey] = model.STATUS_UNHEALTHY
			s[model.STATUS] = model.STATUS_UNHEALTHY
		}

		w.Header().Set(model.STATUS, s[model.STATUS])
		w.Header().Set(dbStatusKey, s[dbStatusKey])
		w.Header().Set(filestoreStatusKey, s[filestoreStatusKey])
	}

	if s[model.STATUS] != model.STATUS_OK {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(model.MapToJson(s)))
}

func testEmail(c *Context, w http.ResponseWriter, r *http.Request) {
	cfg := model.ConfigFromJson(r.Body)
	if cfg == nil {
		cfg = c.App.Config()
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("testEmail", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	err := c.App.TestEmail(c.App.Session().UserId, cfg)
	if err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func testSiteURL(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("testSiteURL", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	props := model.MapFromJson(r.Body)
	siteURL := props["site_url"]
	if siteURL == "" {
		c.SetInvalidParam("site_url")
		return
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT) && siteURL != *c.App.Config().ServiceSettings.SiteURL {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	err := c.App.TestSiteURL(siteURL)
	if err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func getAudits(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("getAudits", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_COMPLIANCE) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_COMPLIANCE)
		return
	}

	audits, err := c.App.GetAuditsPage("", c.Params.Page, c.Params.PerPage)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	auditRec.AddMeta("page", c.Params.Page)
	auditRec.AddMeta("audits_per_page", c.Params.LogsPerPage)

	w.Write([]byte(audits.ToJson()))
}

func databaseRecycle(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT)
		return
	}

	auditRec := c.MakeAuditRecord("databaseRecycle", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("databaseRecycle", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	c.App.RecycleDatabaseConnection()

	auditRec.Success()
	ReturnStatusOK(w)
}

func invalidateCaches(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_WRITE_ENVIRONMENT)
		return
	}

	auditRec := c.MakeAuditRecord("invalidateCaches", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("invalidateCaches", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	err := c.App.Srv().InvalidateAllCaches()
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ReturnStatusOK(w)
}

func getLogs(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("getLogs", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("getLogs", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_REPORTING) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_REPORTING)
		return
	}

	lines, err := c.App.GetLogs(c.Params.Page, c.Params.LogsPerPage)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.AddMeta("page", c.Params.Page)
	auditRec.AddMeta("logs_per_page", c.Params.LogsPerPage)

	w.Write([]byte(model.ArrayToJson(lines)))
}

func postLog(c *Context, w http.ResponseWriter, r *http.Request) {
	forceToDebug := false

	if !*c.App.Config().ServiceSettings.EnableDeveloper {
		if c.App.Session().UserId == "" {
			c.Err = model.NewAppError("postLog", "api.context.permissions.app_error", nil, "", http.StatusForbidden)
			return
		}

		if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
			forceToDebug = true
		}
	}

	m := model.MapFromJson(r.Body)
	lvl := m["level"]
	msg := m["message"]

	if len(msg) > 400 {
		msg = msg[0:399]
	}

	msg = "Client Logs API Endpoint Message: " + msg
	fields := []mlog.Field{
		mlog.String("type", "client_message"),
		mlog.String("user_agent", c.App.UserAgent()),
	}

	if !forceToDebug && lvl == "ERROR" {
		mlog.Error(msg, fields...)
	} else {
		mlog.Debug(msg, fields...)
	}

	m["message"] = msg
	w.Write([]byte(model.MapToJson(m)))
}

func getAnalytics(c *Context, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	teamId := r.URL.Query().Get("team_id")

	if name == "" {
		name = "standard"
	}

	permissions := []*model.Permission{
		model.PERMISSION_SYSCONSOLE_READ_REPORTING,
		model.PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_USERS,
	}
	if !c.App.SessionHasPermissionToAny(*c.App.Session(), permissions) {
		c.SetPermissionError(permissions...)
		return
	}

	rows, err := c.App.GetAnalytics(name, teamId)
	if err != nil {
		c.Err = err
		return
	}

	if rows == nil {
		c.SetInvalidParam("name")
		return
	}

	w.Write([]byte(rows.ToJson()))
}

func getSupportedTimezones(c *Context, w http.ResponseWriter, r *http.Request) {
	supportedTimezones := c.App.Timezones().GetSupported()
	if supportedTimezones == nil {
		supportedTimezones = make([]string, 0)
	}

	b, err := json.Marshal(supportedTimezones)
	if err != nil {
		c.Log.Warn("Unable to marshal JSON in timezones.", mlog.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(b)
}

func testS3(c *Context, w http.ResponseWriter, r *http.Request) {
	cfg := model.ConfigFromJson(r.Body)
	if cfg == nil {
		cfg = c.App.Config()
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("testS3", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	err := filesstore.CheckMandatoryS3Fields(&cfg.FileSettings)
	if err != nil {
		c.Err = err
		return
	}

	if *cfg.FileSettings.AmazonS3SecretAccessKey == model.FAKE_SETTING {
		cfg.FileSettings.AmazonS3SecretAccessKey = c.App.Config().FileSettings.AmazonS3SecretAccessKey
	}

	license := c.App.Srv().License()
	backend, appErr := filesstore.NewFileBackend(&cfg.FileSettings, license != nil && *license.Features.Compliance)
	if appErr == nil {
		appErr = backend.TestConnection()
	}
	if appErr != nil {
		c.Err = appErr
		return
	}

	ReturnStatusOK(w)
}

func getRedirectLocation(c *Context, w http.ResponseWriter, r *http.Request) {
	m := make(map[string]string)
	m["location"] = ""

	if !*c.App.Config().ServiceSettings.EnableLinkPreviews {
		w.Write([]byte(model.MapToJson(m)))
		return
	}

	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		c.SetInvalidParam("url")
		return
	}

	var location string
	if err := redirectLocationDataCache.Get(url, &location); err == nil {
		m["location"] = location
		w.Write([]byte(model.MapToJson(m)))
		return
	}

	client := c.App.HTTPService().MakeClient(false)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := client.Head(url)
	if err != nil {
		// Cache failures to prevent retries.
		redirectLocationDataCache.SetWithExpiry(url, "", 1*time.Hour)
		// Always return a success status and a JSON string to limit information returned to client.
		w.Write([]byte(model.MapToJson(m)))
		return
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	location = res.Header.Get("Location")
	redirectLocationDataCache.SetWithExpiry(url, location, 1*time.Hour)
	m["location"] = location

	w.Write([]byte(model.MapToJson(m)))
}

func pushNotificationAck(c *Context, w http.ResponseWriter, r *http.Request) {
	ack, err := model.PushNotificationAckFromJson(r.Body)
	if err != nil {
		c.Err = model.NewAppError("pushNotificationAck",
			"api.push_notifications_ack.message.parse.app_error",
			nil,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if !*c.App.Config().EmailSettings.SendPushNotifications {
		c.Err = model.NewAppError("pushNotificationAck", "api.push_notification.disabled.app_error", nil, "", http.StatusNotImplemented)
		return
	}

	err = c.App.SendAckToPushProxy(ack)
	if ack.IsIdLoaded {
		if err != nil {
			// Log the error only, then continue to fetch notification message
			c.App.NotificationsLog().Error("Notification ack not sent to push proxy",
				mlog.String("ackId", ack.Id),
				mlog.String("type", ack.NotificationType),
				mlog.String("postId", ack.PostId),
				mlog.String("status", err.Error()),
			)
		}

		notificationInterface := c.App.Notification()

		if notificationInterface == nil {
			c.Err = model.NewAppError("pushNotificationAck", "api.system.id_loaded.not_available.app_error", nil, "", http.StatusFound)
			return
		}

		msg, appError := notificationInterface.GetNotificationMessage(ack, c.App.Session().UserId)
		if appError != nil {
			c.Err = model.NewAppError("pushNotificationAck", "api.push_notification.id_loaded.fetch.app_error", nil, appError.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(msg.ToJson()))

		return
	} else if err != nil {
		c.Err = model.NewAppError("pushNotificationAck", "api.push_notifications_ack.forward.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnStatusOK(w)
}

func setServerBusy(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	// number of seconds to keep server marked busy
	secs := r.URL.Query().Get("seconds")
	if secs == "" {
		secs = strconv.FormatInt(DEFAULT_SERVER_BUSY_SECONDS, 10)
	}

	i, err := strconv.ParseInt(secs, 10, 64)
	if err != nil || i <= 0 || i > MAX_SERVER_BUSY_SECONDS {
		c.SetInvalidUrlParam(fmt.Sprintf("seconds must be 1 - %d", MAX_SERVER_BUSY_SECONDS))
		return
	}

	auditRec := c.MakeAuditRecord("setServerBusy", audit.Fail)
	defer c.LogAuditRec(auditRec)
	auditRec.AddMeta("seconds", i)

	c.App.Srv().Busy.Set(time.Second * time.Duration(i))
	mlog.Warn("server busy state activated - non-critical services disabled", mlog.Int64("seconds", i))

	auditRec.Success()
	ReturnStatusOK(w)
}

func clearServerBusy(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	auditRec := c.MakeAuditRecord("clearServerBusy", audit.Fail)
	defer c.LogAuditRec(auditRec)

	c.App.Srv().Busy.Clear()
	mlog.Info("server busy state cleared - non-critical services enabled")

	auditRec.Success()
	ReturnStatusOK(w)
}

func getServerBusyExpires(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}
	w.Write([]byte(c.App.Srv().Busy.ToJson()))
}

func upgradeToEnterprise(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("upgradeToEnterprise", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	if model.BuildEnterpriseReady == "true" {
		c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.already-enterprise.app_error", nil, "", http.StatusTooManyRequests)
		return
	}

	percentage, _ := c.App.Srv().UpgradeToE0Status()

	if percentage > 0 {
		c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.app_error", nil, "", http.StatusTooManyRequests)
		return
	}
	if percentage == 100 {
		c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.already-done.app_error", nil, "", http.StatusTooManyRequests)
		return
	}

	if err := c.App.Srv().CanIUpgradeToE0(); err != nil {
		var ipErr *upgrader.InvalidPermissions
		var iaErr *upgrader.InvalidArch
		switch {
		case errors.As(err, &ipErr):
			params := map[string]interface{}{
				"MattermostUsername": ipErr.MattermostUsername,
				"FileUsername":       ipErr.FileUsername,
				"Path":               ipErr.Path,
			}
			if ipErr.ErrType == "invalid-user-and-permission" {
				c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.invalid-user-and-permission.app_error", params, err.Error(), http.StatusForbidden)
			} else if ipErr.ErrType == "invalid-user" {
				c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.invalid-user.app_error", params, err.Error(), http.StatusForbidden)
			} else if ipErr.ErrType == "invalid-permission" {
				c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.invalid-permission.app_error", params, err.Error(), http.StatusForbidden)
			}
		case errors.As(err, &iaErr):
			c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.system_not_supported.app_error", nil, err.Error(), http.StatusForbidden)
		default:
			c.Err = model.NewAppError("upgradeToEnterprise", "api.upgrade_to_enterprise.generic_error.app_error", nil, err.Error(), http.StatusForbidden)
		}
		return
	}

	c.App.Srv().Go(func() {
		c.App.Srv().UpgradeToE0()
	})

	auditRec.Success()
	w.WriteHeader(http.StatusAccepted)
	ReturnStatusOK(w)
}

func upgradeToEnterpriseStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	percentage, err := c.App.Srv().UpgradeToE0Status()
	var s map[string]interface{}
	if err != nil {
		var isErr *upgrader.InvalidSignature
		switch {
		case errors.As(err, &isErr):
			appErr := model.NewAppError("upgradeToEnterpriseStatus", "api.upgrade_to_enterprise_status.app_error", nil, err.Error(), http.StatusBadRequest)
			s = map[string]interface{}{"percentage": 0, "error": appErr.Message}
		default:
			appErr := model.NewAppError("upgradeToEnterpriseStatus", "api.upgrade_to_enterprise_status.signature.app_error", nil, err.Error(), http.StatusBadRequest)
			s = map[string]interface{}{"percentage": 0, "error": appErr.Message}
		}
	} else {
		s = map[string]interface{}{"percentage": percentage, "error": nil}
	}

	w.Write([]byte(model.StringInterfaceToJson(s)))
}

func restart(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("restartServer", audit.Fail)
	defer c.LogAuditRec(auditRec)

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
	time.Sleep(1 * time.Second)

	go func() {
		c.App.Srv().Restart()
	}()
}

func getWarnMetricsStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleReadPermissions) {
		c.SetPermissionError(model.SysconsoleReadPermissions...)
		return
	}

	license := c.App.Srv().License()
	if license != nil {
		mlog.Debug("License is present, skip.")
		return
	}

	status, err := c.App.GetWarnMetricsStatus()
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.MapWarnMetricStatusToJson(status)))
}

func sendWarnMetricAckEmail(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("sendWarnMetricAckEmail", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	license := c.App.Srv().License()
	if license != nil {
		mlog.Debug("License is present, skip.")
		return
	}

	user, appErr := c.App.GetUser(c.App.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	ack := model.SendWarnMetricAckFromJson(r.Body)
	if ack == nil {
		c.SetInvalidParam("ack")
		return
	}

	appErr = c.App.NotifyAndSetWarnMetricAck(c.Params.WarnMetricId, user, ack.ForceAck, false)
	if appErr != nil {
		c.Err = appErr
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

func requestTrialLicenseAndAckWarnMetric(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("requestTrialLicenseAndAckWarnMetric", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	if model.BuildEnterpriseReady != "true" {
		mlog.Debug("Not Enterprise Edition, skip.")
		return
	}

	license := c.App.Srv().License()
	if license != nil {
		mlog.Debug("License is present, skip.")
		return
	}

	if err := c.App.RequestLicenseAndAckWarnMetric(c.Params.WarnMetricId, false); err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

func getProductNotices(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	client, parseError := model.NoticeClientTypeFromString(r.URL.Query().Get("client"))
	if parseError != nil {
		c.SetInvalidParam("client")
		return
	}
	clientVersion := r.URL.Query().Get("clientVersion")
	locale := r.URL.Query().Get("locale")

	notices, err := c.App.GetProductNotices(c.App.Session().UserId, c.Params.TeamId, client, clientVersion, locale)

	if err != nil {
		c.Err = err
		return
	}
	result, _ := notices.Marshal()
	_, _ = w.Write(result)
}

func updateViewedProductNotices(c *Context, w http.ResponseWriter, r *http.Request) {
	auditRec := c.MakeAuditRecord("updateViewedProductNotices", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	ids := model.ArrayFromJson(r.Body)
	err := c.App.UpdateViewedProductNotices(c.App.Session().UserId, ids)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

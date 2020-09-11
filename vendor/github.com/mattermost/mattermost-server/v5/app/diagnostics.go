// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/marketplace"
	"github.com/mattermost/mattermost-server/v5/store"
	rudder "github.com/rudderlabs/analytics-go"
)

const (
	RUDDER_KEY           = "placeholder_rudder_key"
	RUDDER_DATAPLANE_URL = "placeholder_rudder_dataplane_url"

	TRACK_CONFIG_SERVICE            = "config_service"
	TRACK_CONFIG_TEAM               = "config_team"
	TRACK_CONFIG_CLIENT_REQ         = "config_client_requirements"
	TRACK_CONFIG_SQL                = "config_sql"
	TRACK_CONFIG_LOG                = "config_log"
	TRACK_CONFIG_AUDIT              = "config_audit"
	TRACK_CONFIG_NOTIFICATION_LOG   = "config_notifications_log"
	TRACK_CONFIG_FILE               = "config_file"
	TRACK_CONFIG_RATE               = "config_rate"
	TRACK_CONFIG_EMAIL              = "config_email"
	TRACK_CONFIG_PRIVACY            = "config_privacy"
	TRACK_CONFIG_THEME              = "config_theme"
	TRACK_CONFIG_OAUTH              = "config_oauth"
	TRACK_CONFIG_LDAP               = "config_ldap"
	TRACK_CONFIG_COMPLIANCE         = "config_compliance"
	TRACK_CONFIG_LOCALIZATION       = "config_localization"
	TRACK_CONFIG_SAML               = "config_saml"
	TRACK_CONFIG_PASSWORD           = "config_password"
	TRACK_CONFIG_CLUSTER            = "config_cluster"
	TRACK_CONFIG_METRICS            = "config_metrics"
	TRACK_CONFIG_SUPPORT            = "config_support"
	TRACK_CONFIG_NATIVEAPP          = "config_nativeapp"
	TRACK_CONFIG_EXPERIMENTAL       = "config_experimental"
	TRACK_CONFIG_ANALYTICS          = "config_analytics"
	TRACK_CONFIG_ANNOUNCEMENT       = "config_announcement"
	TRACK_CONFIG_ELASTICSEARCH      = "config_elasticsearch"
	TRACK_CONFIG_PLUGIN             = "config_plugin"
	TRACK_CONFIG_DATA_RETENTION     = "config_data_retention"
	TRACK_CONFIG_MESSAGE_EXPORT     = "config_message_export"
	TRACK_CONFIG_DISPLAY            = "config_display"
	TRACK_CONFIG_GUEST_ACCOUNTS     = "config_guest_accounts"
	TRACK_CONFIG_IMAGE_PROXY        = "config_image_proxy"
	TRACK_CONFIG_BLEVE              = "config_bleve"
	TRACK_PERMISSIONS_GENERAL       = "permissions_general"
	TRACK_PERMISSIONS_SYSTEM_SCHEME = "permissions_system_scheme"
	TRACK_PERMISSIONS_TEAM_SCHEMES  = "permissions_team_schemes"
	TRACK_ELASTICSEARCH             = "elasticsearch"
	TRACK_GROUPS                    = "groups"
	TRACK_CHANNEL_MODERATION        = "channel_moderation"
	TRACK_WARN_METRICS              = "warn_metrics"

	TRACK_ACTIVITY = "activity"
	TRACK_LICENSE  = "license"
	TRACK_SERVER   = "server"
	TRACK_PLUGINS  = "plugins"
)

// declaring this as var to allow overriding in tests
var SENTRY_DSN = "placeholder_sentry_dsn"

type RudderConfig struct {
	RudderKey    string
	DataplaneUrl string
}

func (s *Server) SendDailyDiagnostics() {
	s.sendDailyDiagnostics(false)
}

func (s *Server) getRudderConfig() RudderConfig {
	if !strings.Contains(RUDDER_KEY, "placeholder") && !strings.Contains(RUDDER_DATAPLANE_URL, "placeholder") {
		return RudderConfig{RUDDER_KEY, RUDDER_DATAPLANE_URL}
	} else if os.Getenv("RUDDER_KEY") != "" && os.Getenv("RUDDER_DATAPLANE_URL") != "" {
		return RudderConfig{os.Getenv("RUDDER_KEY"), os.Getenv("RUDDER_DATAPLANE_URL")}
	} else {
		return RudderConfig{}
	}
}

func (s *Server) diagnosticsEnabled() bool {
	return *s.Config().LogSettings.EnableDiagnostics && s.IsLeader()
}

func (s *Server) sendDailyDiagnostics(override bool) {
	config := s.getRudderConfig()
	if s.diagnosticsEnabled() && ((config.DataplaneUrl != "" && config.RudderKey != "") || override) {
		s.initDiagnostics(config.DataplaneUrl, config.RudderKey)
		s.trackActivity()
		s.trackConfig()
		s.trackLicense()
		s.trackPlugins()
		s.trackServer()
		s.trackPermissions()
		s.trackElasticsearch()
		s.trackGroups()
		s.trackChannelModeration()
		s.trackWarnMetrics()
	}
}

func (s *Server) SendDiagnostic(event string, properties map[string]interface{}) {
	if s.rudderClient != nil {
		s.rudderClient.Enqueue(rudder.Track{
			Event:      event,
			UserId:     s.diagnosticId,
			Properties: properties,
		})
	}
}

func isDefault(setting interface{}, defaultValue interface{}) bool {
	return setting == defaultValue
}

func pluginSetting(pluginSettings *model.PluginSettings, plugin, key string, defaultValue interface{}) interface{} {
	settings, ok := pluginSettings.Plugins[plugin]
	if !ok {
		return defaultValue
	}
	if value, ok := settings[key]; ok {
		return value
	}
	return defaultValue
}

func pluginActivated(pluginStates map[string]*model.PluginState, pluginId string) bool {
	state, ok := pluginStates[pluginId]
	if !ok {
		return false
	}
	return state.Enable
}

func pluginVersion(pluginsAvailable []*model.BundleInfo, pluginId string) string {
	for _, plugin := range pluginsAvailable {
		if plugin.Manifest != nil && plugin.Manifest.Id == pluginId {
			return plugin.Manifest.Version
		}
	}
	return ""
}

func (s *Server) trackActivity() {
	var userCount int64
	var guestAccountsCount int64
	var botAccountsCount int64
	var inactiveUserCount int64
	var publicChannelCount int64
	var privateChannelCount int64
	var directChannelCount int64
	var deletedPublicChannelCount int64
	var deletedPrivateChannelCount int64
	var postsCount int64
	var postsCountPreviousDay int64
	var botPostsCountPreviousDay int64
	var slashCommandsCount int64
	var incomingWebhooksCount int64
	var outgoingWebhooksCount int64

	activeUsersDailyCountChan := make(chan store.StoreResult, 1)
	go func() {
		count, err := s.Store.User().AnalyticsActiveCount(DAY_MILLISECONDS, model.UserCountOptions{IncludeBotAccounts: false, IncludeDeleted: false})
		activeUsersDailyCountChan <- store.StoreResult{Data: count, Err: err}
		close(activeUsersDailyCountChan)
	}()

	activeUsersMonthlyCountChan := make(chan store.StoreResult, 1)
	go func() {
		count, err := s.Store.User().AnalyticsActiveCount(MONTH_MILLISECONDS, model.UserCountOptions{IncludeBotAccounts: false, IncludeDeleted: false})
		activeUsersMonthlyCountChan <- store.StoreResult{Data: count, Err: err}
		close(activeUsersMonthlyCountChan)
	}()

	if count, err := s.Store.User().Count(model.UserCountOptions{IncludeDeleted: true}); err == nil {
		userCount = count
	}

	if count, err := s.Store.User().AnalyticsGetGuestCount(); err == nil {
		guestAccountsCount = count
	}

	if count, err := s.Store.User().Count(model.UserCountOptions{IncludeBotAccounts: true, ExcludeRegularUsers: true}); err == nil {
		botAccountsCount = count
	}

	if iucr, err := s.Store.User().AnalyticsGetInactiveUsersCount(); err == nil {
		inactiveUserCount = iucr
	}

	teamCount, err := s.Store.Team().AnalyticsTeamCount(false)
	if err != nil {
		mlog.Error(err.Error())
	}

	if ucc, err := s.Store.Channel().AnalyticsTypeCount("", "O"); err == nil {
		publicChannelCount = ucc
	}

	if pcc, err := s.Store.Channel().AnalyticsTypeCount("", "P"); err == nil {
		privateChannelCount = pcc
	}

	if dcc, err := s.Store.Channel().AnalyticsTypeCount("", "D"); err == nil {
		directChannelCount = dcc
	}

	if duccr, err := s.Store.Channel().AnalyticsDeletedTypeCount("", "O"); err == nil {
		deletedPublicChannelCount = duccr
	}

	if dpccr, err := s.Store.Channel().AnalyticsDeletedTypeCount("", "P"); err == nil {
		deletedPrivateChannelCount = dpccr
	}

	postsCount, _ = s.Store.Post().AnalyticsPostCount("", false, false)

	postCountsOptions := &model.AnalyticsPostCountsOptions{TeamId: "", BotsOnly: false, YesterdayOnly: true}
	postCountsYesterday, _ := s.Store.Post().AnalyticsPostCountsByDay(postCountsOptions)
	postsCountPreviousDay = 0
	if len(postCountsYesterday) > 0 {
		postsCountPreviousDay = int64(postCountsYesterday[0].Value)
	}

	postCountsOptions = &model.AnalyticsPostCountsOptions{TeamId: "", BotsOnly: true, YesterdayOnly: true}
	botPostCountsYesterday, _ := s.Store.Post().AnalyticsPostCountsByDay(postCountsOptions)
	botPostsCountPreviousDay = 0
	if len(botPostCountsYesterday) > 0 {
		botPostsCountPreviousDay = int64(botPostCountsYesterday[0].Value)
	}

	slashCommandsCount, _ = s.Store.Command().AnalyticsCommandCount("")

	if c, err := s.Store.Webhook().AnalyticsIncomingCount(""); err == nil {
		incomingWebhooksCount = c
	}

	outgoingWebhooksCount, _ = s.Store.Webhook().AnalyticsOutgoingCount("")

	var activeUsersDailyCount int64
	if r := <-activeUsersDailyCountChan; r.Err == nil {
		activeUsersDailyCount = r.Data.(int64)
	}

	var activeUsersMonthlyCount int64
	if r := <-activeUsersMonthlyCountChan; r.Err == nil {
		activeUsersMonthlyCount = r.Data.(int64)
	}

	s.SendDiagnostic(TRACK_ACTIVITY, map[string]interface{}{
		"registered_users":             userCount,
		"bot_accounts":                 botAccountsCount,
		"guest_accounts":               guestAccountsCount,
		"active_users_daily":           activeUsersDailyCount,
		"active_users_monthly":         activeUsersMonthlyCount,
		"registered_deactivated_users": inactiveUserCount,
		"teams":                        teamCount,
		"public_channels":              publicChannelCount,
		"private_channels":             privateChannelCount,
		"direct_message_channels":      directChannelCount,
		"public_channels_deleted":      deletedPublicChannelCount,
		"private_channels_deleted":     deletedPrivateChannelCount,
		"posts_previous_day":           postsCountPreviousDay,
		"bot_posts_previous_day":       botPostsCountPreviousDay,
		"posts":                        postsCount,
		"slash_commands":               slashCommandsCount,
		"incoming_webhooks":            incomingWebhooksCount,
		"outgoing_webhooks":            outgoingWebhooksCount,
	})
}

func (s *Server) trackConfig() {
	cfg := s.Config()
	s.SendDiagnostic(TRACK_CONFIG_SERVICE, map[string]interface{}{
		"web_server_mode":                                         *cfg.ServiceSettings.WebserverMode,
		"enable_security_fix_alert":                               *cfg.ServiceSettings.EnableSecurityFixAlert,
		"enable_insecure_outgoing_connections":                    *cfg.ServiceSettings.EnableInsecureOutgoingConnections,
		"enable_incoming_webhooks":                                cfg.ServiceSettings.EnableIncomingWebhooks,
		"enable_outgoing_webhooks":                                cfg.ServiceSettings.EnableOutgoingWebhooks,
		"enable_commands":                                         *cfg.ServiceSettings.EnableCommands,
		"enable_only_admin_integrations":                          *cfg.ServiceSettings.DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations,
		"enable_post_username_override":                           cfg.ServiceSettings.EnablePostUsernameOverride,
		"enable_post_icon_override":                               cfg.ServiceSettings.EnablePostIconOverride,
		"enable_user_access_tokens":                               *cfg.ServiceSettings.EnableUserAccessTokens,
		"enable_custom_emoji":                                     *cfg.ServiceSettings.EnableCustomEmoji,
		"enable_emoji_picker":                                     *cfg.ServiceSettings.EnableEmojiPicker,
		"enable_gif_picker":                                       *cfg.ServiceSettings.EnableGifPicker,
		"gfycat_api_key":                                          isDefault(*cfg.ServiceSettings.GfycatApiKey, model.SERVICE_SETTINGS_DEFAULT_GFYCAT_API_KEY),
		"gfycat_api_secret":                                       isDefault(*cfg.ServiceSettings.GfycatApiSecret, model.SERVICE_SETTINGS_DEFAULT_GFYCAT_API_SECRET),
		"experimental_enable_authentication_transfer":             *cfg.ServiceSettings.ExperimentalEnableAuthenticationTransfer,
		"restrict_custom_emoji_creation":                          *cfg.ServiceSettings.DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation,
		"enable_testing":                                          cfg.ServiceSettings.EnableTesting,
		"enable_developer":                                        *cfg.ServiceSettings.EnableDeveloper,
		"enable_multifactor_authentication":                       *cfg.ServiceSettings.EnableMultifactorAuthentication,
		"enforce_multifactor_authentication":                      *cfg.ServiceSettings.EnforceMultifactorAuthentication,
		"enable_oauth_service_provider":                           cfg.ServiceSettings.EnableOAuthServiceProvider,
		"connection_security":                                     *cfg.ServiceSettings.ConnectionSecurity,
		"tls_strict_transport":                                    *cfg.ServiceSettings.TLSStrictTransport,
		"uses_letsencrypt":                                        *cfg.ServiceSettings.UseLetsEncrypt,
		"forward_80_to_443":                                       *cfg.ServiceSettings.Forward80To443,
		"maximum_login_attempts":                                  *cfg.ServiceSettings.MaximumLoginAttempts,
		"extend_session_length_with_activity":                     *cfg.ServiceSettings.ExtendSessionLengthWithActivity,
		"session_length_web_in_days":                              *cfg.ServiceSettings.SessionLengthWebInDays,
		"session_length_mobile_in_days":                           *cfg.ServiceSettings.SessionLengthMobileInDays,
		"session_length_sso_in_days":                              *cfg.ServiceSettings.SessionLengthSSOInDays,
		"session_cache_in_minutes":                                *cfg.ServiceSettings.SessionCacheInMinutes,
		"session_idle_timeout_in_minutes":                         *cfg.ServiceSettings.SessionIdleTimeoutInMinutes,
		"isdefault_site_url":                                      isDefault(*cfg.ServiceSettings.SiteURL, model.SERVICE_SETTINGS_DEFAULT_SITE_URL),
		"isdefault_tls_cert_file":                                 isDefault(*cfg.ServiceSettings.TLSCertFile, model.SERVICE_SETTINGS_DEFAULT_TLS_CERT_FILE),
		"isdefault_tls_key_file":                                  isDefault(*cfg.ServiceSettings.TLSKeyFile, model.SERVICE_SETTINGS_DEFAULT_TLS_KEY_FILE),
		"isdefault_read_timeout":                                  isDefault(*cfg.ServiceSettings.ReadTimeout, model.SERVICE_SETTINGS_DEFAULT_READ_TIMEOUT),
		"isdefault_write_timeout":                                 isDefault(*cfg.ServiceSettings.WriteTimeout, model.SERVICE_SETTINGS_DEFAULT_WRITE_TIMEOUT),
		"isdefault_idle_timeout":                                  isDefault(*cfg.ServiceSettings.IdleTimeout, model.SERVICE_SETTINGS_DEFAULT_IDLE_TIMEOUT),
		"isdefault_google_developer_key":                          isDefault(cfg.ServiceSettings.GoogleDeveloperKey, ""),
		"isdefault_allow_cors_from":                               isDefault(*cfg.ServiceSettings.AllowCorsFrom, model.SERVICE_SETTINGS_DEFAULT_ALLOW_CORS_FROM),
		"isdefault_cors_exposed_headers":                          isDefault(cfg.ServiceSettings.CorsExposedHeaders, ""),
		"cors_allow_credentials":                                  *cfg.ServiceSettings.CorsAllowCredentials,
		"cors_debug":                                              *cfg.ServiceSettings.CorsDebug,
		"isdefault_allowed_untrusted_internal_connections":        isDefault(*cfg.ServiceSettings.AllowedUntrustedInternalConnections, ""),
		"restrict_post_delete":                                    *cfg.ServiceSettings.DEPRECATED_DO_NOT_USE_RestrictPostDelete,
		"allow_edit_post":                                         *cfg.ServiceSettings.DEPRECATED_DO_NOT_USE_AllowEditPost,
		"post_edit_time_limit":                                    *cfg.ServiceSettings.PostEditTimeLimit,
		"enable_user_typing_messages":                             *cfg.ServiceSettings.EnableUserTypingMessages,
		"enable_channel_viewed_messages":                          *cfg.ServiceSettings.EnableChannelViewedMessages,
		"time_between_user_typing_updates_milliseconds":           *cfg.ServiceSettings.TimeBetweenUserTypingUpdatesMilliseconds,
		"cluster_log_timeout_milliseconds":                        *cfg.ServiceSettings.ClusterLogTimeoutMilliseconds,
		"enable_post_search":                                      *cfg.ServiceSettings.EnablePostSearch,
		"minimum_hashtag_length":                                  *cfg.ServiceSettings.MinimumHashtagLength,
		"enable_user_statuses":                                    *cfg.ServiceSettings.EnableUserStatuses,
		"close_unused_direct_messages":                            *cfg.ServiceSettings.CloseUnusedDirectMessages,
		"enable_preview_features":                                 *cfg.ServiceSettings.EnablePreviewFeatures,
		"enable_tutorial":                                         *cfg.ServiceSettings.EnableTutorial,
		"experimental_enable_default_channel_leave_join_messages": *cfg.ServiceSettings.ExperimentalEnableDefaultChannelLeaveJoinMessages,
		"experimental_group_unread_channels":                      *cfg.ServiceSettings.ExperimentalGroupUnreadChannels,
		"websocket_url":                                           isDefault(*cfg.ServiceSettings.WebsocketURL, ""),
		"allow_cookies_for_subdomains":                            *cfg.ServiceSettings.AllowCookiesForSubdomains,
		"enable_api_team_deletion":                                *cfg.ServiceSettings.EnableAPITeamDeletion,
		"experimental_enable_hardened_mode":                       *cfg.ServiceSettings.ExperimentalEnableHardenedMode,
		"disable_legacy_mfa":                                      *cfg.ServiceSettings.DisableLegacyMFA,
		"experimental_strict_csrf_enforcement":                    *cfg.ServiceSettings.ExperimentalStrictCSRFEnforcement,
		"enable_email_invitations":                                *cfg.ServiceSettings.EnableEmailInvitations,
		"experimental_channel_organization":                       *cfg.ServiceSettings.ExperimentalChannelOrganization,
		"experimental_channel_sidebar_organization":               *cfg.ServiceSettings.ExperimentalChannelSidebarOrganization,
		"disable_bots_when_owner_is_deactivated":                  *cfg.ServiceSettings.DisableBotsWhenOwnerIsDeactivated,
		"enable_bot_account_creation":                             *cfg.ServiceSettings.EnableBotAccountCreation,
		"enable_svgs":                                             *cfg.ServiceSettings.EnableSVGs,
		"enable_latex":                                            *cfg.ServiceSettings.EnableLatex,
		"enable_opentracing":                                      *cfg.ServiceSettings.EnableOpenTracing,
		"experimental_data_prefetch":                              *cfg.ServiceSettings.ExperimentalDataPrefetch,
		"enable_local_mode":                                       *cfg.ServiceSettings.EnableLocalMode,
	})

	s.SendDiagnostic(TRACK_CONFIG_TEAM, map[string]interface{}{
		"enable_user_creation":                      cfg.TeamSettings.EnableUserCreation,
		"enable_team_creation":                      *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_EnableTeamCreation,
		"restrict_team_invite":                      *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictTeamInvite,
		"restrict_public_channel_creation":          *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation,
		"restrict_private_channel_creation":         *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation,
		"restrict_public_channel_management":        *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement,
		"restrict_private_channel_management":       *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement,
		"restrict_public_channel_deletion":          *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion,
		"restrict_private_channel_deletion":         *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion,
		"enable_open_server":                        *cfg.TeamSettings.EnableOpenServer,
		"enable_user_deactivation":                  *cfg.TeamSettings.EnableUserDeactivation,
		"enable_custom_brand":                       *cfg.TeamSettings.EnableCustomBrand,
		"restrict_direct_message":                   *cfg.TeamSettings.RestrictDirectMessage,
		"max_notifications_per_channel":             *cfg.TeamSettings.MaxNotificationsPerChannel,
		"enable_confirm_notifications_to_channel":   *cfg.TeamSettings.EnableConfirmNotificationsToChannel,
		"max_users_per_team":                        *cfg.TeamSettings.MaxUsersPerTeam,
		"max_channels_per_team":                     *cfg.TeamSettings.MaxChannelsPerTeam,
		"teammate_name_display":                     *cfg.TeamSettings.TeammateNameDisplay,
		"experimental_view_archived_channels":       *cfg.TeamSettings.ExperimentalViewArchivedChannels,
		"lock_teammate_name_display":                *cfg.TeamSettings.LockTeammateNameDisplay,
		"isdefault_site_name":                       isDefault(cfg.TeamSettings.SiteName, "Mattermost"),
		"isdefault_custom_brand_text":               isDefault(*cfg.TeamSettings.CustomBrandText, model.TEAM_SETTINGS_DEFAULT_CUSTOM_BRAND_TEXT),
		"isdefault_custom_description_text":         isDefault(*cfg.TeamSettings.CustomDescriptionText, model.TEAM_SETTINGS_DEFAULT_CUSTOM_DESCRIPTION_TEXT),
		"isdefault_user_status_away_timeout":        isDefault(*cfg.TeamSettings.UserStatusAwayTimeout, model.TEAM_SETTINGS_DEFAULT_USER_STATUS_AWAY_TIMEOUT),
		"restrict_private_channel_manage_members":   *cfg.TeamSettings.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers,
		"enable_X_to_leave_channels_from_LHS":       *cfg.TeamSettings.EnableXToLeaveChannelsFromLHS,
		"experimental_enable_automatic_replies":     *cfg.TeamSettings.ExperimentalEnableAutomaticReplies,
		"experimental_town_square_is_hidden_in_lhs": *cfg.TeamSettings.ExperimentalHideTownSquareinLHS,
		"experimental_town_square_is_read_only":     *cfg.TeamSettings.ExperimentalTownSquareIsReadOnly,
		"experimental_primary_team":                 isDefault(*cfg.TeamSettings.ExperimentalPrimaryTeam, ""),
		"experimental_default_channels":             len(cfg.TeamSettings.ExperimentalDefaultChannels),
	})

	s.SendDiagnostic(TRACK_CONFIG_CLIENT_REQ, map[string]interface{}{
		"android_latest_version": cfg.ClientRequirements.AndroidLatestVersion,
		"android_min_version":    cfg.ClientRequirements.AndroidMinVersion,
		"desktop_latest_version": cfg.ClientRequirements.DesktopLatestVersion,
		"desktop_min_version":    cfg.ClientRequirements.DesktopMinVersion,
		"ios_latest_version":     cfg.ClientRequirements.IosLatestVersion,
		"ios_min_version":        cfg.ClientRequirements.IosMinVersion,
	})

	s.SendDiagnostic(TRACK_CONFIG_SQL, map[string]interface{}{
		"driver_name":                    *cfg.SqlSettings.DriverName,
		"trace":                          cfg.SqlSettings.Trace,
		"max_idle_conns":                 *cfg.SqlSettings.MaxIdleConns,
		"conn_max_lifetime_milliseconds": *cfg.SqlSettings.ConnMaxLifetimeMilliseconds,
		"max_open_conns":                 *cfg.SqlSettings.MaxOpenConns,
		"data_source_replicas":           len(cfg.SqlSettings.DataSourceReplicas),
		"data_source_search_replicas":    len(cfg.SqlSettings.DataSourceSearchReplicas),
		"query_timeout":                  *cfg.SqlSettings.QueryTimeout,
		"disable_database_search":        *cfg.SqlSettings.DisableDatabaseSearch,
	})

	s.SendDiagnostic(TRACK_CONFIG_LOG, map[string]interface{}{
		"enable_console":           cfg.LogSettings.EnableConsole,
		"console_level":            cfg.LogSettings.ConsoleLevel,
		"console_json":             *cfg.LogSettings.ConsoleJson,
		"enable_file":              cfg.LogSettings.EnableFile,
		"file_level":               cfg.LogSettings.FileLevel,
		"file_json":                cfg.LogSettings.FileJson,
		"enable_webhook_debugging": cfg.LogSettings.EnableWebhookDebugging,
		"isdefault_file_location":  isDefault(cfg.LogSettings.FileLocation, ""),
		"advanced_logging_config":  *cfg.LogSettings.AdvancedLoggingConfig != "",
	})

	s.SendDiagnostic(TRACK_CONFIG_AUDIT, map[string]interface{}{
		"file_enabled":            *cfg.ExperimentalAuditSettings.FileEnabled,
		"file_max_size_mb":        *cfg.ExperimentalAuditSettings.FileMaxSizeMB,
		"file_max_age_days":       *cfg.ExperimentalAuditSettings.FileMaxAgeDays,
		"file_max_backups":        *cfg.ExperimentalAuditSettings.FileMaxBackups,
		"file_compress":           *cfg.ExperimentalAuditSettings.FileCompress,
		"file_max_queue_size":     *cfg.ExperimentalAuditSettings.FileMaxQueueSize,
		"advanced_logging_config": *cfg.ExperimentalAuditSettings.AdvancedLoggingConfig != "",
	})

	s.SendDiagnostic(TRACK_CONFIG_NOTIFICATION_LOG, map[string]interface{}{
		"enable_console":          *cfg.NotificationLogSettings.EnableConsole,
		"console_level":           *cfg.NotificationLogSettings.ConsoleLevel,
		"console_json":            *cfg.NotificationLogSettings.ConsoleJson,
		"enable_file":             *cfg.NotificationLogSettings.EnableFile,
		"file_level":              *cfg.NotificationLogSettings.FileLevel,
		"file_json":               *cfg.NotificationLogSettings.FileJson,
		"isdefault_file_location": isDefault(*cfg.NotificationLogSettings.FileLocation, ""),
		"advanced_logging_config": *cfg.NotificationLogSettings.AdvancedLoggingConfig != "",
	})

	s.SendDiagnostic(TRACK_CONFIG_PASSWORD, map[string]interface{}{
		"minimum_length": *cfg.PasswordSettings.MinimumLength,
		"lowercase":      *cfg.PasswordSettings.Lowercase,
		"number":         *cfg.PasswordSettings.Number,
		"uppercase":      *cfg.PasswordSettings.Uppercase,
		"symbol":         *cfg.PasswordSettings.Symbol,
	})

	s.SendDiagnostic(TRACK_CONFIG_FILE, map[string]interface{}{
		"enable_public_links":     cfg.FileSettings.EnablePublicLink,
		"driver_name":             *cfg.FileSettings.DriverName,
		"isdefault_directory":     isDefault(*cfg.FileSettings.Directory, model.FILE_SETTINGS_DEFAULT_DIRECTORY),
		"isabsolute_directory":    filepath.IsAbs(*cfg.FileSettings.Directory),
		"amazon_s3_ssl":           *cfg.FileSettings.AmazonS3SSL,
		"amazon_s3_sse":           *cfg.FileSettings.AmazonS3SSE,
		"amazon_s3_signv2":        *cfg.FileSettings.AmazonS3SignV2,
		"amazon_s3_trace":         *cfg.FileSettings.AmazonS3Trace,
		"max_file_size":           *cfg.FileSettings.MaxFileSize,
		"enable_file_attachments": *cfg.FileSettings.EnableFileAttachments,
		"enable_mobile_upload":    *cfg.FileSettings.EnableMobileUpload,
		"enable_mobile_download":  *cfg.FileSettings.EnableMobileDownload,
	})

	s.SendDiagnostic(TRACK_CONFIG_EMAIL, map[string]interface{}{
		"enable_sign_up_with_email":            cfg.EmailSettings.EnableSignUpWithEmail,
		"enable_sign_in_with_email":            *cfg.EmailSettings.EnableSignInWithEmail,
		"enable_sign_in_with_username":         *cfg.EmailSettings.EnableSignInWithUsername,
		"require_email_verification":           cfg.EmailSettings.RequireEmailVerification,
		"send_email_notifications":             cfg.EmailSettings.SendEmailNotifications,
		"use_channel_in_email_notifications":   *cfg.EmailSettings.UseChannelInEmailNotifications,
		"email_notification_contents_type":     *cfg.EmailSettings.EmailNotificationContentsType,
		"enable_smtp_auth":                     *cfg.EmailSettings.EnableSMTPAuth,
		"connection_security":                  cfg.EmailSettings.ConnectionSecurity,
		"send_push_notifications":              *cfg.EmailSettings.SendPushNotifications,
		"push_notification_contents":           *cfg.EmailSettings.PushNotificationContents,
		"enable_email_batching":                *cfg.EmailSettings.EnableEmailBatching,
		"email_batching_buffer_size":           *cfg.EmailSettings.EmailBatchingBufferSize,
		"email_batching_interval":              *cfg.EmailSettings.EmailBatchingInterval,
		"enable_preview_mode_banner":           *cfg.EmailSettings.EnablePreviewModeBanner,
		"isdefault_feedback_name":              isDefault(cfg.EmailSettings.FeedbackName, ""),
		"isdefault_feedback_email":             isDefault(cfg.EmailSettings.FeedbackEmail, ""),
		"isdefault_reply_to_address":           isDefault(cfg.EmailSettings.ReplyToAddress, ""),
		"isdefault_feedback_organization":      isDefault(*cfg.EmailSettings.FeedbackOrganization, model.EMAIL_SETTINGS_DEFAULT_FEEDBACK_ORGANIZATION),
		"skip_server_certificate_verification": *cfg.EmailSettings.SkipServerCertificateVerification,
		"isdefault_login_button_color":         isDefault(*cfg.EmailSettings.LoginButtonColor, ""),
		"isdefault_login_button_border_color":  isDefault(*cfg.EmailSettings.LoginButtonBorderColor, ""),
		"isdefault_login_button_text_color":    isDefault(*cfg.EmailSettings.LoginButtonTextColor, ""),
		"smtp_server_timeout":                  *cfg.EmailSettings.SMTPServerTimeout,
	})

	s.SendDiagnostic(TRACK_CONFIG_RATE, map[string]interface{}{
		"enable_rate_limiter":      *cfg.RateLimitSettings.Enable,
		"vary_by_remote_address":   *cfg.RateLimitSettings.VaryByRemoteAddr,
		"vary_by_user":             *cfg.RateLimitSettings.VaryByUser,
		"per_sec":                  *cfg.RateLimitSettings.PerSec,
		"max_burst":                *cfg.RateLimitSettings.MaxBurst,
		"memory_store_size":        *cfg.RateLimitSettings.MemoryStoreSize,
		"isdefault_vary_by_header": isDefault(cfg.RateLimitSettings.VaryByHeader, ""),
	})

	s.SendDiagnostic(TRACK_CONFIG_PRIVACY, map[string]interface{}{
		"show_email_address": cfg.PrivacySettings.ShowEmailAddress,
		"show_full_name":     cfg.PrivacySettings.ShowFullName,
	})

	s.SendDiagnostic(TRACK_CONFIG_THEME, map[string]interface{}{
		"enable_theme_selection":  *cfg.ThemeSettings.EnableThemeSelection,
		"isdefault_default_theme": isDefault(*cfg.ThemeSettings.DefaultTheme, model.TEAM_SETTINGS_DEFAULT_TEAM_TEXT),
		"allow_custom_themes":     *cfg.ThemeSettings.AllowCustomThemes,
		"allowed_themes":          len(cfg.ThemeSettings.AllowedThemes),
	})

	s.SendDiagnostic(TRACK_CONFIG_OAUTH, map[string]interface{}{
		"enable_gitlab":    cfg.GitLabSettings.Enable,
		"enable_google":    cfg.GoogleSettings.Enable,
		"enable_office365": cfg.Office365Settings.Enable,
	})

	s.SendDiagnostic(TRACK_CONFIG_SUPPORT, map[string]interface{}{
		"isdefault_terms_of_service_link":              isDefault(*cfg.SupportSettings.TermsOfServiceLink, model.SUPPORT_SETTINGS_DEFAULT_TERMS_OF_SERVICE_LINK),
		"isdefault_privacy_policy_link":                isDefault(*cfg.SupportSettings.PrivacyPolicyLink, model.SUPPORT_SETTINGS_DEFAULT_PRIVACY_POLICY_LINK),
		"isdefault_about_link":                         isDefault(*cfg.SupportSettings.AboutLink, model.SUPPORT_SETTINGS_DEFAULT_ABOUT_LINK),
		"isdefault_help_link":                          isDefault(*cfg.SupportSettings.HelpLink, model.SUPPORT_SETTINGS_DEFAULT_HELP_LINK),
		"isdefault_report_a_problem_link":              isDefault(*cfg.SupportSettings.ReportAProblemLink, model.SUPPORT_SETTINGS_DEFAULT_REPORT_A_PROBLEM_LINK),
		"isdefault_support_email":                      isDefault(*cfg.SupportSettings.SupportEmail, model.SUPPORT_SETTINGS_DEFAULT_SUPPORT_EMAIL),
		"custom_terms_of_service_enabled":              *cfg.SupportSettings.CustomTermsOfServiceEnabled,
		"custom_terms_of_service_re_acceptance_period": *cfg.SupportSettings.CustomTermsOfServiceReAcceptancePeriod,
		"enable_ask_community_link":                    *cfg.SupportSettings.EnableAskCommunityLink,
	})

	s.SendDiagnostic(TRACK_CONFIG_LDAP, map[string]interface{}{
		"enable":                                 *cfg.LdapSettings.Enable,
		"enable_sync":                            *cfg.LdapSettings.EnableSync,
		"enable_admin_filter":                    *cfg.LdapSettings.EnableAdminFilter,
		"connection_security":                    *cfg.LdapSettings.ConnectionSecurity,
		"skip_certificate_verification":          *cfg.LdapSettings.SkipCertificateVerification,
		"sync_interval_minutes":                  *cfg.LdapSettings.SyncIntervalMinutes,
		"query_timeout":                          *cfg.LdapSettings.QueryTimeout,
		"max_page_size":                          *cfg.LdapSettings.MaxPageSize,
		"isdefault_first_name_attribute":         isDefault(*cfg.LdapSettings.FirstNameAttribute, model.LDAP_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE),
		"isdefault_last_name_attribute":          isDefault(*cfg.LdapSettings.LastNameAttribute, model.LDAP_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE),
		"isdefault_email_attribute":              isDefault(*cfg.LdapSettings.EmailAttribute, model.LDAP_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE),
		"isdefault_username_attribute":           isDefault(*cfg.LdapSettings.UsernameAttribute, model.LDAP_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE),
		"isdefault_nickname_attribute":           isDefault(*cfg.LdapSettings.NicknameAttribute, model.LDAP_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE),
		"isdefault_id_attribute":                 isDefault(*cfg.LdapSettings.IdAttribute, model.LDAP_SETTINGS_DEFAULT_ID_ATTRIBUTE),
		"isdefault_position_attribute":           isDefault(*cfg.LdapSettings.PositionAttribute, model.LDAP_SETTINGS_DEFAULT_POSITION_ATTRIBUTE),
		"isdefault_login_id_attribute":           isDefault(*cfg.LdapSettings.LoginIdAttribute, ""),
		"isdefault_login_field_name":             isDefault(*cfg.LdapSettings.LoginFieldName, model.LDAP_SETTINGS_DEFAULT_LOGIN_FIELD_NAME),
		"isdefault_login_button_color":           isDefault(*cfg.LdapSettings.LoginButtonColor, ""),
		"isdefault_login_button_border_color":    isDefault(*cfg.LdapSettings.LoginButtonBorderColor, ""),
		"isdefault_login_button_text_color":      isDefault(*cfg.LdapSettings.LoginButtonTextColor, ""),
		"isempty_group_filter":                   isDefault(*cfg.LdapSettings.GroupFilter, ""),
		"isdefault_group_display_name_attribute": isDefault(*cfg.LdapSettings.GroupDisplayNameAttribute, model.LDAP_SETTINGS_DEFAULT_GROUP_DISPLAY_NAME_ATTRIBUTE),
		"isdefault_group_id_attribute":           isDefault(*cfg.LdapSettings.GroupIdAttribute, model.LDAP_SETTINGS_DEFAULT_GROUP_ID_ATTRIBUTE),
		"isempty_guest_filter":                   isDefault(*cfg.LdapSettings.GuestFilter, ""),
		"isempty_admin_filter":                   isDefault(*cfg.LdapSettings.AdminFilter, ""),
		"isnotempty_picture_attribute":           !isDefault(*cfg.LdapSettings.PictureAttribute, ""),
	})

	s.SendDiagnostic(TRACK_CONFIG_COMPLIANCE, map[string]interface{}{
		"enable":       *cfg.ComplianceSettings.Enable,
		"enable_daily": *cfg.ComplianceSettings.EnableDaily,
	})

	s.SendDiagnostic(TRACK_CONFIG_LOCALIZATION, map[string]interface{}{
		"default_server_locale": *cfg.LocalizationSettings.DefaultServerLocale,
		"default_client_locale": *cfg.LocalizationSettings.DefaultClientLocale,
		"available_locales":     *cfg.LocalizationSettings.AvailableLocales,
	})

	s.SendDiagnostic(TRACK_CONFIG_SAML, map[string]interface{}{
		"enable":                              *cfg.SamlSettings.Enable,
		"enable_sync_with_ldap":               *cfg.SamlSettings.EnableSyncWithLdap,
		"enable_sync_with_ldap_include_auth":  *cfg.SamlSettings.EnableSyncWithLdapIncludeAuth,
		"enable_admin_attribute":              *cfg.SamlSettings.EnableAdminAttribute,
		"verify":                              *cfg.SamlSettings.Verify,
		"encrypt":                             *cfg.SamlSettings.Encrypt,
		"sign_request":                        *cfg.SamlSettings.SignRequest,
		"isdefault_signature_algorithm":       isDefault(*cfg.SamlSettings.SignatureAlgorithm, ""),
		"isdefault_canonical_algorithm":       isDefault(*cfg.SamlSettings.CanonicalAlgorithm, ""),
		"isdefault_scoping_idp_provider_id":   isDefault(*cfg.SamlSettings.ScopingIDPProviderId, ""),
		"isdefault_scoping_idp_name":          isDefault(*cfg.SamlSettings.ScopingIDPName, ""),
		"isdefault_id_attribute":              isDefault(*cfg.SamlSettings.IdAttribute, model.SAML_SETTINGS_DEFAULT_ID_ATTRIBUTE),
		"isdefault_guest_attribute":           isDefault(*cfg.SamlSettings.GuestAttribute, model.SAML_SETTINGS_DEFAULT_GUEST_ATTRIBUTE),
		"isdefault_admin_attribute":           isDefault(*cfg.SamlSettings.AdminAttribute, model.SAML_SETTINGS_DEFAULT_ADMIN_ATTRIBUTE),
		"isdefault_first_name_attribute":      isDefault(*cfg.SamlSettings.FirstNameAttribute, model.SAML_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE),
		"isdefault_last_name_attribute":       isDefault(*cfg.SamlSettings.LastNameAttribute, model.SAML_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE),
		"isdefault_email_attribute":           isDefault(*cfg.SamlSettings.EmailAttribute, model.SAML_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE),
		"isdefault_username_attribute":        isDefault(*cfg.SamlSettings.UsernameAttribute, model.SAML_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE),
		"isdefault_nickname_attribute":        isDefault(*cfg.SamlSettings.NicknameAttribute, model.SAML_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE),
		"isdefault_locale_attribute":          isDefault(*cfg.SamlSettings.LocaleAttribute, model.SAML_SETTINGS_DEFAULT_LOCALE_ATTRIBUTE),
		"isdefault_position_attribute":        isDefault(*cfg.SamlSettings.PositionAttribute, model.SAML_SETTINGS_DEFAULT_POSITION_ATTRIBUTE),
		"isdefault_login_button_text":         isDefault(*cfg.SamlSettings.LoginButtonText, model.USER_AUTH_SERVICE_SAML_TEXT),
		"isdefault_login_button_color":        isDefault(*cfg.SamlSettings.LoginButtonColor, ""),
		"isdefault_login_button_border_color": isDefault(*cfg.SamlSettings.LoginButtonBorderColor, ""),
		"isdefault_login_button_text_color":   isDefault(*cfg.SamlSettings.LoginButtonTextColor, ""),
	})

	s.SendDiagnostic(TRACK_CONFIG_CLUSTER, map[string]interface{}{
		"enable":                                *cfg.ClusterSettings.Enable,
		"network_interface":                     isDefault(*cfg.ClusterSettings.NetworkInterface, ""),
		"bind_address":                          isDefault(*cfg.ClusterSettings.BindAddress, ""),
		"advertise_address":                     isDefault(*cfg.ClusterSettings.AdvertiseAddress, ""),
		"use_ip_address":                        *cfg.ClusterSettings.UseIpAddress,
		"use_experimental_gossip":               *cfg.ClusterSettings.UseExperimentalGossip,
		"enable_experimental_gossip_encryption": *cfg.ClusterSettings.EnableExperimentalGossipEncryption,
		"read_only_config":                      *cfg.ClusterSettings.ReadOnlyConfig,
	})

	s.SendDiagnostic(TRACK_CONFIG_METRICS, map[string]interface{}{
		"enable":             *cfg.MetricsSettings.Enable,
		"block_profile_rate": *cfg.MetricsSettings.BlockProfileRate,
	})

	s.SendDiagnostic(TRACK_CONFIG_NATIVEAPP, map[string]interface{}{
		"isdefault_app_download_link":         isDefault(*cfg.NativeAppSettings.AppDownloadLink, model.NATIVEAPP_SETTINGS_DEFAULT_APP_DOWNLOAD_LINK),
		"isdefault_android_app_download_link": isDefault(*cfg.NativeAppSettings.AndroidAppDownloadLink, model.NATIVEAPP_SETTINGS_DEFAULT_ANDROID_APP_DOWNLOAD_LINK),
		"isdefault_iosapp_download_link":      isDefault(*cfg.NativeAppSettings.IosAppDownloadLink, model.NATIVEAPP_SETTINGS_DEFAULT_IOS_APP_DOWNLOAD_LINK),
	})

	s.SendDiagnostic(TRACK_CONFIG_EXPERIMENTAL, map[string]interface{}{
		"client_side_cert_enable":            *cfg.ExperimentalSettings.ClientSideCertEnable,
		"isdefault_client_side_cert_check":   isDefault(*cfg.ExperimentalSettings.ClientSideCertCheck, model.CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH),
		"link_metadata_timeout_milliseconds": *cfg.ExperimentalSettings.LinkMetadataTimeoutMilliseconds,
		"enable_click_to_reply":              *cfg.ExperimentalSettings.EnableClickToReply,
		"restrict_system_admin":              *cfg.ExperimentalSettings.RestrictSystemAdmin,
		"use_new_saml_library":               *cfg.ExperimentalSettings.UseNewSAMLLibrary,
		"cloud_billing":                      *cfg.ExperimentalSettings.CloudBilling,
	})

	s.SendDiagnostic(TRACK_CONFIG_ANALYTICS, map[string]interface{}{
		"isdefault_max_users_for_statistics": isDefault(*cfg.AnalyticsSettings.MaxUsersForStatistics, model.ANALYTICS_SETTINGS_DEFAULT_MAX_USERS_FOR_STATISTICS),
	})

	s.SendDiagnostic(TRACK_CONFIG_ANNOUNCEMENT, map[string]interface{}{
		"enable_banner":               *cfg.AnnouncementSettings.EnableBanner,
		"isdefault_banner_color":      isDefault(*cfg.AnnouncementSettings.BannerColor, model.ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_COLOR),
		"isdefault_banner_text_color": isDefault(*cfg.AnnouncementSettings.BannerTextColor, model.ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_TEXT_COLOR),
		"allow_banner_dismissal":      *cfg.AnnouncementSettings.AllowBannerDismissal,
	})

	s.SendDiagnostic(TRACK_CONFIG_ELASTICSEARCH, map[string]interface{}{
		"isdefault_connection_url":          isDefault(*cfg.ElasticsearchSettings.ConnectionUrl, model.ELASTICSEARCH_SETTINGS_DEFAULT_CONNECTION_URL),
		"isdefault_username":                isDefault(*cfg.ElasticsearchSettings.Username, model.ELASTICSEARCH_SETTINGS_DEFAULT_USERNAME),
		"isdefault_password":                isDefault(*cfg.ElasticsearchSettings.Password, model.ELASTICSEARCH_SETTINGS_DEFAULT_PASSWORD),
		"enable_indexing":                   *cfg.ElasticsearchSettings.EnableIndexing,
		"enable_searching":                  *cfg.ElasticsearchSettings.EnableSearching,
		"enable_autocomplete":               *cfg.ElasticsearchSettings.EnableAutocomplete,
		"sniff":                             *cfg.ElasticsearchSettings.Sniff,
		"post_index_replicas":               *cfg.ElasticsearchSettings.PostIndexReplicas,
		"post_index_shards":                 *cfg.ElasticsearchSettings.PostIndexShards,
		"channel_index_replicas":            *cfg.ElasticsearchSettings.ChannelIndexReplicas,
		"channel_index_shards":              *cfg.ElasticsearchSettings.ChannelIndexShards,
		"user_index_replicas":               *cfg.ElasticsearchSettings.UserIndexReplicas,
		"user_index_shards":                 *cfg.ElasticsearchSettings.UserIndexShards,
		"isdefault_index_prefix":            isDefault(*cfg.ElasticsearchSettings.IndexPrefix, model.ELASTICSEARCH_SETTINGS_DEFAULT_INDEX_PREFIX),
		"live_indexing_batch_size":          *cfg.ElasticsearchSettings.LiveIndexingBatchSize,
		"bulk_indexing_time_window_seconds": *cfg.ElasticsearchSettings.BulkIndexingTimeWindowSeconds,
		"request_timeout_seconds":           *cfg.ElasticsearchSettings.RequestTimeoutSeconds,
		"skip_tls_verification":             *cfg.ElasticsearchSettings.SkipTLSVerification,
		"trace":                             *cfg.ElasticsearchSettings.Trace,
	})

	s.trackPluginConfig(cfg, model.PLUGIN_SETTINGS_DEFAULT_MARKETPLACE_URL)

	s.SendDiagnostic(TRACK_CONFIG_DATA_RETENTION, map[string]interface{}{
		"enable_message_deletion": *cfg.DataRetentionSettings.EnableMessageDeletion,
		"enable_file_deletion":    *cfg.DataRetentionSettings.EnableFileDeletion,
		"message_retention_days":  *cfg.DataRetentionSettings.MessageRetentionDays,
		"file_retention_days":     *cfg.DataRetentionSettings.FileRetentionDays,
		"deletion_job_start_time": *cfg.DataRetentionSettings.DeletionJobStartTime,
	})

	s.SendDiagnostic(TRACK_CONFIG_MESSAGE_EXPORT, map[string]interface{}{
		"enable_message_export":                 *cfg.MessageExportSettings.EnableExport,
		"export_format":                         *cfg.MessageExportSettings.ExportFormat,
		"daily_run_time":                        *cfg.MessageExportSettings.DailyRunTime,
		"default_export_from_timestamp":         *cfg.MessageExportSettings.ExportFromTimestamp,
		"batch_size":                            *cfg.MessageExportSettings.BatchSize,
		"global_relay_customer_type":            *cfg.MessageExportSettings.GlobalRelaySettings.CustomerType,
		"is_default_global_relay_smtp_username": isDefault(*cfg.MessageExportSettings.GlobalRelaySettings.SmtpUsername, ""),
		"is_default_global_relay_smtp_password": isDefault(*cfg.MessageExportSettings.GlobalRelaySettings.SmtpPassword, ""),
		"is_default_global_relay_email_address": isDefault(*cfg.MessageExportSettings.GlobalRelaySettings.EmailAddress, ""),
		"global_relay_smtp_server_timeout":      *cfg.EmailSettings.SMTPServerTimeout,
	})

	s.SendDiagnostic(TRACK_CONFIG_DISPLAY, map[string]interface{}{
		"experimental_timezone":        *cfg.DisplaySettings.ExperimentalTimezone,
		"isdefault_custom_url_schemes": len(cfg.DisplaySettings.CustomUrlSchemes) != 0,
	})

	s.SendDiagnostic(TRACK_CONFIG_GUEST_ACCOUNTS, map[string]interface{}{
		"enable":                                 *cfg.GuestAccountsSettings.Enable,
		"allow_email_accounts":                   *cfg.GuestAccountsSettings.AllowEmailAccounts,
		"enforce_multifactor_authentication":     *cfg.GuestAccountsSettings.EnforceMultifactorAuthentication,
		"isdefault_restrict_creation_to_domains": isDefault(*cfg.GuestAccountsSettings.RestrictCreationToDomains, ""),
	})

	s.SendDiagnostic(TRACK_CONFIG_IMAGE_PROXY, map[string]interface{}{
		"enable":                               *cfg.ImageProxySettings.Enable,
		"image_proxy_type":                     *cfg.ImageProxySettings.ImageProxyType,
		"isdefault_remote_image_proxy_url":     isDefault(*cfg.ImageProxySettings.RemoteImageProxyURL, ""),
		"isdefault_remote_image_proxy_options": isDefault(*cfg.ImageProxySettings.RemoteImageProxyOptions, ""),
	})

	s.SendDiagnostic(TRACK_CONFIG_BLEVE, map[string]interface{}{
		"enable_indexing":                   *cfg.BleveSettings.EnableIndexing,
		"enable_searching":                  *cfg.BleveSettings.EnableSearching,
		"enable_autocomplete":               *cfg.BleveSettings.EnableAutocomplete,
		"bulk_indexing_time_window_seconds": *cfg.BleveSettings.BulkIndexingTimeWindowSeconds,
	})
}

func (s *Server) trackLicense() {
	if license := s.License(); license != nil {
		data := map[string]interface{}{
			"customer_id": license.Customer.Id,
			"license_id":  license.Id,
			"issued":      license.IssuedAt,
			"start":       license.StartsAt,
			"expire":      license.ExpiresAt,
			"users":       *license.Features.Users,
			"edition":     license.SkuShortName,
		}

		features := license.Features.ToMap()
		for featureName, featureValue := range features {
			data["feature_"+featureName] = featureValue
		}

		s.SendDiagnostic(TRACK_LICENSE, data)
	}
}

func (s *Server) trackPlugins() {
	pluginsEnvironment := s.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return
	}

	totalEnabledCount := 0
	webappEnabledCount := 0
	backendEnabledCount := 0
	totalDisabledCount := 0
	webappDisabledCount := 0
	backendDisabledCount := 0
	brokenManifestCount := 0
	settingsCount := 0

	pluginStates := s.Config().PluginSettings.PluginStates
	plugins, _ := pluginsEnvironment.Available()

	if pluginStates != nil && plugins != nil {
		for _, plugin := range plugins {
			if plugin.Manifest == nil {
				brokenManifestCount += 1
				continue
			}

			if state, ok := pluginStates[plugin.Manifest.Id]; ok && state.Enable {
				totalEnabledCount += 1
				if plugin.Manifest.HasServer() {
					backendEnabledCount += 1
				}
				if plugin.Manifest.HasWebapp() {
					webappEnabledCount += 1
				}
			} else {
				totalDisabledCount += 1
				if plugin.Manifest.HasServer() {
					backendDisabledCount += 1
				}
				if plugin.Manifest.HasWebapp() {
					webappDisabledCount += 1
				}
			}
			if plugin.Manifest.SettingsSchema != nil {
				settingsCount += 1
			}
		}
	} else {
		totalEnabledCount = -1  // -1 to indicate disabled or error
		totalDisabledCount = -1 // -1 to indicate disabled or error
	}

	s.SendDiagnostic(TRACK_PLUGINS, map[string]interface{}{
		"enabled_plugins":               totalEnabledCount,
		"enabled_webapp_plugins":        webappEnabledCount,
		"enabled_backend_plugins":       backendEnabledCount,
		"disabled_plugins":              totalDisabledCount,
		"disabled_webapp_plugins":       webappDisabledCount,
		"disabled_backend_plugins":      backendDisabledCount,
		"plugins_with_settings":         settingsCount,
		"plugins_with_broken_manifests": brokenManifestCount,
	})
}

func (s *Server) trackServer() {
	data := map[string]interface{}{
		"edition":          model.BuildEnterpriseReady,
		"version":          model.CurrentVersion,
		"database_type":    *s.Config().SqlSettings.DriverName,
		"operating_system": runtime.GOOS,
	}

	if scr, err := s.Store.User().AnalyticsGetSystemAdminCount(); err == nil {
		data["system_admins"] = scr
	}

	if scr, err := s.Store.GetDbVersion(); err == nil {
		data["database_version"] = scr
	}

	s.SendDiagnostic(TRACK_SERVER, data)
}

func (s *Server) trackPermissions() {
	phase1Complete := false
	if _, err := s.Store.System().GetByName(ADVANCED_PERMISSIONS_MIGRATION_KEY); err == nil {
		phase1Complete = true
	}

	phase2Complete := false
	if _, err := s.Store.System().GetByName(model.MIGRATION_KEY_ADVANCED_PERMISSIONS_PHASE_2); err == nil {
		phase2Complete = true
	}

	s.SendDiagnostic(TRACK_PERMISSIONS_GENERAL, map[string]interface{}{
		"phase_1_migration_complete": phase1Complete,
		"phase_2_migration_complete": phase2Complete,
	})

	systemAdminPermissions := ""
	if role, err := s.GetRoleByName(model.SYSTEM_ADMIN_ROLE_ID); err == nil {
		systemAdminPermissions = strings.Join(role.Permissions, " ")
	}

	systemUserPermissions := ""
	if role, err := s.GetRoleByName(model.SYSTEM_USER_ROLE_ID); err == nil {
		systemUserPermissions = strings.Join(role.Permissions, " ")
	}

	teamAdminPermissions := ""
	if role, err := s.GetRoleByName(model.TEAM_ADMIN_ROLE_ID); err == nil {
		teamAdminPermissions = strings.Join(role.Permissions, " ")
	}

	teamUserPermissions := ""
	if role, err := s.GetRoleByName(model.TEAM_USER_ROLE_ID); err == nil {
		teamUserPermissions = strings.Join(role.Permissions, " ")
	}

	teamGuestPermissions := ""
	if role, err := s.GetRoleByName(model.TEAM_GUEST_ROLE_ID); err == nil {
		teamGuestPermissions = strings.Join(role.Permissions, " ")
	}

	channelAdminPermissions := ""
	if role, err := s.GetRoleByName(model.CHANNEL_ADMIN_ROLE_ID); err == nil {
		channelAdminPermissions = strings.Join(role.Permissions, " ")
	}

	channelUserPermissions := ""
	if role, err := s.GetRoleByName(model.CHANNEL_USER_ROLE_ID); err == nil {
		channelUserPermissions = strings.Join(role.Permissions, " ")
	}

	channelGuestPermissions := ""
	if role, err := s.GetRoleByName(model.CHANNEL_GUEST_ROLE_ID); err == nil {
		channelGuestPermissions = strings.Join(role.Permissions, " ")
	}

	s.SendDiagnostic(TRACK_PERMISSIONS_SYSTEM_SCHEME, map[string]interface{}{
		"system_admin_permissions":  systemAdminPermissions,
		"system_user_permissions":   systemUserPermissions,
		"team_admin_permissions":    teamAdminPermissions,
		"team_user_permissions":     teamUserPermissions,
		"team_guest_permissions":    teamGuestPermissions,
		"channel_admin_permissions": channelAdminPermissions,
		"channel_user_permissions":  channelUserPermissions,
		"channel_guest_permissions": channelGuestPermissions,
	})

	if schemes, err := s.GetSchemes(model.SCHEME_SCOPE_TEAM, 0, 100); err == nil {
		for _, scheme := range schemes {
			teamAdminPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultTeamAdminRole); err == nil {
				teamAdminPermissions = strings.Join(role.Permissions, " ")
			}

			teamUserPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultTeamUserRole); err == nil {
				teamUserPermissions = strings.Join(role.Permissions, " ")
			}

			teamGuestPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultTeamGuestRole); err == nil {
				teamGuestPermissions = strings.Join(role.Permissions, " ")
			}

			channelAdminPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultChannelAdminRole); err == nil {
				channelAdminPermissions = strings.Join(role.Permissions, " ")
			}

			channelUserPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultChannelUserRole); err == nil {
				channelUserPermissions = strings.Join(role.Permissions, " ")
			}

			channelGuestPermissions := ""
			if role, err := s.GetRoleByName(scheme.DefaultChannelGuestRole); err == nil {
				channelGuestPermissions = strings.Join(role.Permissions, " ")
			}

			count, _ := s.Store.Team().AnalyticsGetTeamCountForScheme(scheme.Id)

			s.SendDiagnostic(TRACK_PERMISSIONS_TEAM_SCHEMES, map[string]interface{}{
				"scheme_id":                 scheme.Id,
				"team_admin_permissions":    teamAdminPermissions,
				"team_user_permissions":     teamUserPermissions,
				"team_guest_permissions":    teamGuestPermissions,
				"channel_admin_permissions": channelAdminPermissions,
				"channel_user_permissions":  channelUserPermissions,
				"channel_guest_permissions": channelGuestPermissions,
				"team_count":                count,
			})
		}
	}
}

func (s *Server) trackElasticsearch() {
	data := map[string]interface{}{}

	for _, engine := range s.SearchEngine.GetActiveEngines() {
		if engine.GetVersion() != 0 && engine.GetName() == "elasticsearch" {
			data["elasticsearch_server_version"] = engine.GetVersion()
		}
	}

	s.SendDiagnostic(TRACK_ELASTICSEARCH, data)
}

func (s *Server) trackGroups() {
	groupCount, err := s.Store.Group().GroupCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupTeamCount, err := s.Store.Group().GroupTeamCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupChannelCount, err := s.Store.Group().GroupChannelCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupSyncedTeamCount, err := s.Store.Team().GroupSyncedTeamCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupSyncedChannelCount, err := s.Store.Channel().GroupSyncedChannelCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupMemberCount, err := s.Store.Group().GroupMemberCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	distinctGroupMemberCount, err := s.Store.Group().DistinctGroupMemberCount()
	if err != nil {
		mlog.Error(err.Error())
	}

	groupCountWithAllowReference, err := s.Store.Group().GroupCountWithAllowReference()
	if err != nil {
		mlog.Error(err.Error())
	}

	s.SendDiagnostic(TRACK_GROUPS, map[string]interface{}{
		"group_count":                      groupCount,
		"group_team_count":                 groupTeamCount,
		"group_channel_count":              groupChannelCount,
		"group_synced_team_count":          groupSyncedTeamCount,
		"group_synced_channel_count":       groupSyncedChannelCount,
		"group_member_count":               groupMemberCount,
		"distinct_group_member_count":      distinctGroupMemberCount,
		"group_count_with_allow_reference": groupCountWithAllowReference,
	})
}

func (s *Server) trackChannelModeration() {
	channelSchemeCount, err := s.Store.Scheme().CountByScope(model.SCHEME_SCOPE_CHANNEL)
	if err != nil {
		mlog.Error(err.Error())
	}

	createPostUser, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_CREATE_POST.Id, model.RoleScopeChannel, model.RoleTypeUser)
	if err != nil {
		mlog.Error(err.Error())
	}

	createPostGuest, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_CREATE_POST.Id, model.RoleScopeChannel, model.RoleTypeGuest)
	if err != nil {
		mlog.Error(err.Error())
	}

	// only need to track one of 'add_reaction' or 'remove_reaction` because they're both toggled together by the channel moderation feature
	postReactionsUser, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_ADD_REACTION.Id, model.RoleScopeChannel, model.RoleTypeUser)
	if err != nil {
		mlog.Error(err.Error())
	}

	postReactionsGuest, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_ADD_REACTION.Id, model.RoleScopeChannel, model.RoleTypeGuest)
	if err != nil {
		mlog.Error(err.Error())
	}

	// only need to track one of 'manage_public_channel_members' or 'manage_private_channel_members` because they're both toggled together by the channel moderation feature
	manageMembersUser, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id, model.RoleScopeChannel, model.RoleTypeUser)
	if err != nil {
		mlog.Error(err.Error())
	}

	useChannelMentionsUser, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_USE_CHANNEL_MENTIONS.Id, model.RoleScopeChannel, model.RoleTypeUser)
	if err != nil {
		mlog.Error(err.Error())
	}

	useChannelMentionsGuest, err := s.Store.Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, model.PERMISSION_USE_CHANNEL_MENTIONS.Id, model.RoleScopeChannel, model.RoleTypeGuest)
	if err != nil {
		mlog.Error(err.Error())
	}

	s.SendDiagnostic(TRACK_CHANNEL_MODERATION, map[string]interface{}{
		"channel_scheme_count": channelSchemeCount,

		"create_post_user_disabled_count":  createPostUser,
		"create_post_guest_disabled_count": createPostGuest,

		"post_reactions_user_disabled_count":  postReactionsUser,
		"post_reactions_guest_disabled_count": postReactionsGuest,

		"manage_members_user_disabled_count": manageMembersUser, // the UI does not allow this to be removed for guests

		"use_channel_mentions_user_disabled_count":  useChannelMentionsUser,
		"use_channel_mentions_guest_disabled_count": useChannelMentionsGuest,
	})
}

func (s *Server) trackWarnMetrics() {
	systemDataList, nErr := s.Store.System().Get()
	if nErr != nil {
		return
	}
	for key, value := range systemDataList {
		if strings.HasPrefix(key, model.WARN_METRIC_STATUS_STORE_PREFIX) {
			if _, ok := model.WarnMetricsTable[key]; ok {
				s.SendDiagnostic(TRACK_WARN_METRICS, map[string]interface{}{
					key: value != "false",
				})
			}
		}
	}
}

func (s *Server) trackPluginConfig(cfg *model.Config, marketplaceURL string) {
	pluginConfigData := map[string]interface{}{
		"enable_nps_survey":             pluginSetting(&cfg.PluginSettings, "com.mattermost.nps", "enablesurvey", true),
		"enable":                        *cfg.PluginSettings.Enable,
		"enable_uploads":                *cfg.PluginSettings.EnableUploads,
		"allow_insecure_download_url":   *cfg.PluginSettings.AllowInsecureDownloadUrl,
		"enable_health_check":           *cfg.PluginSettings.EnableHealthCheck,
		"enable_marketplace":            *cfg.PluginSettings.EnableMarketplace,
		"require_pluginSignature":       *cfg.PluginSettings.RequirePluginSignature,
		"enable_remote_marketplace":     *cfg.PluginSettings.EnableRemoteMarketplace,
		"automatic_prepackaged_plugins": *cfg.PluginSettings.AutomaticPrepackagedPlugins,
		"is_default_marketplace_url":    isDefault(*cfg.PluginSettings.MarketplaceUrl, model.PLUGIN_SETTINGS_DEFAULT_MARKETPLACE_URL),
		"signature_public_key_files":    len(cfg.PluginSettings.SignaturePublicKeyFiles),
	}

	// knownPluginIDs lists all known plugin IDs in the Marketplace
	knownPluginIDs := []string{
		"antivirus",
		"com.github.manland.mattermost-plugin-gitlab",
		"com.github.moussetc.mattermost.plugin.giphy",
		"com.github.phillipahereza.mattermost-plugin-digitalocean",
		"com.mattermost.aws-sns",
		"com.mattermost.confluence",
		"com.mattermost.custom-attributes",
		"com.mattermost.mscalendar",
		"com.mattermost.nps",
		"com.mattermost.plugin-incident-response",
		"com.mattermost.plugin-todo",
		"com.mattermost.webex",
		"com.mattermost.welcomebot",
		"github",
		"jenkins",
		"jira",
		"jitsi",
		"mattermost-autolink",
		"memes",
		"skype4business",
		"zoom",
	}

	marketplacePlugins, err := s.getAllMarketplaceplugins(marketplaceURL)
	if err != nil {
		mlog.Info("Failed to fetch marketplace plugins for telemetry. Using predefined list.", mlog.Err(err))

		for _, id := range knownPluginIDs {
			pluginConfigData["enable_"+id] = pluginActivated(cfg.PluginSettings.PluginStates, id)
		}
	} else {
		for _, p := range marketplacePlugins {
			id := p.Manifest.Id

			pluginConfigData["enable_"+id] = pluginActivated(cfg.PluginSettings.PluginStates, id)
		}
	}

	pluginsEnvironment := s.GetPluginsEnvironment()
	if pluginsEnvironment != nil {
		if plugins, appErr := pluginsEnvironment.Available(); appErr != nil {
			mlog.Error("Unable to add plugin versions to diagnostics", mlog.Err(appErr))
		} else {
			// If marketplace request failed, use predefined list
			if marketplacePlugins == nil {
				for _, id := range knownPluginIDs {
					pluginConfigData["version_"+id] = pluginActivated(cfg.PluginSettings.PluginStates, id)
				}
			} else {
				for _, p := range marketplacePlugins {
					id := p.Manifest.Id

					pluginConfigData["version_"+id] = pluginVersion(plugins, id)
				}
			}
		}
	}

	s.SendDiagnostic(TRACK_CONFIG_PLUGIN, pluginConfigData)
}

func (s *Server) getAllMarketplaceplugins(marketplaceURL string) ([]*model.BaseMarketplacePlugin, error) {
	marketplaceClient, err := marketplace.NewClient(
		marketplaceURL,
		s.HTTPService,
	)
	if err != nil {
		return nil, err
	}

	// Fetch all plugins from marketplace.
	filter := &model.MarketplacePluginFilter{
		PerPage:       -1,
		ServerVersion: model.CurrentVersion,
	}

	license := s.License()
	if license != nil && *license.Features.EnterprisePlugins {
		filter.EnterprisePlugins = true
	}

	if model.BuildEnterpriseReady == "true" {
		filter.BuildEnterpriseReady = true
	}

	return marketplaceClient.GetPlugins(filter)
}

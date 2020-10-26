// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"context"
	dbsql "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/utils"
)

const (
	INDEX_TYPE_FULL_TEXT = "full_text"
	INDEX_TYPE_DEFAULT   = "default"
	DB_PING_ATTEMPTS     = 18
	DB_PING_TIMEOUT_SECS = 10
)

const (
	EXIT_GENERIC_FAILURE             = 1
	EXIT_CREATE_TABLE                = 100
	EXIT_DB_OPEN                     = 101
	EXIT_PING                        = 102
	EXIT_NO_DRIVER                   = 103
	EXIT_TABLE_EXISTS                = 104
	EXIT_TABLE_EXISTS_MYSQL          = 105
	EXIT_COLUMN_EXISTS               = 106
	EXIT_DOES_COLUMN_EXISTS_POSTGRES = 107
	EXIT_DOES_COLUMN_EXISTS_MYSQL    = 108
	EXIT_DOES_COLUMN_EXISTS_MISSING  = 109
	EXIT_CREATE_COLUMN_POSTGRES      = 110
	EXIT_CREATE_COLUMN_MYSQL         = 111
	EXIT_CREATE_COLUMN_MISSING       = 112
	EXIT_REMOVE_COLUMN               = 113
	EXIT_RENAME_COLUMN               = 114
	EXIT_MAX_COLUMN                  = 115
	EXIT_ALTER_COLUMN                = 116
	EXIT_CREATE_INDEX_POSTGRES       = 117
	EXIT_CREATE_INDEX_MYSQL          = 118
	EXIT_CREATE_INDEX_FULL_MYSQL     = 119
	EXIT_CREATE_INDEX_MISSING        = 120
	EXIT_REMOVE_INDEX_POSTGRES       = 121
	EXIT_REMOVE_INDEX_MYSQL          = 122
	EXIT_REMOVE_INDEX_MISSING        = 123
	EXIT_REMOVE_TABLE                = 134
	EXIT_CREATE_INDEX_SQLITE         = 135
	EXIT_REMOVE_INDEX_SQLITE         = 136
	EXIT_TABLE_EXISTS_SQLITE         = 137
	EXIT_DOES_COLUMN_EXISTS_SQLITE   = 138
	EXIT_ALTER_PRIMARY_KEY           = 139
)

type SqlSupplierStores struct {
	team                 store.TeamStore
	channel              store.ChannelStore
	post                 store.PostStore
	thread               store.ThreadStore
	user                 store.UserStore
	bot                  store.BotStore
	audit                store.AuditStore
	cluster              store.ClusterDiscoveryStore
	compliance           store.ComplianceStore
	session              store.SessionStore
	oauth                store.OAuthStore
	system               store.SystemStore
	webhook              store.WebhookStore
	command              store.CommandStore
	commandWebhook       store.CommandWebhookStore
	preference           store.PreferenceStore
	license              store.LicenseStore
	token                store.TokenStore
	emoji                store.EmojiStore
	status               store.StatusStore
	fileInfo             store.FileInfoStore
	uploadSession        store.UploadSessionStore
	reaction             store.ReactionStore
	job                  store.JobStore
	userAccessToken      store.UserAccessTokenStore
	plugin               store.PluginStore
	channelMemberHistory store.ChannelMemberHistoryStore
	role                 store.RoleStore
	scheme               store.SchemeStore
	TermsOfService       store.TermsOfServiceStore
	productNotices       store.ProductNoticesStore
	group                store.GroupStore
	UserTermsOfService   store.UserTermsOfServiceStore
	linkMetadata         store.LinkMetadataStore
}

type SqlSupplier struct {
	// rrCounter and srCounter should be kept first.
	// See https://github.com/mattermost/mattermost-server/v5/pull/7281
	rrCounter      int64
	srCounter      int64
	master         *gorp.DbMap
	replicas       []*gorp.DbMap
	searchReplicas []*gorp.DbMap
	stores         SqlSupplierStores
	settings       *model.SqlSettings
	lockedToMaster bool
	context        context.Context
	license        *model.License
	licenseMutex   sync.RWMutex
}

type TraceOnAdapter struct{}

func (t *TraceOnAdapter) Printf(format string, v ...interface{}) {
	originalString := fmt.Sprintf(format, v...)
	newString := strings.ReplaceAll(originalString, "\n", " ")
	newString = strings.ReplaceAll(newString, "\t", " ")
	newString = strings.ReplaceAll(newString, "\"", "")
	mlog.Debug(newString)
}

func NewSqlSupplier(settings model.SqlSettings, metrics einterfaces.MetricsInterface) *SqlSupplier {
	supplier := &SqlSupplier{
		rrCounter: 0,
		srCounter: 0,
		settings:  &settings,
	}

	supplier.initConnection()

	supplier.stores.team = newSqlTeamStore(supplier)
	supplier.stores.channel = newSqlChannelStore(supplier, metrics)
	supplier.stores.post = newSqlPostStore(supplier, metrics)
	supplier.stores.user = newSqlUserStore(supplier, metrics)
	supplier.stores.bot = newSqlBotStore(supplier, metrics)
	supplier.stores.audit = newSqlAuditStore(supplier)
	supplier.stores.cluster = newSqlClusterDiscoveryStore(supplier)
	supplier.stores.compliance = newSqlComplianceStore(supplier)
	supplier.stores.session = newSqlSessionStore(supplier)
	supplier.stores.oauth = newSqlOAuthStore(supplier)
	supplier.stores.system = newSqlSystemStore(supplier)
	supplier.stores.webhook = newSqlWebhookStore(supplier, metrics)
	supplier.stores.command = newSqlCommandStore(supplier)
	supplier.stores.commandWebhook = newSqlCommandWebhookStore(supplier)
	supplier.stores.preference = newSqlPreferenceStore(supplier)
	supplier.stores.license = newSqlLicenseStore(supplier)
	supplier.stores.token = newSqlTokenStore(supplier)
	supplier.stores.emoji = newSqlEmojiStore(supplier, metrics)
	supplier.stores.status = newSqlStatusStore(supplier)
	supplier.stores.fileInfo = newSqlFileInfoStore(supplier, metrics)
	supplier.stores.uploadSession = newSqlUploadSessionStore(supplier)
	supplier.stores.thread = newSqlThreadStore(supplier)
	supplier.stores.job = newSqlJobStore(supplier)
	supplier.stores.userAccessToken = newSqlUserAccessTokenStore(supplier)
	supplier.stores.channelMemberHistory = newSqlChannelMemberHistoryStore(supplier)
	supplier.stores.plugin = newSqlPluginStore(supplier)
	supplier.stores.TermsOfService = newSqlTermsOfServiceStore(supplier, metrics)
	supplier.stores.UserTermsOfService = newSqlUserTermsOfServiceStore(supplier)
	supplier.stores.linkMetadata = newSqlLinkMetadataStore(supplier)
	supplier.stores.reaction = newSqlReactionStore(supplier)
	supplier.stores.role = newSqlRoleStore(supplier)
	supplier.stores.scheme = newSqlSchemeStore(supplier)
	supplier.stores.group = newSqlGroupStore(supplier)
	supplier.stores.productNotices = newSqlProductNoticesStore(supplier)
	err := supplier.GetMaster().CreateTablesIfNotExists()
	if err != nil {
		mlog.Critical("Error creating database tables.", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_TABLE)
	}

	err = upgradeDatabase(supplier, model.CurrentVersion)
	if err != nil {
		mlog.Critical("Failed to upgrade database.", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_GENERIC_FAILURE)
	}

	supplier.stores.team.(*SqlTeamStore).createIndexesIfNotExists()
	supplier.stores.channel.(*SqlChannelStore).createIndexesIfNotExists()
	supplier.stores.post.(*SqlPostStore).createIndexesIfNotExists()
	supplier.stores.thread.(*SqlThreadStore).createIndexesIfNotExists()
	supplier.stores.user.(*SqlUserStore).createIndexesIfNotExists()
	supplier.stores.bot.(*SqlBotStore).createIndexesIfNotExists()
	supplier.stores.audit.(*SqlAuditStore).createIndexesIfNotExists()
	supplier.stores.compliance.(*SqlComplianceStore).createIndexesIfNotExists()
	supplier.stores.session.(*SqlSessionStore).createIndexesIfNotExists()
	supplier.stores.oauth.(*SqlOAuthStore).createIndexesIfNotExists()
	supplier.stores.system.(*SqlSystemStore).createIndexesIfNotExists()
	supplier.stores.webhook.(*SqlWebhookStore).createIndexesIfNotExists()
	supplier.stores.command.(*SqlCommandStore).createIndexesIfNotExists()
	supplier.stores.commandWebhook.(*SqlCommandWebhookStore).createIndexesIfNotExists()
	supplier.stores.preference.(*SqlPreferenceStore).createIndexesIfNotExists()
	supplier.stores.license.(*SqlLicenseStore).createIndexesIfNotExists()
	supplier.stores.token.(*SqlTokenStore).createIndexesIfNotExists()
	supplier.stores.emoji.(*SqlEmojiStore).createIndexesIfNotExists()
	supplier.stores.status.(*SqlStatusStore).createIndexesIfNotExists()
	supplier.stores.fileInfo.(*SqlFileInfoStore).createIndexesIfNotExists()
	supplier.stores.uploadSession.(*SqlUploadSessionStore).createIndexesIfNotExists()
	supplier.stores.job.(*SqlJobStore).createIndexesIfNotExists()
	supplier.stores.userAccessToken.(*SqlUserAccessTokenStore).createIndexesIfNotExists()
	supplier.stores.plugin.(*SqlPluginStore).createIndexesIfNotExists()
	supplier.stores.TermsOfService.(SqlTermsOfServiceStore).createIndexesIfNotExists()
	supplier.stores.productNotices.(SqlProductNoticesStore).createIndexesIfNotExists()
	supplier.stores.UserTermsOfService.(SqlUserTermsOfServiceStore).createIndexesIfNotExists()
	supplier.stores.linkMetadata.(*SqlLinkMetadataStore).createIndexesIfNotExists()
	supplier.stores.group.(*SqlGroupStore).createIndexesIfNotExists()
	supplier.stores.scheme.(*SqlSchemeStore).createIndexesIfNotExists()
	supplier.stores.preference.(*SqlPreferenceStore).deleteUnusedFeatures()

	return supplier
}

func setupConnection(con_type string, dataSource string, settings *model.SqlSettings) *gorp.DbMap {
	db, err := dbsql.Open(*settings.DriverName, dataSource)
	if err != nil {
		mlog.Critical("Failed to open SQL connection to err.", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_DB_OPEN)
	}

	for i := 0; i < DB_PING_ATTEMPTS; i++ {
		mlog.Info("Pinging SQL", mlog.String("database", con_type))
		ctx, cancel := context.WithTimeout(context.Background(), DB_PING_TIMEOUT_SECS*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			break
		} else {
			if i == DB_PING_ATTEMPTS-1 {
				mlog.Critical("Failed to ping DB, server will exit.", mlog.Err(err))
				time.Sleep(time.Second)
				os.Exit(EXIT_PING)
			} else {
				mlog.Error("Failed to ping DB", mlog.Err(err), mlog.Int("retrying in seconds", DB_PING_TIMEOUT_SECS))
				time.Sleep(DB_PING_TIMEOUT_SECS * time.Second)
			}
		}
	}

	db.SetMaxIdleConns(*settings.MaxIdleConns)
	db.SetMaxOpenConns(*settings.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(*settings.ConnMaxLifetimeMilliseconds) * time.Millisecond)

	var dbmap *gorp.DbMap

	connectionTimeout := time.Duration(*settings.QueryTimeout) * time.Second

	if *settings.DriverName == model.DATABASE_DRIVER_SQLITE {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.SqliteDialect{}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_MYSQL {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}, QueryTimeout: connectionTimeout}
	} else if *settings.DriverName == model.DATABASE_DRIVER_POSTGRES {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: mattermConverter{}, Dialect: gorp.PostgresDialect{}, QueryTimeout: connectionTimeout}
	} else {
		mlog.Critical("Failed to create dialect specific driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_NO_DRIVER)
	}

	if settings.Trace != nil && *settings.Trace {
		dbmap.TraceOn("sql-trace:", &TraceOnAdapter{})
	}

	return dbmap
}

func (ss *SqlSupplier) SetContext(context context.Context) {
	ss.context = context
}

func (ss *SqlSupplier) Context() context.Context {
	return ss.context
}

func (ss *SqlSupplier) initConnection() {
	ss.master = setupConnection("master", *ss.settings.DataSource, ss.settings)

	if len(ss.settings.DataSourceReplicas) > 0 {
		ss.replicas = make([]*gorp.DbMap, len(ss.settings.DataSourceReplicas))
		for i, replica := range ss.settings.DataSourceReplicas {
			ss.replicas[i] = setupConnection(fmt.Sprintf("replica-%v", i), replica, ss.settings)
		}
	}

	if len(ss.settings.DataSourceSearchReplicas) > 0 {
		ss.searchReplicas = make([]*gorp.DbMap, len(ss.settings.DataSourceSearchReplicas))
		for i, replica := range ss.settings.DataSourceSearchReplicas {
			ss.searchReplicas[i] = setupConnection(fmt.Sprintf("search-replica-%v", i), replica, ss.settings)
		}
	}
}

func (ss *SqlSupplier) DriverName() string {
	return *ss.settings.DriverName
}

func (ss *SqlSupplier) GetCurrentSchemaVersion() string {
	version, _ := ss.GetMaster().SelectStr("SELECT Value FROM Systems WHERE Name='Version'")
	return version
}

func (ss *SqlSupplier) GetDbVersion() (string, error) {
	var sqlVersion string
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		sqlVersion = `SHOW server_version`
	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		sqlVersion = `SELECT version()`
	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		sqlVersion = `SELECT sqlite_version()`
	} else {
		return "", errors.New("Not supported driver")
	}

	version, err := ss.GetReplica().SelectStr(sqlVersion)
	if err != nil {
		return "", err
	}

	return version, nil

}

func (ss *SqlSupplier) GetMaster() *gorp.DbMap {
	return ss.master
}

func (ss *SqlSupplier) GetSearchReplica() *gorp.DbMap {
	ss.licenseMutex.RLock()
	license := ss.license
	ss.licenseMutex.RUnlock()
	if license == nil {
		return ss.GetMaster()
	}

	if len(ss.settings.DataSourceSearchReplicas) == 0 {
		return ss.GetReplica()
	}

	rrNum := atomic.AddInt64(&ss.srCounter, 1) % int64(len(ss.searchReplicas))
	return ss.searchReplicas[rrNum]
}

func (ss *SqlSupplier) GetReplica() *gorp.DbMap {
	ss.licenseMutex.RLock()
	license := ss.license
	ss.licenseMutex.RUnlock()
	if len(ss.settings.DataSourceReplicas) == 0 || ss.lockedToMaster || license == nil {
		return ss.GetMaster()
	}

	rrNum := atomic.AddInt64(&ss.rrCounter, 1) % int64(len(ss.replicas))
	return ss.replicas[rrNum]
}

func (ss *SqlSupplier) TotalMasterDbConnections() int {
	return ss.GetMaster().Db.Stats().OpenConnections
}

func (ss *SqlSupplier) TotalReadDbConnections() int {
	if len(ss.settings.DataSourceReplicas) == 0 {
		return 0
	}

	count := 0
	for _, db := range ss.replicas {
		count = count + db.Db.Stats().OpenConnections
	}

	return count
}

func (ss *SqlSupplier) TotalSearchDbConnections() int {
	if len(ss.settings.DataSourceSearchReplicas) == 0 {
		return 0
	}

	count := 0
	for _, db := range ss.searchReplicas {
		count = count + db.Db.Stats().OpenConnections
	}

	return count
}

func (ss *SqlSupplier) MarkSystemRanUnitTests() {
	props, err := ss.System().Get()
	if err != nil {
		return
	}

	unitTests := props[model.SYSTEM_RAN_UNIT_TESTS]
	if len(unitTests) == 0 {
		systemTests := &model.System{Name: model.SYSTEM_RAN_UNIT_TESTS, Value: "1"}
		ss.System().Save(systemTests)
	}
}

func (ss *SqlSupplier) DoesTableExist(tableName string) bool {
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		count, err := ss.GetMaster().SelectInt(
			`SELECT count(relname) FROM pg_class WHERE relname=$1`,
			strings.ToLower(tableName),
		)

		if err != nil {
			mlog.Critical("Failed to check if table exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_TABLE_EXISTS)
		}

		return count > 0

	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {

		count, err := ss.GetMaster().SelectInt(
			`SELECT
		    COUNT(0) AS table_exists
			FROM
			    information_schema.TABLES
			WHERE
			    TABLE_SCHEMA = DATABASE()
			        AND TABLE_NAME = ?
		    `,
			tableName,
		)

		if err != nil {
			mlog.Critical("Failed to check if table exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_TABLE_EXISTS_MYSQL)
		}

		return count > 0

	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		count, err := ss.GetMaster().SelectInt(
			`SELECT count(name) FROM sqlite_master WHERE type='table' AND name=?`,
			tableName,
		)

		if err != nil {
			mlog.Critical("Failed to check if table exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_TABLE_EXISTS_SQLITE)
		}

		return count > 0

	} else {
		mlog.Critical("Failed to check if column exists because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_COLUMN_EXISTS)
		return false
	}
}

func (ss *SqlSupplier) DoesColumnExist(tableName string, columnName string) bool {
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		count, err := ss.GetMaster().SelectInt(
			`SELECT COUNT(0)
			FROM   pg_attribute
			WHERE  attrelid = $1::regclass
			AND    attname = $2
			AND    NOT attisdropped`,
			strings.ToLower(tableName),
			strings.ToLower(columnName),
		)

		if err != nil {
			if err.Error() == "pq: relation \""+strings.ToLower(tableName)+"\" does not exist" {
				return false
			}

			mlog.Critical("Failed to check if column exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_DOES_COLUMN_EXISTS_POSTGRES)
		}

		return count > 0

	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {

		count, err := ss.GetMaster().SelectInt(
			`SELECT
		    COUNT(0) AS column_exists
		FROM
		    information_schema.COLUMNS
		WHERE
		    TABLE_SCHEMA = DATABASE()
		        AND TABLE_NAME = ?
		        AND COLUMN_NAME = ?`,
			tableName,
			columnName,
		)

		if err != nil {
			mlog.Critical("Failed to check if column exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_DOES_COLUMN_EXISTS_MYSQL)
		}

		return count > 0

	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		count, err := ss.GetMaster().SelectInt(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?`,
			tableName,
			columnName,
		)

		if err != nil {
			mlog.Critical("Failed to check if column exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_DOES_COLUMN_EXISTS_SQLITE)
		}

		return count > 0

	} else {
		mlog.Critical("Failed to check if column exists because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_DOES_COLUMN_EXISTS_MISSING)
		return false
	}
}

func (ss *SqlSupplier) DoesTriggerExist(triggerName string) bool {
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		count, err := ss.GetMaster().SelectInt(`
			SELECT
				COUNT(0)
			FROM
				pg_trigger
			WHERE
				tgname = $1
		`, triggerName)

		if err != nil {
			mlog.Critical("Failed to check if trigger exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_GENERIC_FAILURE)
		}

		return count > 0

	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		count, err := ss.GetMaster().SelectInt(`
			SELECT
				COUNT(0)
			FROM
				information_schema.triggers
			WHERE
				trigger_schema = DATABASE()
			AND	trigger_name = ?
		`, triggerName)

		if err != nil {
			mlog.Critical("Failed to check if trigger exists", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_GENERIC_FAILURE)
		}

		return count > 0

	} else {
		mlog.Critical("Failed to check if column exists because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_GENERIC_FAILURE)
		return false
	}
}

func (ss *SqlSupplier) CreateColumnIfNotExists(tableName string, columnName string, mySqlColType string, postgresColType string, defaultValue string) bool {

	if ss.DoesColumnExist(tableName, columnName) {
		return false
	}

	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, err := ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ADD " + columnName + " " + postgresColType + " DEFAULT '" + defaultValue + "'")
		if err != nil {
			mlog.Critical("Failed to create column", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_COLUMN_POSTGRES)
		}

		return true

	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		_, err := ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ADD " + columnName + " " + mySqlColType + " DEFAULT '" + defaultValue + "'")
		if err != nil {
			mlog.Critical("Failed to create column", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_COLUMN_MYSQL)
		}

		return true

	} else {
		mlog.Critical("Failed to create column because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_COLUMN_MISSING)
		return false
	}
}

func (ss *SqlSupplier) CreateColumnIfNotExistsNoDefault(tableName string, columnName string, mySqlColType string, postgresColType string) bool {

	if ss.DoesColumnExist(tableName, columnName) {
		return false
	}

	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, err := ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ADD " + columnName + " " + postgresColType)
		if err != nil {
			mlog.Critical("Failed to create column", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_COLUMN_POSTGRES)
		}

		return true

	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		_, err := ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ADD " + columnName + " " + mySqlColType)
		if err != nil {
			mlog.Critical("Failed to create column", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_COLUMN_MYSQL)
		}

		return true

	} else {
		mlog.Critical("Failed to create column because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_COLUMN_MISSING)
		return false
	}
}

func (ss *SqlSupplier) RemoveColumnIfExists(tableName string, columnName string) bool {

	if !ss.DoesColumnExist(tableName, columnName) {
		return false
	}

	_, err := ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " DROP COLUMN " + columnName)
	if err != nil {
		mlog.Critical("Failed to drop column", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_REMOVE_COLUMN)
	}

	return true
}

func (ss *SqlSupplier) RemoveTableIfExists(tableName string) bool {
	if !ss.DoesTableExist(tableName) {
		return false
	}

	_, err := ss.GetMaster().ExecNoTimeout("DROP TABLE " + tableName)
	if err != nil {
		mlog.Critical("Failed to drop table", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_REMOVE_TABLE)
	}

	return true
}

func (ss *SqlSupplier) RenameColumnIfExists(tableName string, oldColumnName string, newColumnName string, colType string) bool {
	if !ss.DoesColumnExist(tableName, oldColumnName) {
		return false
	}

	var err error
	if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " CHANGE " + oldColumnName + " " + newColumnName + " " + colType)
	} else if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " RENAME COLUMN " + oldColumnName + " TO " + newColumnName)
	}

	if err != nil {
		mlog.Critical("Failed to rename column", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_RENAME_COLUMN)
	}

	return true
}

func (ss *SqlSupplier) GetMaxLengthOfColumnIfExists(tableName string, columnName string) string {
	if !ss.DoesColumnExist(tableName, columnName) {
		return ""
	}

	var result string
	var err error
	if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		result, err = ss.GetMaster().SelectStr("SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.columns WHERE table_name = '" + tableName + "' AND COLUMN_NAME = '" + columnName + "'")
	} else if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		result, err = ss.GetMaster().SelectStr("SELECT character_maximum_length FROM information_schema.columns WHERE table_name = '" + strings.ToLower(tableName) + "' AND column_name = '" + strings.ToLower(columnName) + "'")
	}

	if err != nil {
		mlog.Critical("Failed to get max length of column", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_MAX_COLUMN)
	}

	return result
}

func (ss *SqlSupplier) AlterColumnTypeIfExists(tableName string, columnName string, mySqlColType string, postgresColType string) bool {
	if !ss.DoesColumnExist(tableName, columnName) {
		return false
	}

	var err error
	if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " MODIFY " + columnName + " " + mySqlColType)
	} else if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + strings.ToLower(tableName) + " ALTER COLUMN " + strings.ToLower(columnName) + " TYPE " + postgresColType)
	}

	if err != nil {
		mlog.Critical("Failed to alter column type", mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_ALTER_COLUMN)
	}

	return true
}

func (ss *SqlSupplier) AlterColumnDefaultIfExists(tableName string, columnName string, mySqlColDefault *string, postgresColDefault *string) bool {
	if !ss.DoesColumnExist(tableName, columnName) {
		return false
	}

	var defaultValue string
	if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		// Some column types in MySQL cannot have defaults, so don't try to configure anything.
		if mySqlColDefault == nil {
			return true
		}

		defaultValue = *mySqlColDefault
	} else if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		// Postgres doesn't have the same limitation, but preserve the interface.
		if postgresColDefault == nil {
			return true
		}

		tableName = strings.ToLower(tableName)
		columnName = strings.ToLower(columnName)
		defaultValue = *postgresColDefault
	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		// SQLite doesn't support altering column defaults, but we don't use this in
		// production so just ignore.
		return true
	} else {
		mlog.Critical("Failed to alter column default because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_GENERIC_FAILURE)
		return false
	}

	var err error
	if defaultValue == "" {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ALTER COLUMN " + columnName + " DROP DEFAULT")
	} else {
		_, err = ss.GetMaster().ExecNoTimeout("ALTER TABLE " + tableName + " ALTER COLUMN " + columnName + " SET DEFAULT " + defaultValue)
	}

	if err != nil {
		mlog.Critical("Failed to alter column", mlog.String("table", tableName), mlog.String("column", columnName), mlog.String("default value", defaultValue), mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_GENERIC_FAILURE)
		return false
	}

	return true
}

func (ss *SqlSupplier) AlterPrimaryKey(tableName string, columnNames []string) bool {
	var currentPrimaryKey string
	var err error
	// get the current primary key as a comma separated list of columns
	switch ss.DriverName() {
	case model.DATABASE_DRIVER_MYSQL:
		query := `
			SELECT GROUP_CONCAT(column_name ORDER BY seq_in_index) AS PK
		FROM
			information_schema.statistics
		WHERE
			table_schema = DATABASE()
		AND table_name = ?
		AND index_name = 'PRIMARY'
		GROUP BY
			index_name`
		currentPrimaryKey, err = ss.GetMaster().SelectStr(query, tableName)
	case model.DATABASE_DRIVER_POSTGRES:
		query := `
			SELECT string_agg(a.attname, ',') AS pk
		FROM
			pg_constraint AS c
		CROSS JOIN
			(SELECT unnest(conkey) FROM pg_constraint WHERE conrelid='` + strings.ToLower(tableName) + `'::REGCLASS AND contype='p') AS cols(colnum)
		INNER JOIN
			pg_attribute AS a ON a.attrelid = c.conrelid
		AND cols.colnum = a.attnum
		WHERE
			c.contype = 'p'
		AND c.conrelid = '` + strings.ToLower(tableName) + `'::REGCLASS`
		currentPrimaryKey, err = ss.GetMaster().SelectStr(query)
	case model.DATABASE_DRIVER_SQLITE:
		// SQLite doesn't support altering primary key
		return true
	}
	if err != nil {
		mlog.Critical("Failed to get current primary key", mlog.String("table", tableName), mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_ALTER_PRIMARY_KEY)
	}

	primaryKey := strings.Join(columnNames, ",")
	if strings.EqualFold(currentPrimaryKey, primaryKey) {
		return false
	}
	// alter primary key
	var alterQuery string
	if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		alterQuery = "ALTER TABLE " + tableName + " DROP PRIMARY KEY, ADD PRIMARY KEY (" + primaryKey + ")"
	} else if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		alterQuery = "ALTER TABLE " + tableName + " DROP CONSTRAINT " + strings.ToLower(tableName) + "_pkey, ADD PRIMARY KEY (" + strings.ToLower(primaryKey) + ")"
	}
	_, err = ss.GetMaster().ExecNoTimeout(alterQuery)
	if err != nil {
		mlog.Critical("Failed to alter primary key", mlog.String("table", tableName), mlog.Err(err))
		time.Sleep(time.Second)
		os.Exit(EXIT_ALTER_PRIMARY_KEY)
	}
	return true
}

func (ss *SqlSupplier) CreateUniqueIndexIfNotExists(indexName string, tableName string, columnName string) bool {
	return ss.createIndexIfNotExists(indexName, tableName, []string{columnName}, INDEX_TYPE_DEFAULT, true)
}

func (ss *SqlSupplier) CreateIndexIfNotExists(indexName string, tableName string, columnName string) bool {
	return ss.createIndexIfNotExists(indexName, tableName, []string{columnName}, INDEX_TYPE_DEFAULT, false)
}

func (ss *SqlSupplier) CreateCompositeIndexIfNotExists(indexName string, tableName string, columnNames []string) bool {
	return ss.createIndexIfNotExists(indexName, tableName, columnNames, INDEX_TYPE_DEFAULT, false)
}

func (ss *SqlSupplier) CreateUniqueCompositeIndexIfNotExists(indexName string, tableName string, columnNames []string) bool {
	return ss.createIndexIfNotExists(indexName, tableName, columnNames, INDEX_TYPE_DEFAULT, true)
}

func (ss *SqlSupplier) CreateFullTextIndexIfNotExists(indexName string, tableName string, columnName string) bool {
	return ss.createIndexIfNotExists(indexName, tableName, []string{columnName}, INDEX_TYPE_FULL_TEXT, false)
}

func (ss *SqlSupplier) createIndexIfNotExists(indexName string, tableName string, columnNames []string, indexType string, unique bool) bool {

	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE "
	}

	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, errExists := ss.GetMaster().SelectStr("SELECT $1::regclass", indexName)
		// It should fail if the index does not exist
		if errExists == nil {
			return false
		}

		query := ""
		if indexType == INDEX_TYPE_FULL_TEXT {
			if len(columnNames) != 1 {
				mlog.Critical("Unable to create multi column full text index")
				os.Exit(EXIT_CREATE_INDEX_POSTGRES)
			}
			columnName := columnNames[0]
			postgresColumnNames := convertMySQLFullTextColumnsToPostgres(columnName)
			query = "CREATE INDEX " + indexName + " ON " + tableName + " USING gin(to_tsvector('english', " + postgresColumnNames + "))"
		} else {
			query = "CREATE " + uniqueStr + "INDEX " + indexName + " ON " + tableName + " (" + strings.Join(columnNames, ", ") + ")"
		}

		_, err := ss.GetMaster().ExecNoTimeout(query)
		if err != nil {
			mlog.Critical("Failed to create index", mlog.Err(errExists), mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_INDEX_POSTGRES)
		}
	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {

		count, err := ss.GetMaster().SelectInt("SELECT COUNT(0) AS index_exists FROM information_schema.statistics WHERE TABLE_SCHEMA = DATABASE() and table_name = ? AND index_name = ?", tableName, indexName)
		if err != nil {
			mlog.Critical("Failed to check index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_INDEX_MYSQL)
		}

		if count > 0 {
			return false
		}

		fullTextIndex := ""
		if indexType == INDEX_TYPE_FULL_TEXT {
			fullTextIndex = " FULLTEXT "
		}

		_, err = ss.GetMaster().ExecNoTimeout("CREATE  " + uniqueStr + fullTextIndex + " INDEX " + indexName + " ON " + tableName + " (" + strings.Join(columnNames, ", ") + ")")
		if err != nil {
			mlog.Critical("Failed to create index", mlog.String("table", tableName), mlog.String("index_name", indexName), mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_INDEX_FULL_MYSQL)
		}
	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		_, err := ss.GetMaster().ExecNoTimeout("CREATE INDEX IF NOT EXISTS " + indexName + " ON " + tableName + " (" + strings.Join(columnNames, ", ") + ")")
		if err != nil {
			mlog.Critical("Failed to create index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_CREATE_INDEX_SQLITE)
		}
	} else {
		mlog.Critical("Failed to create index because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_INDEX_MISSING)
	}

	return true
}

func (ss *SqlSupplier) RemoveIndexIfExists(indexName string, tableName string) bool {

	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, err := ss.GetMaster().SelectStr("SELECT $1::regclass", indexName)
		// It should fail if the index does not exist
		if err != nil {
			return false
		}

		_, err = ss.GetMaster().ExecNoTimeout("DROP INDEX " + indexName)
		if err != nil {
			mlog.Critical("Failed to remove index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_REMOVE_INDEX_POSTGRES)
		}

		return true
	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {

		count, err := ss.GetMaster().SelectInt("SELECT COUNT(0) AS index_exists FROM information_schema.statistics WHERE TABLE_SCHEMA = DATABASE() and table_name = ? AND index_name = ?", tableName, indexName)
		if err != nil {
			mlog.Critical("Failed to check index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_REMOVE_INDEX_MYSQL)
		}

		if count <= 0 {
			return false
		}

		_, err = ss.GetMaster().ExecNoTimeout("DROP INDEX " + indexName + " ON " + tableName)
		if err != nil {
			mlog.Critical("Failed to remove index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_REMOVE_INDEX_MYSQL)
		}
	} else if ss.DriverName() == model.DATABASE_DRIVER_SQLITE {
		_, err := ss.GetMaster().ExecNoTimeout("DROP INDEX IF EXISTS " + indexName)
		if err != nil {
			mlog.Critical("Failed to remove index", mlog.Err(err))
			time.Sleep(time.Second)
			os.Exit(EXIT_REMOVE_INDEX_SQLITE)
		}
	} else {
		mlog.Critical("Failed to create index because of missing driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_REMOVE_INDEX_MISSING)
	}

	return true
}

func IsUniqueConstraintError(err error, indexName []string) bool {
	unique := false
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		unique = true
	}

	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		unique = true
	}

	field := false
	for _, contain := range indexName {
		if strings.Contains(err.Error(), contain) {
			field = true
			break
		}
	}

	return unique && field
}

func (ss *SqlSupplier) GetAllConns() []*gorp.DbMap {
	all := make([]*gorp.DbMap, len(ss.replicas)+1)
	copy(all, ss.replicas)
	all[len(ss.replicas)] = ss.master
	return all
}

// RecycleDBConnections closes active connections by setting the max conn lifetime
// to d, and then resets them back to their original duration.
func (ss *SqlSupplier) RecycleDBConnections(d time.Duration) {
	// Get old time.
	originalDuration := time.Duration(*ss.settings.ConnMaxLifetimeMilliseconds) * time.Millisecond
	// Set the max lifetimes for all connections.
	for _, conn := range ss.GetAllConns() {
		conn.Db.SetConnMaxLifetime(d)
	}
	// Wait for that period with an additional 2 seconds of scheduling delay.
	time.Sleep(d + 2*time.Second)
	// Reset max lifetime back to original value.
	for _, conn := range ss.GetAllConns() {
		conn.Db.SetConnMaxLifetime(originalDuration)
	}
}

func (ss *SqlSupplier) Close() {
	ss.master.Db.Close()
	for _, replica := range ss.replicas {
		replica.Db.Close()
	}
}

func (ss *SqlSupplier) LockToMaster() {
	ss.lockedToMaster = true
}

func (ss *SqlSupplier) UnlockFromMaster() {
	ss.lockedToMaster = false
}

func (ss *SqlSupplier) Team() store.TeamStore {
	return ss.stores.team
}

func (ss *SqlSupplier) Channel() store.ChannelStore {
	return ss.stores.channel
}

func (ss *SqlSupplier) Post() store.PostStore {
	return ss.stores.post
}

func (ss *SqlSupplier) User() store.UserStore {
	return ss.stores.user
}

func (ss *SqlSupplier) Bot() store.BotStore {
	return ss.stores.bot
}

func (ss *SqlSupplier) Session() store.SessionStore {
	return ss.stores.session
}

func (ss *SqlSupplier) Audit() store.AuditStore {
	return ss.stores.audit
}

func (ss *SqlSupplier) ClusterDiscovery() store.ClusterDiscoveryStore {
	return ss.stores.cluster
}

func (ss *SqlSupplier) Compliance() store.ComplianceStore {
	return ss.stores.compliance
}

func (ss *SqlSupplier) OAuth() store.OAuthStore {
	return ss.stores.oauth
}

func (ss *SqlSupplier) System() store.SystemStore {
	return ss.stores.system
}

func (ss *SqlSupplier) Webhook() store.WebhookStore {
	return ss.stores.webhook
}

func (ss *SqlSupplier) Command() store.CommandStore {
	return ss.stores.command
}

func (ss *SqlSupplier) CommandWebhook() store.CommandWebhookStore {
	return ss.stores.commandWebhook
}

func (ss *SqlSupplier) Preference() store.PreferenceStore {
	return ss.stores.preference
}

func (ss *SqlSupplier) License() store.LicenseStore {
	return ss.stores.license
}

func (ss *SqlSupplier) Token() store.TokenStore {
	return ss.stores.token
}

func (ss *SqlSupplier) Emoji() store.EmojiStore {
	return ss.stores.emoji
}

func (ss *SqlSupplier) Status() store.StatusStore {
	return ss.stores.status
}

func (ss *SqlSupplier) FileInfo() store.FileInfoStore {
	return ss.stores.fileInfo
}

func (ss *SqlSupplier) UploadSession() store.UploadSessionStore {
	return ss.stores.uploadSession
}

func (ss *SqlSupplier) Reaction() store.ReactionStore {
	return ss.stores.reaction
}

func (ss *SqlSupplier) Job() store.JobStore {
	return ss.stores.job
}

func (ss *SqlSupplier) UserAccessToken() store.UserAccessTokenStore {
	return ss.stores.userAccessToken
}

func (ss *SqlSupplier) ChannelMemberHistory() store.ChannelMemberHistoryStore {
	return ss.stores.channelMemberHistory
}

func (ss *SqlSupplier) Plugin() store.PluginStore {
	return ss.stores.plugin
}

func (ss *SqlSupplier) Thread() store.ThreadStore {
	return ss.stores.thread
}

func (ss *SqlSupplier) Role() store.RoleStore {
	return ss.stores.role
}

func (ss *SqlSupplier) TermsOfService() store.TermsOfServiceStore {
	return ss.stores.TermsOfService
}

func (ss *SqlSupplier) ProductNotices() store.ProductNoticesStore {
	return ss.stores.productNotices
}

func (ss *SqlSupplier) UserTermsOfService() store.UserTermsOfServiceStore {
	return ss.stores.UserTermsOfService
}

func (ss *SqlSupplier) Scheme() store.SchemeStore {
	return ss.stores.scheme
}

func (ss *SqlSupplier) Group() store.GroupStore {
	return ss.stores.group
}

func (ss *SqlSupplier) LinkMetadata() store.LinkMetadataStore {
	return ss.stores.linkMetadata
}

func (ss *SqlSupplier) DropAllTables() {
	ss.master.TruncateTables()
}

func (ss *SqlSupplier) getQueryBuilder() sq.StatementBuilderType {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	return builder
}

func (ss *SqlSupplier) CheckIntegrity() <-chan model.IntegrityCheckResult {
	results := make(chan model.IntegrityCheckResult)
	go CheckRelationalIntegrity(ss, results)
	return results
}

func (ss *SqlSupplier) UpdateLicense(license *model.License) {
	ss.licenseMutex.Lock()
	defer ss.licenseMutex.Unlock()
	ss.license = license
}

type mattermConverter struct{}

func (me mattermConverter) ToDb(val interface{}) (interface{}, error) {

	switch t := val.(type) {
	case model.StringMap:
		return model.MapToJson(t), nil
	case map[string]string:
		return model.MapToJson(model.StringMap(t)), nil
	case model.StringArray:
		return model.ArrayToJson(t), nil
	case model.StringInterface:
		return model.StringInterfaceToJson(t), nil
	case map[string]interface{}:
		return model.StringInterfaceToJson(model.StringInterface(t)), nil
	case JSONSerializable:
		return t.ToJson(), nil
	case *opengraph.OpenGraph:
		return json.Marshal(t)
	}

	return val, nil
}

func (me mattermConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *model.StringMap:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *map[string]string:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *model.StringArray:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_array"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *model.StringInterface:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_interface"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *map[string]interface{}:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(utils.T("store.sql.convert_string_interface"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	}

	return gorp.CustomScanner{}, false
}

type JSONSerializable interface {
	ToJson() string
}

func convertMySQLFullTextColumnsToPostgres(columnNames string) string {
	columns := strings.Split(columnNames, ", ")
	concatenatedColumnNames := ""
	for i, c := range columns {
		concatenatedColumnNames += c
		if i < len(columns)-1 {
			concatenatedColumnNames += " || ' ' || "
		}
	}

	return concatenatedColumnNames
}

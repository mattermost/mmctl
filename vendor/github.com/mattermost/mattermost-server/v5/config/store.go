// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package config

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils/jsonutils"
)

var (
	// ErrReadOnlyStore is returned when an attempt to modify a read-only
	// configuration store is made.
	ErrReadOnlyStore = errors.New("configuration store is read-only")
)

// Listener is a callback function invoked when the configuration changes.
type Listener func(oldConfig *model.Config, newConfig *model.Config)

type BackingStore interface {
	// Set replaces the current configuration in its entirety and updates the backing store.
	Set(*model.Config) error

	// Load retrieves the configuration stored. If there is no configuration stored
	// the io.ReadCloser will be nil
	Load() ([]byte, error)

	// GetFile fetches the contents of a previously persisted configuration file.
	// If no such file exists, an empty byte array will be returned without error.
	GetFile(name string) ([]byte, error)

	// SetFile sets or replaces the contents of a configuration file.
	SetFile(name string, data []byte) error

	// HasFile returns true if the given file was previously persisted.
	HasFile(name string) (bool, error)

	// RemoveFile removes a previously persisted configuration file.
	RemoveFile(name string) error

	// String describes the backing store for the config.
	String() string

	Watch(callback func()) error

	// Close cleans up resources associated with the store.
	Close() error
}

// NewStore creates a database or file store given a data source name by which to connect.
func NewStore(dsn string, watch, readOnly bool, customDefaults *model.Config) (*Store, error) {
	backingStore, err := getBackingStore(dsn, watch)
	if err != nil {
		return nil, err
	}

	store, err := NewStoreFromBacking(backingStore, customDefaults, readOnly)
	if err != nil {
		backingStore.Close()
		return nil, errors.Wrap(err, "failed to create store")
	}

	return store, nil
}

func NewStoreFromBacking(backingStore BackingStore, customDefaults *model.Config, readOnly bool) (*Store, error) {
	store := &Store{
		backingStore:         backingStore,
		configCustomDefaults: customDefaults,
		readOnly:             readOnly,
	}

	if err := store.Load(); err != nil {
		return nil, errors.Wrap(err, "unable to load on store creation")
	}

	if err := backingStore.Watch(func() {
		store.Load()
	}); err != nil {
		return nil, errors.Wrap(err, "failed to watch backing store")
	}

	return store, nil
}

func getBackingStore(dsn string, watch bool) (BackingStore, error) {
	if IsDatabaseDSN(dsn) {
		return NewDatabaseStore(dsn)
	}

	return NewFileStore(dsn, watch)
}

func NewTestMemoryStore() *Store {
	memoryStore, err := NewMemoryStore()
	if err != nil {
		panic("failed to initialize memory store: " + err.Error())
	}

	configStore, err := NewStoreFromBacking(memoryStore, nil, false)
	if err != nil {
		panic("failed to initialize config store: " + err.Error())
	}

	return configStore
}

type Store struct {
	emitter
	backingStore BackingStore

	configLock           sync.RWMutex
	config               *model.Config
	configNoEnv          *model.Config
	configCustomDefaults *model.Config

	persistFeatureFlags bool
	readOnly            bool
}

// Get fetches the current, cached configuration.
func (s *Store) Get() *model.Config {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return s.config
}

// Get fetches the current, cached configuration without environment variable overrides.
func (s *Store) GetNoEnv() *model.Config {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return s.configNoEnv
}

// GetEnvironmentOverrides fetches the configuration fields overridden by environment variables.
func (s *Store) GetEnvironmentOverrides() map[string]interface{} {
	return generateEnvironmentMap(GetEnvironment())
}

// RemoveEnvironmentOverrides returns a new config without the environment
// overrides
func (s *Store) RemoveEnvironmentOverrides(cfg *model.Config) *model.Config {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return removeEnvOverrides(cfg, s.configNoEnv, s.GetEnvironmentOverrides())
}

// PersistFeatures sets if the store should persist feature flags.
func (s *Store) PersistFeatures(persist bool) {
	s.configLock.Lock()
	defer s.configLock.Unlock()
	s.persistFeatureFlags = persist
}

// Set replaces the current configuration in its entirety and updates the backing store.
func (s *Store) Set(newCfg *model.Config) (*model.Config, error) {
	s.configLock.Lock()
	var unlockOnce sync.Once
	defer unlockOnce.Do(s.configLock.Unlock)

	if s.readOnly {
		return nil, ErrReadOnlyStore
	}

	oldCfg := s.config.Clone()

	// Really just for some tests we need to set defaults here
	newCfg.SetDefaults()

	// Sometimes the config is received with "fake" data in sensitive fields. Apply the real
	// data from the existing config as necessary.
	desanitize(oldCfg, newCfg)

	if err := newCfg.IsValid(); err != nil {
		return nil, errors.Wrap(err, "new configuration is invalid")
	}

	newCfg = removeEnvOverrides(newCfg, s.configNoEnv, s.GetEnvironmentOverrides())

	// Don't persist feature flags unless we are on MM cloud
	// MM cloud uses config in the DB as a cache of the feature flag
	// settings in case the management system is down when a pod starts.
	if !s.persistFeatureFlags {
		newCfg.FeatureFlags = nil
	}

	if err := s.backingStore.Set(newCfg); err != nil {
		return nil, errors.Wrap(err, "failed to persist")
	}

	if err := s.loadLockedWithOld(oldCfg, &unlockOnce); err != nil {
		return nil, errors.Wrap(err, "failed to load on save")
	}

	return oldCfg, nil
}

func (s *Store) loadLockedWithOld(oldCfg *model.Config, unlockOnce *sync.Once) error {
	configBytes, err := s.backingStore.Load()
	if err != nil {
		return err
	}

	loadedConfig := &model.Config{}
	if len(configBytes) != 0 {
		if err = json.Unmarshal(configBytes, &loadedConfig); err != nil {
			return jsonutils.HumanizeJSONError(err, configBytes)
		}
	}

	loadedFeatureFlags := loadedConfig.FeatureFlags

	// If we have custom defaults set, the initial config is merged on
	// top of them and we delete them not to be used again in the
	// configuration reloads
	if s.configCustomDefaults != nil {
		var mErr error
		loadedConfig, mErr = Merge(s.configCustomDefaults, loadedConfig, nil)
		if mErr != nil {
			return errors.Wrap(mErr, "failed to merge custom config defaults")
		}
		s.configCustomDefaults = nil
	}

	loadedConfig.SetDefaults()

	s.configNoEnv = loadedConfig.Clone()
	fixConfig(s.configNoEnv)

	loadedConfig = applyEnvironmentMap(loadedConfig, GetEnvironment())

	fixConfig(loadedConfig)

	if err := loadedConfig.IsValid(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// Apply changes that may have happened on load to the backing store.
	oldCfgBytes, err := json.Marshal(oldCfg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal old config")
	}
	newCfgBytes, err := json.Marshal(loadedConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal loaded config")
	}

	var shouldStore bool
	hasChanged := len(configBytes) == 0 || !bytes.Equal(oldCfgBytes, newCfgBytes)
	if hasChanged {
		featureFlags := s.configNoEnv.FeatureFlags
		// Don't persist feature flags unless we are on MM cloud
		// MM cloud uses config in the DB as a cache of the feature flag
		// settings in case the management system is down when a pod starts.
		if !s.persistFeatureFlags {
			s.configNoEnv.FeatureFlags = loadedFeatureFlags
		}
		toStoreBytes, err := json.Marshal(s.configNoEnv)
		if err != nil {
			return errors.Wrap(err, "failed to marshal old config")
		}
		shouldStore = !bytes.Equal(toStoreBytes, configBytes)
		// We write back to the backing store only if
		// the config has changed and the store is not read-only.
		if !s.readOnly && shouldStore {
			err := s.backingStore.Set(s.configNoEnv)
			s.configNoEnv.FeatureFlags = featureFlags
			if err != nil && !errors.Is(err, ErrReadOnlyConfiguration) {
				return errors.Wrap(err, "failed to persist")
			}
		}
	}

	s.config = loadedConfig

	unlockOnce.Do(s.configLock.Unlock)

	if hasChanged {
		s.invokeConfigListeners(oldCfg, loadedConfig)
	}

	return nil
}

// Load updates the current configuration from the backing store, possibly initializing.
func (s *Store) Load() error {
	s.configLock.Lock()
	var unlockOnce sync.Once
	defer unlockOnce.Do(s.configLock.Unlock)

	oldCfg := s.config.Clone()

	return s.loadLockedWithOld(oldCfg, &unlockOnce)
}

// GetFile fetches the contents of a previously persisted configuration file.
// If no such file exists, an empty byte array will be returned without error.
func (s *Store) GetFile(name string) ([]byte, error) {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return s.backingStore.GetFile(name)
}

// SetFile sets or replaces the contents of a configuration file.
func (s *Store) SetFile(name string, data []byte) error {
	s.configLock.Lock()
	defer s.configLock.Unlock()
	if s.readOnly {
		return ErrReadOnlyStore
	}
	return s.backingStore.SetFile(name, data)
}

// HasFile returns true if the given file was previously persisted.
func (s *Store) HasFile(name string) (bool, error) {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return s.backingStore.HasFile(name)
}

// RemoveFile removes a previously persisted configuration file.
func (s *Store) RemoveFile(name string) error {
	s.configLock.Lock()
	defer s.configLock.Unlock()
	if s.readOnly {
		return ErrReadOnlyStore
	}
	return s.backingStore.RemoveFile(name)
}

// String describes the backing store for the config.
func (s *Store) String() string {
	return s.backingStore.String()
}

// Close cleans up resources associated with the store.
func (s *Store) Close() error {
	s.configLock.Lock()
	defer s.configLock.Unlock()
	return s.backingStore.Close()
}

// IsReadOnly returns whether or not the store is read-only.
func (s *Store) IsReadOnly() bool {
	s.configLock.RLock()
	defer s.configLock.RUnlock()
	return s.readOnly
}

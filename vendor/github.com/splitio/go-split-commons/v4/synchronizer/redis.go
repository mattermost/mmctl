package synchronizer

import (
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/sync"
)

// ManagerImpl struct
type ManagerRedisImpl struct {
	synchronizer Synchronizer
	running      sync.AtomicBool
	logger       logging.LoggerInterface
}

// NewSynchronizerManagerRedis creates new sync manager for redis
func NewSynchronizerManagerRedis(synchronizer Synchronizer, logger logging.LoggerInterface) Manager {
	return &ManagerRedisImpl{
		synchronizer: synchronizer,
		running:      *sync.NewAtomicBool(false),
		logger:       logger,
	}
}

func (m *ManagerRedisImpl) Start() {
	if !m.running.TestAndSet() {
		m.logger.Info("Manager is already running, skipping start")
		return
	}
	m.synchronizer.StartPeriodicDataRecording()
}

func (m *ManagerRedisImpl) Stop() {
	if !m.running.TestAndClear() {
		m.logger.Info("sync manager not yet running, skipping shutdown.")
		return
	}
	m.synchronizer.StopPeriodicDataRecording()
}

func (m *ManagerRedisImpl) IsRunning() bool {
	return m.running.IsSet()
}

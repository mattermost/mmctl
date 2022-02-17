package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-morph/morph/drivers"
)

// Mutex is similar to sync.Mutex, except usable by morph to lock the db.
//
// Pick a unique name for each mutex your plugin requires.
//
// A Mutex must not be copied after first use.
type Mutex struct {
	noCopy
	key string

	// lock guards the variables used to manage the refresh task, and is not itself related to
	// the db lock.
	lock        sync.Mutex
	stopRefresh chan bool
	refreshDone chan bool
	conn        *sql.Conn
}

// NewMutex creates a mutex with the given key name.
//
// returns error if key is empty.
func NewMutex(key string, driver drivers.Driver) (*Mutex, error) {
	key, err := drivers.MakeLockKey(key)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), drivers.TTL)
	defer cancel()

	ms, ok := driver.(*mysql)
	if !ok {
		return nil, errors.New("incorrect implementation of the driver")
	}

	conn, err := ms.db.Conn(context.Background())
	if err != nil {
		return nil, err
	}

	createTableIfNotExistsQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (Id varchar(64) NOT NULL, ExpireAt bigint(20) NOT NULL, PRIMARY KEY (Id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", drivers.MutexTableName)
	if _, err = conn.ExecContext(ctx, createTableIfNotExistsQuery); err != nil {
		return nil, err
	}

	return &Mutex{
		key:  key,
		conn: conn,
	}, nil
}

// lock makes a single attempt to lock the mutex, returning true only if successful.
func (m *Mutex) tryLock(ctx context.Context) (bool, error) {
	now := time.Now()
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("INSERT INTO %s (Id, ExpireAt) VALUES (?, ?)", drivers.MutexTableName)
	if _, err := tx.Exec(query, m.key, now.Add(drivers.TTL).Unix()); err != nil {
		err2 := m.releaseLock(tx, now)
		if err2 == nil { // lock has been released due to expiration
			return true, nil
		}

		return false, fmt.Errorf("failed to lock mutex: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return false, txErr
		}

		return false, err
	}

	return true, nil
}

func (m *Mutex) releaseLock(tx *sql.Tx, t time.Time) error {
	e, err := m.getExpireAt(tx)
	if err != nil {
		return err
	}

	if t.Unix() < e {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("could not rollback: %w", txErr)
		}

		return errors.New("could not release the lock")
	}

	query := fmt.Sprintf("UPDATE %s SET ExpireAt = ? WHERE Id = ?", drivers.MutexTableName)
	if err = executeTx(tx, query, t.Add(drivers.TTL).Unix(), m.key); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("could not rollback transaction: %w", txErr)
		}

		return fmt.Errorf("unable to set new expireat for mutex: %w", err)
	}

	return nil
}

func (m *Mutex) getExpireAt(tx *sql.Tx) (int64, error) {
	var expireAt int64
	query := fmt.Sprintf("SELECT ExpireAt FROM %s WHERE Id = ?", drivers.MutexTableName)
	err := tx.QueryRow(query, m.key).Scan(&expireAt)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return -1, fmt.Errorf("could not rollback: %w", txErr)
		}

		return -1, fmt.Errorf("failed to fetch mutex from db: %w", err)
	}

	return expireAt, nil
}

// refreshLock rewrites the lock key value with a new expiry, returning nil only if successful.
func (m *Mutex) refreshLock(ctx context.Context) error {
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	e, err := m.getExpireAt(tx)
	if err != nil {
		return err
	}

	tmp := time.Unix(e, 0)
	query := fmt.Sprintf("UPDATE %s SET ExpireAt = ? WHERE Id = ?", drivers.MutexTableName)
	if err = executeTx(tx, query, tmp.Add(drivers.TTL).Unix(), m.key); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("could not rollback: %w", txErr)
		}

		return fmt.Errorf("unable to refresh expireat for mutex: %w", err)
	}

	return nil
}

// Lock locks m. If the mutex is already locked by any other morph instance, including the current one,
// the calling goroutine blocks until the mutex can be locked.
func (m *Mutex) Lock() {
	_ = m.LockWithContext(context.Background())
}

// LockWithContext locks m unless the context is canceled. If the mutex is already locked by any other
// instance, including the current one, the calling goroutine blocks until the mutex can be locked,
// or the context is canceled.
//
// The mutex is locked only if a nil error is returned.
func (m *Mutex) LockWithContext(ctx context.Context) error {
	var waitInterval time.Duration

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitInterval):
		}

		ok, err := m.tryLock(ctx)
		if err != nil || !ok {
			waitInterval = drivers.NextWaitInterval(waitInterval, err)
			continue
		}

		break
	}

	stop := make(chan bool)
	done := make(chan bool)
	go func() {
		defer close(done)
		t := time.NewTicker(drivers.RefreshInterval)
		for {
			select {
			case <-t.C:
				err := m.refreshLock(ctx)
				if err != nil {
					return
				}
			case <-stop:
				return
			}
		}
	}()

	m.lock.Lock()
	m.stopRefresh = stop
	m.refreshDone = done
	m.lock.Unlock()

	return nil
}

// Unlock unlocks m. It is a run-time error if m is not locked on entry to Unlock.
//
// Just like sync.Mutex, a locked Lock is not associated with a particular goroutine or a process.
func (m *Mutex) Unlock() {
	m.lock.Lock()
	if m.stopRefresh == nil {
		m.lock.Unlock()
		panic("mutex has not been acquired")
	}

	close(m.stopRefresh)
	m.stopRefresh = nil
	<-m.refreshDone
	m.lock.Unlock()

	defer m.conn.Close()

	// If an error occurs deleting, the mutex will still expire, allowing later retry.
	query := fmt.Sprintf("DELETE FROM %s WHERE Id = ?", drivers.MutexTableName)
	_, _ = m.conn.ExecContext(context.Background(), query, m.key)
}

func executeTx(tx *sql.Tx, query string, args ...interface{}) error {
	if _, err := tx.Exec(query, args...); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("could not rollback transaction: %w", txErr)
		}

		return err
	}

	return nil
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock() {}

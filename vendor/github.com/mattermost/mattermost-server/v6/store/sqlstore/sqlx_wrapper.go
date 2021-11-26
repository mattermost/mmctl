// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

// namedParamRegex is used to capture all named parameters and convert them
// to lowercase. This is necessary to be able to use a single query for both
// Postgres and MySQL.
// This will also lowercase any constant strings containing a :, but sqlx
// will fail the query, so it won't be checked in inadvertently.
var namedParamRegex = regexp.MustCompile(`:\w+`)

type sqlxDBWrapper struct {
	*sqlx.DB
	queryTimeout time.Duration
	trace        bool
}

func newSqlxDBWrapper(db *sqlx.DB, timeout time.Duration, trace bool) *sqlxDBWrapper {
	return &sqlxDBWrapper{
		DB:           db,
		queryTimeout: timeout,
		trace:        trace,
	}
}

func (w *sqlxDBWrapper) Beginx() (*sqlxTxWrapper, error) {
	tx, err := w.DB.Beginx()
	if err != nil {
		return nil, err
	}

	return newSqlxTxWrapper(tx, w.queryTimeout, w.trace), nil
}

func (w *sqlxDBWrapper) BeginXWithIsolation(opts *sql.TxOptions) (*sqlxTxWrapper, error) {
	tx, err := w.DB.BeginTxx(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return newSqlxTxWrapper(tx, w.queryTimeout, w.trace), nil
}

func (w *sqlxDBWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	query = w.DB.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.DB.GetContext(ctx, dest, query, args...)
}

func (w *sqlxDBWrapper) NamedExec(query string, arg interface{}) (sql.Result, error) {
	if w.DB.DriverName() == model.DatabaseDriverPostgres {
		query = namedParamRegex.ReplaceAllStringFunc(query, strings.ToLower)
	}
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), arg)
		}(time.Now())
	}

	return w.DB.NamedExecContext(ctx, query, arg)
}

func (w *sqlxDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	query = w.DB.Rebind(query)

	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.DB.ExecContext(ctx, query, args...)
}

func (w *sqlxDBWrapper) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	if w.DB.DriverName() == model.DatabaseDriverPostgres {
		query = namedParamRegex.ReplaceAllStringFunc(query, strings.ToLower)
	}
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), arg)
		}(time.Now())
	}

	return w.DB.NamedQueryContext(ctx, query, arg)
}

func (w *sqlxDBWrapper) QueryRowX(query string, args ...interface{}) *sqlx.Row {
	query = w.DB.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.DB.QueryRowxContext(ctx, query, args...)
}

func (w *sqlxDBWrapper) QueryX(query string, args ...interface{}) (*sqlx.Rows, error) {
	query = w.DB.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.DB.QueryxContext(ctx, query, args)
}

func (w *sqlxDBWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	query = w.DB.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.DB.SelectContext(ctx, dest, query, args...)
}

type sqlxTxWrapper struct {
	*sqlx.Tx
	queryTimeout time.Duration
	trace        bool
}

func newSqlxTxWrapper(tx *sqlx.Tx, timeout time.Duration, trace bool) *sqlxTxWrapper {
	return &sqlxTxWrapper{
		Tx:           tx,
		queryTimeout: timeout,
		trace:        trace,
	}
}

func (w *sqlxTxWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	query = w.Tx.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.Tx.GetContext(ctx, dest, query, args...)
}

func (w *sqlxTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	query = w.Tx.Rebind(query)

	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.Tx.ExecContext(ctx, query, args...)
}

func (w *sqlxTxWrapper) NamedExec(query string, arg interface{}) (sql.Result, error) {
	if w.Tx.DriverName() == model.DatabaseDriverPostgres {
		query = namedParamRegex.ReplaceAllStringFunc(query, strings.ToLower)
	}
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), arg)
		}(time.Now())
	}

	return w.Tx.NamedExecContext(ctx, query, arg)
}

func (w *sqlxTxWrapper) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	if w.Tx.DriverName() == model.DatabaseDriverPostgres {
		query = namedParamRegex.ReplaceAllStringFunc(query, strings.ToLower)
	}
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), arg)
		}(time.Now())
	}

	// There is no tx.NamedQueryContext support in the sqlx API. (https://github.com/jmoiron/sqlx/issues/447)
	// So we need to implement this ourselves.
	type result struct {
		rows *sqlx.Rows
		err  error
	}

	// Need to add a buffer of 1 to prevent goroutine leak.
	resChan := make(chan *result, 1)
	go func() {
		rows, err := w.Tx.NamedQuery(query, arg)
		resChan <- &result{
			rows: rows,
			err:  err,
		}
	}()

	// staticcheck fails to check that res gets re-assigned later.
	res := &result{} //nolint:staticcheck
	select {
	case res = <-resChan:
	case <-ctx.Done():
		res = &result{
			rows: nil,
			err:  ctx.Err(),
		}
	}

	return res.rows, res.err
}

func (w *sqlxTxWrapper) QueryRowX(query string, args ...interface{}) *sqlx.Row {
	query = w.Tx.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.Tx.QueryRowxContext(ctx, query, args...)
}

func (w *sqlxTxWrapper) QueryX(query string, args ...interface{}) (*sqlx.Rows, error) {
	query = w.Tx.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.Tx.QueryxContext(ctx, query, args)
}

func (w *sqlxTxWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	query = w.Tx.Rebind(query)
	ctx, cancel := context.WithTimeout(context.Background(), w.queryTimeout)
	defer cancel()

	if w.trace {
		defer func(then time.Time) {
			printArgs(query, time.Since(then), args)
		}(time.Now())
	}

	return w.Tx.SelectContext(ctx, dest, query, args...)
}

func removeSpace(r rune) rune {
	// Strip everything except ' '
	// This also strips out more than one space,
	// but we ignore it for now until someone complains.
	if unicode.IsSpace(r) && r != ' ' {
		return -1
	}
	return r
}

func printArgs(query string, dur time.Duration, args ...interface{}) {
	query = strings.Map(removeSpace, query)
	fields := make([]mlog.Field, 0, len(args)+1)
	fields = append(fields, mlog.Duration("duration", dur))
	for i, arg := range args {
		fields = append(fields, mlog.Any("arg"+strconv.Itoa(i), arg))
	}
	mlog.Debug(query, fields...)
}

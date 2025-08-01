package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"github.com/jmoiron/sqlx"
)

type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Transactor interface {
	WithTx(ctx context.Context, opts *sql.TxOptions, fn func(tx Tx) error) error
}

/* DB Transactor */

type DbTranscator struct {
	db *sqlx.DB
}

func NewDbTransactor(db *sqlx.DB) *DbTranscator {
	return &DbTranscator{
		db,
	}
}

// WithTx is a decorator that wraps a function in a transaction
func (t *DbTranscator) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(tx Tx) error) error {
	tx, err := t.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logging.Logger.Error("failed to rollback", "error", rbErr)
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

/* Test Transactor */

type TestTxResult struct{}

func (r *TestTxResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r *TestTxResult) RowsAffected() (int64, error) {
	return 0, nil
}

type TestTx struct{}

func (t *TestTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return &TestTxResult{}, nil
}

func (t *TestTx) QueryContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return &sqlx.Rows{}, nil
}

type TestTransactor struct {
	NumCalls int
}

func (t *TestTransactor) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(tx Tx) error) error {
	tx := &sql.Tx{}

	t.NumCalls += 1

	if err := fn(tx); err != nil {
		return err
	}

	return nil
}

package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/multierr"
)

type txkey string

const key = txkey("tx")

type QueryEngine interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine // tx/pool
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) RunTransaction(
	ctx context.Context,
	txOpts pgx.TxOptions,
	fx func(ctxTX context.Context) error,
) error {

	tx, err := tm.pool.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}

	if err := fx(context.WithValue(ctx, key, tx)); err != nil {
		return multierr.Combine(err, tx.Rollback(ctx))
	}

	if err := tx.Commit(ctx); err != nil {
		return multierr.Combine(err, tx.Rollback(ctx))
	}

	return nil
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(key).(QueryEngine)
	if ok && tx != nil {
		return tx
	}

	return tm.pool
}

// RunReadCommitted execs f func in runTransaction with LevelReadCommitted isolation level
func (tm *TransactionManager) RunReadCommitted(ctx context.Context, f func(txCtx context.Context) error) error {
	return tm.RunTransaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
		// AccessMode: accessMode,
	}, f)
}

// RunRepeatableRead execs f func in runTransaction with LevelRepeatableRead isolation level
func (tm *TransactionManager) RunRepeatableRead(ctx context.Context, f func(txCtx context.Context) error) error {
	return tm.RunTransaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
		// AccessMode: accessMode,
	}, f)
}

// RunSerializable execs f func in runTransaction with LevelSerializable isolation level
func (tm *TransactionManager) RunSerializable(ctx context.Context, f func(txCtx context.Context) error) error {
	return tm.RunTransaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
		// AccessMode: accessMode,
	}, f)
}

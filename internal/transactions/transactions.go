package txdb

import (
	"context"
	"database/sql"
	"fmt"
)

type SqlExecutor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type TxManager struct {
	DB *sql.DB
}

type txKey struct{}

func (tm *TxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	if tx, ok := ctx.Value(txKey{}).(SqlExecutor); ok && tx != nil {
		return fn(ctx)
	}

	tx, err := tm.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	ctxTx := context.WithValue(ctx, txKey{}, tx)
	err = fn(ctxTx)
	switch { //TODO: handle panics here
	case err != nil:
		_ = tx.Rollback()
	default:
		err = tx.Commit()
	}

	return err
}

func GetExecutor(ctx context.Context, db *sql.DB) SqlExecutor {
	if tx, ok := ctx.Value(txKey{}).(SqlExecutor); ok && tx != nil {
		return tx
	}
	return db
}

package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SQLTransactionManager struct {
	db *sqlx.DB
}

func NewSQLTransactionManager(db *sqlx.DB) *SQLTransactionManager {
	return &SQLTransactionManager{db: db}
}

type transactionKey struct{}

func (m SQLTransactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	if tx := getTxFromContext(ctx); tx != nil {
		return fn(ctx)
	}
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ctxWithTx := context.WithValue(ctx, transactionKey{}, tx)

	err = fn(ctxWithTx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func TxOrDb(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := ctx.Value(transactionKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}

func getTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(transactionKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

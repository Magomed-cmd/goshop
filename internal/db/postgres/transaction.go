package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/db"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// PgxTransaction - обертка над pgx.Tx для имплементации db.Transaction
type PgxTransaction struct {
	Tx pgx.Tx // экспортируем для использования в репозиториях
}

func (t *PgxTransaction) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *PgxTransaction) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

// BeginTx начинает транзакцию и возвращает интерфейс db.Transaction
func (tm *TxManager) BeginTx(ctx context.Context) (db.Transaction, error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &PgxTransaction{Tx: tx}, nil
}

// WithTransaction - хелпер для выполнения функции в транзакции
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

package db

import "context"

// Transaction - интерфейс для работы с транзакциями (абстракция от pgx)
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TxManager - интерфейс для управления транзакциями
type TxManager interface {
	BeginTx(ctx context.Context) (Transaction, error)
}

// Querier - общий интерфейс для DB и Tx
type Querier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) interface{}
}

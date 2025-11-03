package pgxrunner

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	portrepo "goshop/internal/core/ports/repositories"
	dtx "goshop/internal/core/ports/transaction"
)

type Runner struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool}
}

func (r *Runner) WithinTransaction(ctx context.Context, fn func(ctx context.Context, conn portrepo.DBConn) error) (err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

var _ dtx.Runner = (*Runner)(nil)

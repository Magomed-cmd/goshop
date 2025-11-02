package transaction

import (
	"context"
	"goshop/internal/domain/repository"
)

type Runner interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, conn repository.DBConn) error) error
}

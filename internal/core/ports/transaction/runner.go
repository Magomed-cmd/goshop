package transaction

import (
	"context"

	portrepo "goshop/internal/core/ports/repositories"
)

type Runner interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, conn portrepo.DBConn) error) error
}

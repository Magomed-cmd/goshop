package pgxrunner

import (
	"context"
	"fmt"

	databaseadapter "goshop/internal/adapters/output/database"
	portrepo "goshop/internal/core/ports/repositories"
	dtx "goshop/internal/core/ports/transaction"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork struct {
	pool    *pgxpool.Pool
	factory *databaseadapter.Factory
}

func NewUnitOfWork(pool *pgxpool.Pool, factory *databaseadapter.Factory) *UnitOfWork {
	return &UnitOfWork{pool: pool, factory: factory}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos dtx.Repositories) error) (err error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	repos := newReposAdapter(u.factory.WithConn(tx))

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(repos); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *UnitOfWork) DoRead(ctx context.Context, fn func(repos dtx.Repositories) error) error {
	repos := newReposAdapter(u.factory.WithPool())
	return fn(repos)
}

type reposAdapter struct {
	set databaseadapter.Set
}

func newReposAdapter(set databaseadapter.Set) *reposAdapter {
	return &reposAdapter{set: set}
}

func (r *reposAdapter) Users() portrepo.UserRepository          { return r.set.Users }
func (r *reposAdapter) Roles() portrepo.RoleRepository          { return r.set.Roles }
func (r *reposAdapter) Addresses() portrepo.AddressRepository   { return r.set.Addresses }
func (r *reposAdapter) Categories() portrepo.CategoryRepository { return r.set.Categories }
func (r *reposAdapter) Products() portrepo.ProductRepository    { return r.set.Products }
func (r *reposAdapter) Carts() portrepo.CartRepository          { return r.set.Carts }
func (r *reposAdapter) Orders() portrepo.OrderRepository        { return r.set.Orders }
func (r *reposAdapter) OrderItems() portrepo.OrderItemRepository {
	return r.set.OrderItems
}
func (r *reposAdapter) Reviews() portrepo.ReviewRepository { return r.set.Reviews }

var _ dtx.UnitOfWork = (*UnitOfWork)(nil)

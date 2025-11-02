package pgx

import (
	"goshop/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Factory struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

type Set struct {
	Users      *UserRepository
	Roles      *RoleRepository
	Addresses  *AddressRepository
	Categories *CategoryRepository
	Products   *ProductRepository
	Carts      *CartRepository
	Orders     *OrderRepository
	OrderItems *OrderItemRepository
	Reviews    *ReviewRepository
}

func NewFactory(pool *pgxpool.Pool, logger *zap.Logger) *Factory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Factory{
		pool:   pool,
		logger: logger,
	}
}

func (f *Factory) WithConn(conn repository.DBConn) Set {
	if conn == nil {
		conn = f.pool
	}

	return Set{
		Users:      NewUserRepository(conn, f.logger),
		Roles:      NewRoleRepository(conn),
		Addresses:  NewAddressRepository(conn),
		Categories: NewCategoryRepository(conn),
		Products:   NewProductRepository(conn, f.logger),
		Carts:      NewCartRepository(conn, f.logger),
		Orders:     NewOrderRepository(conn, f.logger),
		OrderItems: NewOrderItemRepository(conn, f.logger),
		Reviews:    NewReviewRepository(conn, f.logger),
	}
}

func (f *Factory) WithPool() Set {
	return f.WithConn(f.pool)
}

package transaction

import (
	"context"

	portrepo "goshop/internal/core/ports/repositories"
)

type Repositories interface {
	Users() portrepo.UserRepository
	Roles() portrepo.RoleRepository
	Addresses() portrepo.AddressRepository
	Categories() portrepo.CategoryRepository
	Products() portrepo.ProductRepository
	Carts() portrepo.CartRepository
	Orders() portrepo.OrderRepository
	OrderItems() portrepo.OrderItemRepository
	Reviews() portrepo.ReviewRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
	DoRead(ctx context.Context, fn func(repos Repositories) error) error
}

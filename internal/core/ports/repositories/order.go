package repositories

import (
	"context"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/types"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entities.Order) (*int64, error)
	GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) ([]*entities.Order, int64, error)
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (*entities.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) ([]*entities.Order, int64, error)
}

type OrderItemRepository interface {
	Create(ctx context.Context, items []*entities.OrderItem) error
}

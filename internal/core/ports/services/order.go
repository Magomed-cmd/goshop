package services

import (
	"context"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/types"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, addressID *int64) (*entities.Order, error)
	GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) ([]*entities.Order, int64, error)
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (*entities.Order, error)
	CancelOrder(ctx context.Context, userID int64, orderID int64) error
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) ([]*entities.Order, int64, error)
}

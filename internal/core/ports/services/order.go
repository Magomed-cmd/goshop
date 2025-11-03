package services

import (
	"context"

	"goshop/internal/core/domain/types"
	"goshop/internal/dto"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) (*dto.OrdersListResponse, error)
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID int64, orderID int64) error
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) (*dto.OrdersListResponse, error)
}

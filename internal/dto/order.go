package dto

import "time"

type CreateOrderRequest struct {
	AddressID *int64 `json:"address_id" binding:"omitempty"`
}

type OrderResponse struct {
	ID         int64               `json:"id"`
	UUID       string              `json:"uuid"`
	UserID     int64               `json:"user_id"`
	AddressID  *int64              `json:"address_id"`
	TotalPrice string              `json:"total_price"`
	Status     string              `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Items      []OrderItemResponse `json:"items"`
	Address    *AddressResponse    `json:"address,omitempty"`
}

type OrderItemResponse struct {
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	PriceAtOrder string `json:"price_at_order"`
	Quantity     int    `json:"quantity"`
	Subtotal     string `json:"subtotal"`
}

type OrdersListResponse struct {
	Orders      []OrderResponse `json:"orders"`
	TotalCount  int64           `json:"total_count"`
	TotalAmount string          `json:"total_amount"`
	Page        int             `json:"page,omitempty"`
	Limit       int             `json:"limit,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending paid shipped delivered cancelled"`
}

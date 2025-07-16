package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Cart struct {
	ID        int64      `db:"id" json:"id"`
	UUID      uuid.UUID  `db:"uuid" json:"uuid"`
	UserID    int64      `db:"user_id" json:"user_id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	Items     []CartItem `db:"-" json:"items,omitempty"`
}

type CartItem struct {
	CartID    int64    `db:"cart_id" json:"cart_id"`
	ProductID int64    `db:"product_id" json:"product_id"`
	Quantity  int      `db:"quantity" json:"quantity"`
	Product   *Product `db:"-" json:"product,omitempty"`
}

type Order struct {
	ID         int64           `db:"id" json:"id"`
	UUID       uuid.UUID       `db:"uuid" json:"uuid"`
	UserID     int64           `db:"user_id" json:"user_id"`
	AddressID  *int64          `db:"address_id" json:"address_id"`
	TotalPrice decimal.Decimal `db:"total_price" json:"total_price"`
	Status     OrderStatus     `db:"status" json:"status"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updated_at"`
	User       *User           `db:"-" json:"user,omitempty"`
	Address    *UserAddress    `db:"-" json:"address,omitempty"`
	Items      []OrderItem     `db:"-" json:"items,omitempty"`
}

type OrderItem struct {
	OrderID      int64           `db:"order_id" json:"order_id"`
	ProductID    int64           `db:"product_id" json:"product_id"`
	ProductName  string          `db:"product_name" json:"product_name"`
	PriceAtOrder decimal.Decimal `db:"price_at_order" json:"price_at_order"`
	Quantity     int             `db:"quantity" json:"quantity"`
}

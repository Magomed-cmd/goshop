package services

import (
	"context"

	"goshop/internal/core/domain/entities"
)

type CartService interface {
	GetCart(ctx context.Context, userID int64) (*entities.Cart, error)
	AddItem(ctx context.Context, userID int64, productID int64, quantity int) error
	UpdateItem(ctx context.Context, userID int64, productID int64, quantity int) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

package services

import (
	"context"

	"goshop/internal/dto"
)

type CartService interface {
	GetCart(ctx context.Context, userID int64) (*dto.CartResponse, error)
	AddItem(ctx context.Context, userID int64, req *dto.AddToCartRequest) error
	UpdateItem(ctx context.Context, userID int64, productID int64, quantity int) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

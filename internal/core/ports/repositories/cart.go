package repositories

import (
	"context"

	"goshop/internal/core/domain/entities"
)

type CartRepository interface {
	GetUserCart(ctx context.Context, userID int64) (*entities.Cart, error)
	CreateCart(ctx context.Context, cart *entities.Cart) error
	AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error
	UpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error
	RemoveItem(ctx context.Context, cartID int64, productID int64) error
	ClearCart(ctx context.Context, cartID int64) error
}

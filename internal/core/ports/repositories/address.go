package repositories

import (
	"context"

	"goshop/internal/core/domain/entities"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *entities.UserAddress) error
	GetUserAddresses(ctx context.Context, userID int64) ([]*entities.UserAddress, error)
	GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error)
	UpdateAddress(ctx context.Context, address *entities.UserAddress) error
	DeleteAddress(ctx context.Context, addressID int64) error
}

package services

import (
	"context"

	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

type AddressService interface {
	CreateAddress(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*entities.UserAddress, error)
	GetUserAddresses(ctx context.Context, userID int64) ([]*entities.UserAddress, error)
	GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error)
	UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.UpdateAddressRequest) (*entities.UserAddress, error)
	GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*entities.UserAddress, error)
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error
}

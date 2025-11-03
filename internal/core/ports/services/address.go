package services

import (
	"context"

	"goshop/internal/dto"
)

type AddressService interface {
	CreateAddress(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*dto.AddressResponse, error)
	GetUserAddresses(ctx context.Context, userID int64) ([]*dto.AddressResponse, error)
	GetAddressByID(ctx context.Context, addressID int64) (*dto.AddressResponse, error)
	UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.UpdateAddressRequest) (*dto.AddressResponse, error)
	GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*dto.AddressResponse, error)
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error
}

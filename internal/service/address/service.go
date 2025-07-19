package address

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"goshop/internal/dto"
	"goshop/internal/models"
	"time"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *models.UserAddress) error
	GetUserAddresses(ctx context.Context, userID int64) ([]*models.UserAddress, error)
	GetAddressByID(ctx context.Context, addressID int64) (*models.UserAddress, error)
	UpdateAddress(ctx context.Context, address *models.UserAddress) error
	DeleteAddress(ctx context.Context, addressID int64) error
}

//type AddressRepository interface {
//	CreateAddress(ctx context.Context, address *models.UserAddress) error
//	GetUserAddresses(ctx context.Context, userID int64) ([]*models.UserAddress, error)
//	GetAddressByID(ctx context.Context, addressID int64) (*models.UserAddress, error)
//	UpdateAddress(ctx context.Context, address *models.UserAddress) error
//	DeleteAddress(ctx context.Context, addressID int64) error
//}

type AddressService struct {
	addressRepo AddressRepository
}

func NewAddressService(addressRepo AddressRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
	}
}

func (s *AddressService) CreateAddress(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*models.UserAddress, error) {

	address := &models.UserAddress{
		UUID:       uuid.New(),
		UserID:     userID,
		Address:    req.Address,
		City:       req.City,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		CreatedAt:  time.Now(),
	}

	err := s.addressRepo.CreateAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

func (s *AddressService) GetUserAddresses(ctx context.Context, userID int64) ([]*models.UserAddress, error) {
	addresses, err := s.addressRepo.GetUserAddresses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user addresses: %w", err)
	}
	return addresses, nil
}

func (s *AddressService) GetAddressByID(ctx context.Context, addressID int64) (*models.UserAddress, error) {
	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}
	return address, nil
}

func (s *AddressService) UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.UpdateAddressRequest) (*models.UserAddress, error) {

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	if address.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	if req.Address != nil {
		address.Address = *req.Address
	}
	if req.City != nil {
		address.City = req.City
	}
	if req.PostalCode != nil {
		address.PostalCode = req.PostalCode
	}
	if req.Country != nil {
		address.Country = req.Country
	}

	err = s.addressRepo.UpdateAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	return address, nil
}

func (s *AddressService) GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*models.UserAddress, error) {

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	if address.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return address, nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to get address by ID: %w", err)
	}
	if address.UserID != userID {
		return fmt.Errorf("user does not own this address")
	}

	err = s.addressRepo.DeleteAddress(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	return nil
}

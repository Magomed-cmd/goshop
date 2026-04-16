package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/vo"
	"goshop/internal/core/ports/repositories"
)

type AddressService struct {
	addressRepo repositories.AddressRepository
}

func NewAddressService(addressRepo repositories.AddressRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
	}
}

func (s *AddressService) CreateAddress(ctx context.Context, userID int64, address string, city, postalCode, country *string) (*entities.UserAddress, error) {
	parsedUserID, err := vo.NewUserID(userID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	userAddress := &entities.UserAddress{
		UUID:       uuid.New(),
		UserID:     parsedUserID.Int64(),
		Address:    address,
		City:       city,
		PostalCode: postalCode,
		Country:    country,
		CreatedAt:  time.Now(),
	}

	err = s.addressRepo.CreateAddress(ctx, userAddress)
	if err != nil {
		return nil, err
	}

	return userAddress, nil
}

func (s *AddressService) GetUserAddresses(ctx context.Context, userID int64) ([]*entities.UserAddress, error) {
	parsedUserID, err := vo.NewUserID(userID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	addresses, err := s.addressRepo.GetUserAddresses(ctx, parsedUserID.Int64())
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (s *AddressService) GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error) {
	parsedAddressID, err := vo.NewAddressID(addressID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, parsedAddressID.Int64())
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) UpdateAddress(ctx context.Context, userID int64, addressID int64, newAddress, city, postalCode, country *string) (*entities.UserAddress, error) {
	parsedUserID, err := vo.NewUserID(userID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}
	parsedAddressID, err := vo.NewAddressID(addressID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	if newAddress == nil && city == nil && postalCode == nil && country == nil {
		return nil, errors.ErrInvalidAddressData
	}

	address, err := s.addressRepo.GetAddressByID(ctx, parsedAddressID.Int64())
	if err != nil {
		return nil, err
	}

	if address.UserID != parsedUserID.Int64() {
		return nil, errors.ErrForbidden
	}

	if newAddress != nil {
		address.Address = *newAddress
	}
	if city != nil {
		address.City = city
	}
	if postalCode != nil {
		address.PostalCode = postalCode
	}
	if country != nil {
		address.Country = country
	}

	err = s.addressRepo.UpdateAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*entities.UserAddress, error) {
	parsedUserID, err := vo.NewUserID(userID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}
	parsedAddressID, err := vo.NewAddressID(addressID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, parsedAddressID.Int64())
	if err != nil {
		return nil, err
	}

	if address.UserID != parsedUserID.Int64() {
		return nil, errors.ErrForbidden
	}

	return address, nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	parsedUserID, err := vo.NewUserID(userID)
	if err != nil {
		return errors.ErrInvalidInput
	}
	parsedAddressID, err := vo.NewAddressID(addressID)
	if err != nil {
		return errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, parsedAddressID.Int64())
	if err != nil {
		return err
	}

	if address.UserID != parsedUserID.Int64() {
		return errors.ErrForbidden
	}

	err = s.addressRepo.DeleteAddress(ctx, parsedAddressID.Int64())
	if err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/mappers"
	"goshop/internal/core/ports/repositories"
	"goshop/internal/dto"
)

type AddressService struct {
	addressRepo repositories.AddressRepository
}

func NewAddressService(addressRepo repositories.AddressRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
	}
}

func (s *AddressService) CreateAddress(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*dto.AddressResponse, error) {

	if userID <= 0 {
		return nil, errors.ErrInvalidInput
	}

	address := &entities.UserAddress{
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
		return nil, err
	}

	resp := mappers.ToAddressResponse(address)

	return resp, nil
}

func (s *AddressService) GetUserAddresses(ctx context.Context, userID int64) ([]*dto.AddressResponse, error) {

	if userID <= 0 {
		return nil, errors.ErrInvalidInput
	}

	addresses, err := s.addressRepo.GetUserAddresses(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.AddressResponse, 0, len(addresses))
	for _, address := range addresses {
		resp := mappers.ToAddressResponse(address)
		response = append(response, resp)
	}

	return response, nil
}

func (s *AddressService) GetAddressByID(ctx context.Context, addressID int64) (*dto.AddressResponse, error) {
	if addressID <= 0 {
		return nil, errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	resp := mappers.ToAddressResponse(address)

	return resp, nil
}

func (s *AddressService) UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.UpdateAddressRequest) (*dto.AddressResponse, error) {
	if userID <= 0 || addressID <= 0 {
		return nil, errors.ErrInvalidInput
	}

	if req.Address == nil && req.City == nil && req.PostalCode == nil && req.Country == nil {
		return nil, errors.ErrInvalidAddressData
	}

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.UserID != userID {
		return nil, errors.ErrForbidden
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
		return nil, err
	}

	resp := mappers.ToAddressResponse(address)

	return resp, nil
}

func (s *AddressService) GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*dto.AddressResponse, error) {
	if userID <= 0 || addressID <= 0 {
		return nil, errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.UserID != userID {
		return nil, errors.ErrForbidden
	}

	resp := mappers.ToAddressResponse(address)

	return resp, nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	if userID <= 0 || addressID <= 0 {
		return errors.ErrInvalidInput
	}

	address, err := s.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address.UserID != userID {
		return errors.ErrForbidden
	}

	err = s.addressRepo.DeleteAddress(ctx, addressID)
	if err != nil {
		return err
	}

	return nil
}

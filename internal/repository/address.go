package repository

import (
	"context"
	"gorm.io/gorm"
	"goshop/internal/models"
)

type AddressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}

}

// CreateAddress создает новый адрес
func (r *AddressRepository) CreateAddress(ctx context.Context, address *models.UserAddress) error {
	result := r.db.WithContext(ctx).Create(address)
	return result.Error
}

func (r *AddressRepository) GetUserAddresses(ctx context.Context, userID int64) ([]*models.UserAddress, error) {
	var addresses []*models.UserAddress
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses)
	if result.Error != nil {
		return nil, result.Error
	}
	return addresses, nil
}

func (r *AddressRepository) GetAddressByID(ctx context.Context, addressID int64) (*models.UserAddress, error) {
	var address models.UserAddress
	result := r.db.WithContext(ctx).First(&address, addressID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &address, nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, address *models.UserAddress) error {
	result := r.db.WithContext(ctx).Save(address)
	return result.Error
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, addressID int64) error {
	result := r.db.WithContext(ctx).Delete(&models.UserAddress{}, addressID)
	return result.Error
}

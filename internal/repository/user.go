package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"goshop/internal/models"
)

type UserRepository struct {
	dbConnection *gorm.DB
}

func NewUserRepository(conn *gorm.DB) *UserRepository {
	return &UserRepository{dbConnection: conn}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.dbConnection.WithContext(ctx).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.dbConnection.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

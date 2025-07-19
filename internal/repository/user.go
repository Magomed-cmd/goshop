package repository

import (
	"context"
	"fmt"
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
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	result := r.dbConnection.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error {
	updates := make(map[string]interface{})

	if name != nil {
		updates["name"] = *name
	}
	if phone != nil {
		updates["phone"] = *phone
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update") // Добавь проверку!
	}
	query := r.dbConnection.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID)

	return query.Updates(updates).Error
}

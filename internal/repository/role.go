package repository

import (
	"context"
	"gorm.io/gorm"
	"goshop/internal/models"
)

type RoleRepository struct {
	dbConnection *gorm.DB
}

func NewRoleRepository(conn *gorm.DB) *RoleRepository {
	return &RoleRepository{dbConnection: conn}
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*models.Role, error) {
	var role models.Role
	result := r.dbConnection.WithContext(ctx).First(&role, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	result := r.dbConnection.WithContext(ctx).Where("name = ?", name).First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

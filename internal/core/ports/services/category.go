package services

import (
	"context"

	"goshop/internal/core/domain/entities"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error)
	CreateCategory(ctx context.Context, name string, description *string) (*entities.Category, error)
	UpdateCategory(ctx context.Context, id int64, name *string, description *string) (*entities.CategoryWithCount, error)
	DeleteCategory(ctx context.Context, id int64) error
}

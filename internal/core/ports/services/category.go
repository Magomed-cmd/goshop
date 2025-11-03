package services

import (
	"context"

	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error)
	GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

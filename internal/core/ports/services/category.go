package services

import (
	"context"

	"goshop/internal/dto"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error)
	GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id int64) error
}

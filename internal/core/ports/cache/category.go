package cache

import (
	"context"
	"time"

	"goshop/internal/dto"
)

type CategoryCache interface {
	GetCategory(ctx context.Context, categoryID int64) (*dto.CategoryResponse, error)
	SetCategory(ctx context.Context, category *dto.CategoryResponse, ttl time.Duration) error
	GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error)
	SetAllCategories(ctx context.Context, response *dto.CategoriesListResponse, ttl time.Duration) error
	DeleteCategory(ctx context.Context, categoryID int64) error
	DeleteAllCategories(ctx context.Context) error
}

package cache

import (
	"context"
	"time"

	"goshop/internal/core/domain/types"
	"goshop/internal/dto"
)

type ProductCache interface {
	SetProduct(ctx context.Context, product *dto.ProductResponse, ttl time.Duration) error
	GetProduct(ctx context.Context, productID int64) (*dto.ProductResponse, error)
	InvalidateProduct(ctx context.Context, productID int64) error
	SetProductsWithFilters(ctx context.Context, filters types.ProductFilters, products *dto.ProductCatalogResponse, ttl time.Duration) error
	GetProductsWithFilters(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error)
	InvalidateProductLists(ctx context.Context) error
	InvalidateProductsByCategory(ctx context.Context, categoryID int64) error
}

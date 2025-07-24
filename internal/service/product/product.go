package product

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
)

type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, product *entities.Product) error
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
	UpdateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error)
	AddProductToCategories(ctx context.Context, productID int64, categoryIDs []int64) error
	RemoveProductFromCategories(ctx context.Context, productID int64) error
	GetProductCategories(ctx context.Context, productID int64) ([]*entities.Category, error)
}

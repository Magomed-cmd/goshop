package services

import (
	"context"
	"io"

	"github.com/shopspring/decimal"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/types"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *entities.Product, categoryIDs []int64) (*entities.Product, []*entities.Category, error)
	GetProductByID(ctx context.Context, id int64) (*entities.Product, []*entities.Category, []*entities.ProductImage, error)
	UpdateProduct(ctx context.Context, id int64, name *string, description *string, price *decimal.Decimal, stock *int, categoryIDs []int64) (*entities.Product, []*entities.Category, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error)
	SaveProductImg(ctx context.Context, reader io.ReadCloser, size, productID int64, contentType, extension string) (*string, error)
	DeleteProductImg(ctx context.Context, productID, imgID int64) error
}

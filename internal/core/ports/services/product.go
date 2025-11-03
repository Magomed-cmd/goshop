package services

import (
	"context"
	"io"

	"goshop/internal/core/domain/types"
	"goshop/internal/dto"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id int64) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error)
	SaveProductImg(ctx context.Context, reader io.ReadCloser, size, productID int64, contentType, extension string) (*string, error)
	DeleteProductImg(ctx context.Context, productID, imgID int64) error
}

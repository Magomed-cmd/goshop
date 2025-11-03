package repositories

import (
    "context"

    "goshop/internal/core/domain/entities"
    "goshop/internal/core/domain/types"
)

type ProductRepository interface {
    CreateProduct(ctx context.Context, product *entities.Product) error
    GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
    UpdateProduct(ctx context.Context, product *entities.Product) error
    DeleteProduct(ctx context.Context, id int64) error
    GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error)
    AddProductToCategories(ctx context.Context, productID int64, categoryIDs []int64) error
    RemoveProductFromCategories(ctx context.Context, productID int64) error
    GetProductCategories(ctx context.Context, productID int64) ([]*entities.Category, error)
    SaveProductImage(ctx context.Context, productImage *entities.ProductImage) (int64, int, error)
    GetProductImgs(ctx context.Context, productID int64) ([]*entities.ProductImage, error)
    DeleteProductImg(ctx context.Context, productID, imageID int64) error
}


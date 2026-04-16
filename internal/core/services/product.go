package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	cacheports "goshop/internal/core/ports/cache"
	repositories "goshop/internal/core/ports/repositories"
	storageports "goshop/internal/core/ports/storage"
	"goshop/internal/validation"
)

const (
	ProductCacheTTL     = 1 * time.Hour
	ProductListCacheTTL = 15 * time.Minute
	CategoryCacheTTL    = 24 * time.Hour
)

type ProductService struct {
	ProductRepo  repositories.ProductRepository
	CategoryRepo repositories.CategoryRepository
	ImgStorage   storageports.ImgStorage
	ProductCache cacheports.ProductCache
	logger       *zap.Logger
}

func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository, imgStorage storageports.ImgStorage, cache cacheports.ProductCache, logger *zap.Logger) *ProductService {
	return &ProductService{
		ProductRepo:  productRepo,
		CategoryRepo: categoryRepo,
		ProductCache: cache,
		ImgStorage:   imgStorage,
		logger:       logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *entities.Product, categoryIDs []int64) (*entities.Product, []*entities.Category, error) {
	s.logger.Info("Creating product", zap.String("product_name", product.Name))

	if err := validation.ValidateProductName(product.Name); err != nil {
		return nil, nil, err
	}
	if err := validation.ValidateProductDescription(product.Description); err != nil {
		return nil, nil, err
	}
	if err := validation.ValidateProductPrice(product.Price); err != nil {
		return nil, nil, err
	}
	if err := validation.ValidateProductStock(product.Stock); err != nil {
		return nil, nil, err
	}
	if len(categoryIDs) == 0 {
		return nil, nil, errors.ErrInvalidInput
	}

	exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, categoryIDs)
	if err != nil {
		s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Any("category_ids", categoryIDs))
		return nil, nil, err
	}
	if !exists {
		s.logger.Warn("Some categories not found", zap.Any("category_ids", categoryIDs))
		return nil, nil, errors.ErrCategoryNotFound
	}

	now := time.Now()
	product.UUID = uuid.New()
	product.Name = strings.TrimSpace(product.Name)
	product.CreatedAt = now
	product.UpdatedAt = now

	if err := s.ProductRepo.CreateProduct(ctx, product); err != nil {
		s.logger.Error("Failed to create product in repository", zap.Error(err), zap.String("product_name", product.Name))
		return nil, nil, err
	}

	if err := s.ProductRepo.AddProductToCategories(ctx, product.ID, categoryIDs); err != nil {
		s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", product.ID))
		return nil, nil, err
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", product.ID))
		return nil, nil, err
	}

	s.logger.Info("Product created successfully", zap.Int64("product_id", product.ID), zap.String("product_name", product.Name))
	return product, categories, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*entities.Product, []*entities.Category, []*entities.ProductImage, error) {
	s.logger.Debug("Getting product by ID", zap.Int64("product_id", id))

	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, nil, err
	}

	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, nil, err
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, nil, err
	}

	productImgs, err := s.ProductRepo.GetProductImgs(ctx, product.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	s.logger.Debug("Product retrieved successfully", zap.Int64("product_id", id), zap.String("product_name", product.Name))
	return product, categories, productImgs, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, name *string, description *string, price *decimal.Decimal, stock *int, categoryIDs []int64) (*entities.Product, []*entities.Category, error) {
	s.logger.Info("Updating product", zap.Int64("product_id", id))

	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, err
	}

	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, err
	}

	hasChanges := false

	if name != nil {
		product.Name = strings.TrimSpace(*name)
		hasChanges = true
	}
	if description != nil {
		product.Description = description
		hasChanges = true
	}
	if price != nil {
		product.Price = *price
		hasChanges = true
	}
	if stock != nil {
		product.Stock = *stock
		hasChanges = true
	}

	if !hasChanges && len(categoryIDs) == 0 {
		s.logger.Warn("No changes provided for product update", zap.Int64("product_id", id))
		return nil, nil, errors.ErrInvalidInput
	}

	if hasChanges {
		product.UpdatedAt = time.Now()
		if err := s.ProductRepo.UpdateProduct(ctx, product); err != nil {
			s.logger.Error("Failed to update product in repository", zap.Error(err), zap.Int64("product_id", id))
			return nil, nil, err
		}
	}

	if len(categoryIDs) > 0 {
		exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, categoryIDs)
		if err != nil {
			s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Int64("product_id", id))
			return nil, nil, err
		}
		if !exists {
			s.logger.Warn("Some categories not found", zap.Int64("product_id", id), zap.Any("category_ids", categoryIDs))
			return nil, nil, errors.ErrCategoryNotFound
		}

		if err := s.ProductRepo.RemoveProductFromCategories(ctx, product.ID); err != nil {
			s.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, nil, err
		}
		if err := s.ProductRepo.AddProductToCategories(ctx, product.ID, categoryIDs); err != nil {
			s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, nil, err
		}
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", id))
		return nil, nil, err
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.InvalidateProduct(ctx, product.ID); err != nil {
			s.logger.Error("Failed to invalidate product cache", zap.Error(err), zap.Int64("product_id", product.ID))
		}
	}

	s.logger.Info("Product updated successfully", zap.Int64("product_id", id), zap.String("product_name", product.Name))

	return product, categories, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	s.logger.Info("Deleting product", zap.Int64("product_id", id))

	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	if err := s.ProductRepo.RemoveProductFromCategories(ctx, id); err != nil {
		s.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	if err := s.ProductRepo.DeleteProduct(ctx, id); err != nil {
		s.logger.Error("Failed to delete product from repository", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.InvalidateProduct(ctx, id); err != nil {
			s.logger.Error("Failed to invalidate product cache", zap.Error(err), zap.Int64("product_id", id))
		}
	}

	s.logger.Info("Product deleted successfully", zap.Int64("product_id", id))
	return nil
}

func (s *ProductService) GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error) {
	s.logger.Debug("Getting products with filters", zap.Any("filters", filters))

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	products, total, err := s.ProductRepo.GetProducts(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to get products from repository", zap.Error(err), zap.Any("filters", filters))
		return nil, 0, err
	}

	s.logger.Info("Products retrieved successfully",
		zap.Int("products_count", len(products)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
		zap.Int("limit", filters.Limit))

	return products, total, nil
}

func (s *ProductService) SaveProductImg(ctx context.Context, reader io.ReadCloser, size, productID int64, contentType, extension string) (*string, error) {

	s.logger.Info("Saving product image",
		zap.Int64("product_id", productID),
		zap.String("content_type", contentType),
		zap.String("extension", extension),
		zap.Int64("size", size))

	imgUUID := uuid.New()
	objectName := fmt.Sprintf("products/%d/%s.%s", productID, imgUUID.String(), extension)

	imageURL, err := s.ImgStorage.UploadImage(ctx, objectName, reader, size, contentType)
	if err != nil {
		s.logger.Error("Failed to upload image to storage", zap.Error(err))
		return nil, errors.ErrProductImageUploadFail
	}
	if imageURL == nil {
		s.logger.Error("Storage returned empty image URL", zap.Int64("product_id", productID))
		return nil, errors.ErrProductImageUploadFail
	}

	productImgInfo := &entities.ProductImage{
		ID:        0,
		ProductID: productID,
		ImageURL:  *imageURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UUID:      imgUUID,
	}

	s.logger.Info("Product image uploaded to storage",
		zap.Int64("product_id", productID),
		zap.String("image_url", *imageURL),
		zap.Int64("image_id", productImgInfo.ID))

	id, position, err := s.ProductRepo.SaveProductImage(ctx, productImgInfo)
	if err != nil {
		_ = s.ImgStorage.DeleteImage(ctx, objectName)
		return nil, err
	}

	s.logger.Info("Product image saved in repository",
		zap.Int64("image_id", productImgInfo.ID),
		zap.Int64("product_id", productID),
		zap.Int("position", productImgInfo.Position),
		zap.String("image_url", *imageURL),
	)

	productImgInfo.ID = id
	productImgInfo.Position = position
	s.logger.Info("Product image saved in repository")

	return imageURL, nil
}

func (s *ProductService) DeleteProductImg(ctx context.Context, productID, imgID int64) error {
	s.logger.Info("Deleting product image", zap.Int64("product_id", productID), zap.Int64("image_id", imgID))

	if err := validation.ValidateProductID(productID); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", productID))
		return errors.ErrInvalidProductID
	}

	if imgID <= 0 {
		s.logger.Error("Invalid image ID for deletion", zap.Int64("image_id", imgID))
		return errors.ErrInvalidInput
	}

	productImgs, err := s.ProductRepo.GetProductImgs(ctx, productID)
	if err != nil {
		s.logger.Error("Failed to get product images from repository", zap.Error(err), zap.Int64("product_id", productID))
		return err
	}

	var imgToDelete *entities.ProductImage
	for _, img := range productImgs {
		if img.ID == imgID {
			imgToDelete = img
			break
		}
	}

	if imgToDelete == nil {
		s.logger.Warn("Product image not found for deletion", zap.Int64("product_id", productID), zap.Int64("image_id", imgID))
		return errors.ErrProductImageNotFound
	}

	if err := s.ImgStorage.DeleteImage(ctx, imgToDelete.UUID.String()); err != nil {
		s.logger.Error("Failed to delete image from storage", zap.Error(err), zap.String("image_url", imgToDelete.ImageURL))
		return errors.ErrProductImageDeleteFail
	}

	if err := s.ProductRepo.DeleteProductImg(ctx, productID, imgToDelete.ID); err != nil {
		s.logger.Error("Failed to delete product image from repository", zap.Error(err), zap.Int64("image_id", imgToDelete.ID))
		return errors.ErrProductImageDeleteFail
	}

	s.logger.Info("Product image deleted successfully",
		zap.Int64("product_id", productID),
		zap.Int64("image_id", imgToDelete.ID))

	return nil
}

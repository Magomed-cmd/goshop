package product

import (
	"context"
	"fmt"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/errors"
	"goshop/internal/domain/types"
	"goshop/internal/dto"
	"goshop/internal/validation"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	ProductCacheTTL     = 1 * time.Hour
	ProductListCacheTTL = 15 * time.Minute
	CategoryCacheTTL    = 24 * time.Hour
)

type CategoryRepository interface {
	CheckCategoriesExist(ctx context.Context, categoryIDs []int64) (bool, error)
}

type ImgStorage interface {
	UploadImage(ctx context.Context, objectName string, reader io.ReadCloser, size int64, contentType string) (*string, error)
	DeleteImage(ctx context.Context, objectName string) error
	GetImageURL(objectName string) string
	DownloadImage(ctx context.Context, objectName string) (io.ReadCloser, error)
}

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

type ProductCache interface {
	SetProduct(ctx context.Context, product *dto.ProductResponse, ttl time.Duration) error
	GetProduct(ctx context.Context, productID int64) (*dto.ProductResponse, error)
	InvalidateProduct(ctx context.Context, productID int64) error
	SetProductsWithFilters(ctx context.Context, filters types.ProductFilters, products *dto.ProductCatalogResponse, ttl time.Duration) error
	GetProductsWithFilters(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error)

	InvalidateProductLists(ctx context.Context) error
	InvalidateProductsByCategory(ctx context.Context, categoryID int64) error
}

type ProductService struct {
	ProductRepo  ProductRepository
	CategoryRepo CategoryRepository
	ImgStorage   ImgStorage
	ProductCache ProductCache
	logger       *zap.Logger
}

func NewProductService(productRepo ProductRepository, categoryRepo CategoryRepository, imgStorage ImgStorage, cache ProductCache, logger *zap.Logger) *ProductService {
	return &ProductService{
		ProductRepo:  productRepo,
		CategoryRepo: categoryRepo,
		ProductCache: cache,
		ImgStorage:   imgStorage,
		logger:       logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	s.logger.Info("Creating product", zap.String("product_name", req.Name))

	if err := validation.ValidateCreateProduct(req); err != nil {
		s.logger.Error("Create product validation failed", zap.Error(err), zap.String("product_name", req.Name))
		return nil, err
	}

	exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, req.CategoryIDs)
	if err != nil {
		s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Any("category_ids", req.CategoryIDs))
		return nil, err
	}
	if !exists {
		s.logger.Warn("Some categories not found", zap.Any("category_ids", req.CategoryIDs))
		return nil, errors.ErrCategoryNotFound
	}

	now := time.Now()
	product := &entities.Product{
		UUID:        uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.ProductRepo.CreateProduct(ctx, product); err != nil {
		s.logger.Error("Failed to create product in repository", zap.Error(err), zap.String("product_name", req.Name))
		return nil, err
	}

	if err := s.ProductRepo.AddProductToCategories(ctx, product.ID, req.CategoryIDs); err != nil {
		s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", product.ID))
		return nil, err
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", product.ID))
		return nil, err
	}

	categoryResponses := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		categoryResponses[i] = dto.CategoryResponse{
			ID:          cat.ID,
			UUID:        cat.UUID.String(),
			Name:        cat.Name,
			Description: cat.Description,
		}
	}

	resp := &dto.ProductResponse{
		ID:          product.ID,
		UUID:        product.UUID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price.StringFixed(2),
		Stock:       product.Stock,
		Categories:  categoryResponses,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.SetProduct(ctx, resp, ProductCacheTTL); err != nil {
			s.logger.Error("Failed to cache product after creation", zap.Error(err), zap.Int64("product_id", product.ID))
		}
	}

	s.logger.Info("Product created successfully", zap.Int64("product_id", product.ID), zap.String("product_name", product.Name))
	return resp, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*dto.ProductResponse, error) {
	s.logger.Debug("Getting product by ID", zap.Int64("product_id", id))

	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	if s.ProductCache != nil {
		if res, err := s.ProductCache.GetProduct(ctx, id); err == nil && res != nil {
			s.logger.Debug("Cache hit: product retrieved from cache",
				zap.Int64("product_id", id),
				zap.String("product_name", res.Name))
			return res, nil
		}
	}

	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	categoryResponses := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		categoryResponses[i] = dto.CategoryResponse{
			ID:          cat.ID,
			UUID:        cat.UUID.String(),
			Name:        cat.Name,
			Description: cat.Description,
		}
	}

	productImgs, err := s.ProductRepo.GetProductImgs(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	resp := &dto.ProductResponse{
		ID:          product.ID,
		UUID:        product.UUID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price.StringFixed(2),
		Stock:       product.Stock,
		ProductImgs: productImgs,
		Categories:  categoryResponses,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.SetProduct(ctx, resp, ProductCacheTTL); err != nil {
			s.logger.Error("Failed to cache product", zap.Error(err), zap.Int64("product_id", id))
		}
	}

	s.logger.Debug("Product retrieved successfully", zap.Int64("product_id", id), zap.String("product_name", product.Name))
	return resp, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	s.logger.Info("Updating product", zap.Int64("product_id", id))

	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	if err := validation.ValidateUpdateProduct(req); err != nil {
		s.logger.Error("Update product validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	hasChanges := false

	if req.Name != nil {
		product.Name = strings.TrimSpace(*req.Name)
		hasChanges = true
	}
	if req.Description != nil {
		product.Description = req.Description
		hasChanges = true
	}
	if req.Price != nil {
		product.Price = *req.Price
		hasChanges = true
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
		hasChanges = true
	}

	if !hasChanges && len(req.CategoryIDs) == 0 {
		s.logger.Warn("No changes provided for product update", zap.Int64("product_id", id))
		return nil, errors.ErrInvalidInput
	}

	if hasChanges {
		product.UpdatedAt = time.Now()
		if err := s.ProductRepo.UpdateProduct(ctx, product); err != nil {
			s.logger.Error("Failed to update product in repository", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
	}

	if len(req.CategoryIDs) > 0 {
		exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, req.CategoryIDs)
		if err != nil {
			s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
		if !exists {
			s.logger.Warn("Some categories not found", zap.Int64("product_id", id), zap.Any("category_ids", req.CategoryIDs))
			return nil, errors.ErrCategoryNotFound
		}

		if err := s.ProductRepo.RemoveProductFromCategories(ctx, product.ID); err != nil {
			s.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
		if err := s.ProductRepo.AddProductToCategories(ctx, product.ID, req.CategoryIDs); err != nil {
			s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
	}

	categories, err := s.ProductRepo.GetProductCategories(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	categoryResponses := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		categoryResponses[i] = dto.CategoryResponse{
			ID:          cat.ID,
			UUID:        cat.UUID.String(),
			Name:        cat.Name,
			Description: cat.Description,
		}
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.InvalidateProduct(ctx, product.ID); err != nil {
			s.logger.Error("Failed to invalidate product cache", zap.Error(err), zap.Int64("product_id", product.ID))
		}
	}

	s.logger.Info("Product updated successfully", zap.Int64("product_id", id), zap.String("product_name", product.Name))

	return &dto.ProductResponse{
		ID:          product.ID,
		UUID:        product.UUID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price.StringFixed(2),
		Stock:       product.Stock,
		Categories:  categoryResponses,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}, nil
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

func (s *ProductService) GetProducts(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error) {
	s.logger.Debug("Getting products with filters", zap.Any("filters", filters))

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	if s.ProductCache != nil {
		if res, err := s.ProductCache.GetProductsWithFilters(ctx, filters); err == nil && res != nil {
			s.logger.Debug("Cache hit: products retrieved from cache",
				zap.Int64p("category_id", filters.CategoryID),
				zap.Stringp("sort_by", filters.SortBy),
				zap.Int("page", filters.Page),
				zap.Int("limit", filters.Limit))
			return res, nil
		}
	}

	products, total, err := s.ProductRepo.GetProducts(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to get products from repository", zap.Error(err), zap.Any("filters", filters))
		return nil, err
	}

	productResponses := make([]dto.ProductCatalogItem, len(products))
	for i, p := range products {
		productResponses[i] = dto.ProductCatalogItem{
			ID:    p.ID,
			UUID:  p.UUID.String(),
			Name:  p.Name,
			Price: p.Price.StringFixed(2),
			Stock: p.Stock,
		}
	}

	resp := &dto.ProductCatalogResponse{
		Products: productResponses,
		Total:    total,
		Page:     filters.Page,
		Limit:    filters.Limit,
	}

	if s.ProductCache != nil {
		if err := s.ProductCache.SetProductsWithFilters(ctx, filters, resp, ProductListCacheTTL); err != nil {
			s.logger.Error("Failed to cache products list", zap.Error(err))
		}
	}

	s.logger.Info("Products retrieved successfully",
		zap.Int("products_count", len(products)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
		zap.Int("limit", filters.Limit))

	return resp, nil
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
		return nil, fmt.Errorf("failed to upload product image: %w", err)
	}
	if imageURL == nil {
		return nil, fmt.Errorf("image URL is nil after upload")
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
		return fmt.Errorf("failed to delete image from storage: %w", err)
	}

	if err := s.ProductRepo.DeleteProductImg(ctx, productID, imgToDelete.ID); err != nil {
		s.logger.Error("Failed to delete product image from repository", zap.Error(err), zap.Int64("image_id", imgToDelete.ID))
		return errors.ErrProductImageNotFound
	}

	s.logger.Info("Product image deleted successfully",
		zap.Int64("product_id", productID),
		zap.Int64("image_id", imgToDelete.ID))

	return nil
}

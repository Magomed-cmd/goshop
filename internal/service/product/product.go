package product

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"goshop/internal/validation"
	"strings"
	"time"
)

type CategoryRepository interface {
	CheckCategoriesExist(ctx context.Context, categoryIDs []int64) (bool, error)
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
}

type ProductService struct {
	ProductRepo  ProductRepository
	CategoryRepo CategoryRepository
	logger       *zap.Logger
}

func NewProductService(productRepo ProductRepository, categoryRepo CategoryRepository, logger *zap.Logger) *ProductService {
	return &ProductService{
		ProductRepo:  productRepo,
		CategoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	s.logger.Info("Creating product", zap.String("product_name", req.Name))

	s.logger.Debug("Validating create product request")
	if err := validation.ValidateCreateProduct(req); err != nil {
		s.logger.Error("Create product validation failed", zap.Error(err), zap.String("product_name", req.Name))
		return nil, err
	}

	s.logger.Debug("Checking if categories exist", zap.Any("category_ids", req.CategoryIDs))
	exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, req.CategoryIDs)
	if err != nil {
		s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Any("category_ids", req.CategoryIDs))
		return nil, err
	}
	if !exists {
		s.logger.Warn("Some categories not found", zap.Any("category_ids", req.CategoryIDs))
		return nil, domain_errors.ErrCategoryNotFound
	}

	product := &entities.Product{
		UUID:        uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.logger.Debug("Creating product in repository", zap.String("product_uuid", product.UUID.String()))
	err = s.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		s.logger.Error("Failed to create product in repository", zap.Error(err), zap.String("product_name", req.Name))
		return nil, err
	}

	s.logger.Debug("Adding product to categories", zap.Int64("product_id", product.ID), zap.Any("category_ids", req.CategoryIDs))
	err = s.ProductRepo.AddProductToCategories(ctx, product.ID, req.CategoryIDs)
	if err != nil {
		s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", product.ID))
		return nil, err
	}

	s.logger.Debug("Getting product categories", zap.Int64("product_id", product.ID))
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

	s.logger.Info("Product created successfully", zap.Int64("product_id", product.ID), zap.String("product_name", product.Name))

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

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*dto.ProductResponse, error) {
	s.logger.Debug("Getting product by ID", zap.Int64("product_id", id))

	s.logger.Debug("Validating product ID", zap.Int64("product_id", id))
	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	s.logger.Debug("Getting product from repository", zap.Int64("product_id", id))
	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	s.logger.Debug("Getting product categories", zap.Int64("product_id", id))
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

	s.logger.Debug("Product retrieved successfully", zap.Int64("product_id", id), zap.String("product_name", product.Name))

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

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	s.logger.Info("Updating product", zap.Int64("product_id", id))

	s.logger.Debug("Validating product ID", zap.Int64("product_id", id))
	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	s.logger.Debug("Validating update product request", zap.Int64("product_id", id))
	if err := validation.ValidateUpdateProduct(req); err != nil {
		s.logger.Error("Update product validation failed", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	s.logger.Debug("Getting product from repository", zap.Int64("product_id", id))
	product, err := s.ProductRepo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get product from repository", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	hasChanges := false

	if req.Name != nil {
		s.logger.Debug("Updating product name", zap.Int64("product_id", id), zap.String("new_name", *req.Name))
		product.Name = strings.TrimSpace(*req.Name)
		hasChanges = true
	}

	if req.Description != nil {
		s.logger.Debug("Updating product description", zap.Int64("product_id", id))
		product.Description = req.Description
		hasChanges = true
	}

	if req.Price != nil {
		s.logger.Debug("Updating product price", zap.Int64("product_id", id), zap.Any("new_price", *req.Price))
		product.Price = *req.Price
		hasChanges = true
	}

	if req.Stock != nil {
		s.logger.Debug("Updating product stock", zap.Int64("product_id", id), zap.Int("new_stock", *req.Stock))
		product.Stock = *req.Stock
		hasChanges = true
	}

	if !hasChanges && len(req.CategoryIDs) == 0 {
		s.logger.Warn("No changes provided for product update", zap.Int64("product_id", id))
		return nil, domain_errors.ErrInvalidInput
	}

	if hasChanges {
		s.logger.Debug("Updating product in repository", zap.Int64("product_id", id))
		product.UpdatedAt = time.Now()
		err = s.ProductRepo.UpdateProduct(ctx, product)
		if err != nil {
			s.logger.Error("Failed to update product in repository", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
	}

	if len(req.CategoryIDs) > 0 {
		s.logger.Debug("Updating product categories", zap.Int64("product_id", id), zap.Any("category_ids", req.CategoryIDs))
		exists, err := s.CategoryRepo.CheckCategoriesExist(ctx, req.CategoryIDs)
		if err != nil {
			s.logger.Error("Failed to check categories exist", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
		if !exists {
			s.logger.Warn("Some categories not found", zap.Int64("product_id", id), zap.Any("category_ids", req.CategoryIDs))
			return nil, domain_errors.ErrCategoryNotFound
		}

		s.logger.Debug("Removing product from all categories", zap.Int64("product_id", id))
		err = s.ProductRepo.RemoveProductFromCategories(ctx, product.ID)
		if err != nil {
			s.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}

		s.logger.Debug("Adding product to new categories", zap.Int64("product_id", id), zap.Any("category_ids", req.CategoryIDs))
		err = s.ProductRepo.AddProductToCategories(ctx, product.ID, req.CategoryIDs)
		if err != nil {
			s.logger.Error("Failed to add product to categories", zap.Error(err), zap.Int64("product_id", id))
			return nil, err
		}
	}

	s.logger.Debug("Getting updated product categories", zap.Int64("product_id", id))
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

	s.logger.Debug("Validating product ID", zap.Int64("product_id", id))
	if err := validation.ValidateProductID(id); err != nil {
		s.logger.Error("Product ID validation failed", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	s.logger.Debug("Removing product from categories", zap.Int64("product_id", id))
	err := s.ProductRepo.RemoveProductFromCategories(ctx, id)
	if err != nil {
		s.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	s.logger.Debug("Deleting product from repository", zap.Int64("product_id", id))
	err = s.ProductRepo.DeleteProduct(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete product from repository", zap.Error(err), zap.Int64("product_id", id))
		return err
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

	s.logger.Debug("Getting products from repository", zap.Int("page", filters.Page), zap.Int("limit", filters.Limit))
	products, total, err := s.ProductRepo.GetProducts(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to get products from repository", zap.Error(err), zap.Any("filters", filters))
		return nil, err
	}

	productResponses := make([]dto.ProductCatalogItem, len(products))
	for i, product := range products {
		productResponses[i] = dto.ProductCatalogItem{
			ID:    product.ID,
			UUID:  product.UUID.String(),
			Name:  product.Name,
			Price: product.Price.StringFixed(2),
			Stock: product.Stock,
		}
	}

	s.logger.Info("Products retrieved successfully", zap.Int("products_count", len(products)), zap.Int("total", total))

	return &dto.ProductCatalogResponse{
		Products: productResponses,
		Total:    total,
		Page:     filters.Page,
		Limit:    filters.Limit,
	}, nil
}

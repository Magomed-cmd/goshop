package product

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id int64) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error)
}

type ProductHandler struct {
	productService ProductService
	logger         *zap.Logger
}

func NewProductHandler(productService ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	h.logger.Debug("GetProducts handler started")

	filters, err := h.parseProductFilters(c)
	if err != nil {
		h.logger.Error("Failed to parse product filters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters"})
		return
	}

	h.logger.Debug("Calling productService.GetProducts", zap.Any("filters", filters))
	result, err := h.productService.GetProducts(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("GetProducts service failed", zap.Error(err), zap.Any("filters", filters))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	h.logger.Debug("GetProducts successful", zap.Int("products_count", len(result.Products)))
	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	h.logger.Debug("GetProductByID handler started")

	id, err := h.parseID(c, "id")
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err), zap.String("id_param", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	h.logger.Debug("Calling productService.GetProductByID", zap.Int64("product_id", id))
	result, err := h.productService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("GetProductByID service failed", zap.Error(err), zap.Int64("product_id", id))
		if errors.Is(err, domain_errors.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
		return
	}

	h.logger.Debug("GetProductByID successful", zap.Int64("product_id", id), zap.String("product_name", result.Name))
	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	h.logger.Debug("GetProductsByCategory handler started")

	categoryID, err := h.parseID(c, "id")
	if err != nil {
		h.logger.Error("Failed to parse category ID", zap.Error(err), zap.String("id_param", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	filters, err := h.parseProductFilters(c)
	if err != nil {
		h.logger.Error("Failed to parse product filters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters"})
		return
	}

	filters.CategoryID = &categoryID

	h.logger.Debug("Calling productService.GetProducts for category", zap.Int64("category_id", categoryID), zap.Any("filters", filters))
	result, err := h.productService.GetProducts(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("GetProductsByCategory service failed", zap.Error(err), zap.Int64("category_id", categoryID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	h.logger.Debug("GetProductsByCategory successful", zap.Int64("category_id", categoryID), zap.Int("products_count", len(result.Products)))
	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	h.logger.Info("CreateProduct handler started")

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON in CreateProduct", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	h.logger.Debug("Calling productService.CreateProduct", zap.String("product_name", req.Name))
	result, err := h.productService.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("CreateProduct service failed", zap.Error(err), zap.String("product_name", req.Name))
		statusCode, message := h.mapServiceError(err)
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	h.logger.Info("CreateProduct successful", zap.Int64("product_id", result.ID), zap.String("product_name", result.Name))
	c.JSON(http.StatusCreated, result)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	h.logger.Info("UpdateProduct handler started")

	id, err := h.parseID(c, "id")
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err), zap.String("id_param", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON in UpdateProduct", zap.Error(err), zap.Int64("product_id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	h.logger.Debug("Calling productService.UpdateProduct", zap.Int64("product_id", id), zap.Any("request", req))
	result, err := h.productService.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("UpdateProduct service failed", zap.Error(err), zap.Int64("product_id", id))
		statusCode, message := h.mapServiceError(err)
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	h.logger.Info("UpdateProduct successful", zap.Int64("product_id", id), zap.String("product_name", result.Name))
	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	h.logger.Info("DeleteProduct handler started")

	id, err := h.parseID(c, "id")
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err), zap.String("id_param", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	h.logger.Debug("Calling productService.DeleteProduct", zap.Int64("product_id", id))
	err = h.productService.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("DeleteProduct service failed", zap.Error(err), zap.Int64("product_id", id))
		statusCode, message := h.mapServiceError(err)
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	h.logger.Info("DeleteProduct successful", zap.Int64("product_id", id))
	c.Status(http.StatusNoContent)
}

func (h *ProductHandler) ToggleProductStatus(c *gin.Context) {
	h.logger.Info("ToggleProductStatus handler started")

	id, err := h.parseID(c, "id")
	if err != nil {
		h.logger.Error("Failed to parse product ID", zap.Error(err), zap.String("id_param", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	h.logger.Debug("Calling productService.DeleteProduct for toggle", zap.Int64("product_id", id))
	err = h.productService.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("ToggleProductStatus service failed", zap.Error(err), zap.Int64("product_id", id))
		statusCode, message := h.mapServiceError(err)
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	h.logger.Info("ToggleProductStatus successful", zap.Int64("product_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Product status toggled"})
}

func (h *ProductHandler) parseID(c *gin.Context, param string) (int64, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

func (h *ProductHandler) parseProductFilters(c *gin.Context) (types.ProductFilters, error) {
	filters := types.ProductFilters{
		Page:  1,
		Limit: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil && categoryID > 0 {
			filters.CategoryID = &categoryID
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		switch sortBy {
		case "price", "name", "created_at":
			filters.SortBy = &sortBy
		}
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		switch sortOrder {
		case "asc", "desc":
			filters.SortOrder = &sortOrder
		}
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := decimal.NewFromString(minPriceStr); err == nil && minPrice.IsPositive() {
			filters.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := decimal.NewFromString(maxPriceStr); err == nil && maxPrice.IsPositive() {
			filters.MaxPrice = &maxPrice
		}
	}

	return filters, nil
}

func (h *ProductHandler) mapServiceError(err error) (int, string) {
	switch {
	case errors.Is(err, domain_errors.ErrInvalidInput):
		return http.StatusBadRequest, "Invalid input data"
	case errors.Is(err, domain_errors.ErrInvalidProductData):
		return http.StatusUnprocessableEntity, "Invalid product data"
	case errors.Is(err, domain_errors.ErrInvalidPrice):
		return http.StatusUnprocessableEntity, "Invalid price"
	case errors.Is(err, domain_errors.ErrInvalidStock):
		return http.StatusUnprocessableEntity, "Invalid stock value"
	case errors.Is(err, domain_errors.ErrProductNotFound):
		return http.StatusNotFound, "Product not found"
	case errors.Is(err, domain_errors.ErrCategoryNotFound):
		return http.StatusUnprocessableEntity, "Category not found"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}

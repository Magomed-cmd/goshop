package product

import (
	"context"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"goshop/internal/domain/types"
	"goshop/internal/dto"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
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

func (h *ProductHandler) GetProducts(c *gin.Context) {
	h.logger.Debug("GetProducts handler started")

	filters := types.ProductFilters{}
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
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
		if errors.Is(err, errors2.ErrProductNotFound) {
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

	filters := types.ProductFilters{}
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
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

func (h *ProductHandler) SaveProductImg(c *gin.Context) {
	h.logger.Info("SaveProductImg handler started")

	productID, ok := h.getProductID(c)
	if !ok {
		return
	}

	file, ok := h.getFormFile(c, "image")
	if !ok {
		return
	}
	defer func() {
		if err := file.Reader.Close(); err != nil {
			h.logger.Error("Failed to close file reader", zap.Error(err), zap.String("filename", file.Filename))
			c.JSON(500, gin.H{"error": "Failed to process file"})
			return
		}
	}()

	if !h.validateFileSize(c, file) {
		return
	}

	ext, ok := h.getFileExtension(c, file)
	if !ok {
		return
	}

	if !h.validateExtension(c, ext) {
		return
	}

	detectedType, ok := h.detectContentType(c, file.Reader, file.Header.Header.Get("Content-Type"))
	if !ok {
		return
	}

	ctx := c.Request.Context()
	url, err := h.productService.SaveProductImg(ctx, file.Reader, file.Size, productID, detectedType, ext[1:])
	if err != nil {
		h.logger.Error("service failed to save product image", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if url == nil {
		h.logger.Error("service returned nil URL for product image")
		c.JSON(500, gin.H{"error": "Failed to save product image"})
		return
	}

	h.logger.Info("Product image uploaded successfully",
		zap.Int64("product_id", productID),
		zap.String("url", *url))

	c.JSON(200, gin.H{
		"message":  "Product image uploaded successfully",
		"imageURL": *url,
	})
}

func (h *ProductHandler) DeleteProductImg(c *gin.Context) {
	h.logger.Info("Starting DeleteProductImg handler")

	ctx := c.Request.Context()

	productID, ok := h.getProductID(c)
	if !ok {
		h.logger.Error("Failed to get product ID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
	}

	imgID, err := h.parseID(c, "img_id")
	if err != nil {
		h.logger.Error("Failed to parse image ID", zap.Error(err), zap.String("img_id_param", c.Param("img_id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	h.logger.Debug("Calling productService.DeleteProductImg", zap.Int64("product_id", productID), zap.Int64("img_id", imgID))
	err = h.productService.DeleteProductImg(ctx, productID, imgID)
	if err != nil {
		h.logger.Error("DeleteProductImg service failed", zap.Error(err), zap.Int64("product_id", productID), zap.Int64("img_id", imgID))
		if errors.Is(err, errors2.ErrProductImageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product image"})
		return
	}
	h.logger.Info("DeleteProductImg successful",
		zap.Int64("product_id", productID),
		zap.Int64("img_id", imgID))

	c.JSON(http.StatusOK, gin.H{"message": "Product image deleted successfully"})
}

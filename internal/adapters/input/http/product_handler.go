package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	httpErrors "goshop/internal/adapters/input/http/errors"
	"goshop/internal/core/domain/types"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
)

type ProductHandler struct {
	productService serviceports.ProductService
	logger         *zap.Logger
}

func NewProductHandler(productService serviceports.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

// CreateProduct godoc
// @Summary     Create product
// @Description Creates a new product (admin only)
// @Tags        admin/products
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateProductRequest true "Product payload"
// @Success     201 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products [post]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Info("CreateProduct successful", zap.Int64("product_id", result.ID), zap.String("product_name", result.Name))
	c.JSON(http.StatusCreated, result)
}

// GetProducts godoc
// @Summary     List products
// @Description Returns products with optional filters
// @Tags        products
// @Produce     json
// @Param       page        query int    false "Page number"
// @Param       limit       query int    false "Page size"
// @Param       category_id query int    false "Category ID filter"
// @Param       sort_by     query string false "Sort field"
// @Param       sort_order  query string false "Sort order"
// @Param       min_price   query number false "Minimum price"
// @Param       max_price   query number false "Maximum price"
// @Success     200 {object} dto.ProductCatalogResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /products [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Debug("GetProducts successful", zap.Int("products_count", len(result.Products)))
	c.JSON(http.StatusOK, result)
}

// GetProductByID godoc
// @Summary     Get product
// @Description Returns product details by identifier
// @Tags        products
// @Produce     json
// @Param       id path int true "Product ID"
// @Success     200 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /products/{id} [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Debug("GetProductByID successful", zap.Int64("product_id", id), zap.String("product_name", result.Name))
	c.JSON(http.StatusOK, result)
}

// GetProductsByCategory godoc
// @Summary     Products by category
// @Description Returns products that belong to the specified category
// @Tags        products
// @Produce     json
// @Param       id path int true "Category ID"
// @Param       page        query int    false "Page number"
// @Param       limit       query int    false "Page size"
// @Param       sort_by     query string false "Sort field"
// @Param       sort_order  query string false "Sort order"
// @Param       min_price   query number false "Minimum price"
// @Param       max_price   query number false "Maximum price"
// @Success     200 {object} dto.ProductCatalogResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /products/category/{id} [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Debug("GetProductsByCategory successful", zap.Int64("category_id", categoryID), zap.Int("products_count", len(result.Products)))
	c.JSON(http.StatusOK, result)
}

// UpdateProduct godoc
// @Summary     Update product
// @Description Updates product details (admin only)
// @Tags        admin/products
// @Accept      json
// @Produce     json
// @Param       id      path int                      true "Product ID"
// @Param       request body dto.UpdateProductRequest true "Product payload"
// @Success     200 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products/{id} [put]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Info("UpdateProduct successful", zap.Int64("product_id", id), zap.String("product_name", result.Name))
	c.JSON(http.StatusOK, result)
}

// DeleteProduct godoc
// @Summary     Delete product
// @Description Deletes a product (admin only)
// @Tags        admin/products
// @Produce     json
// @Param       id path int true "Product ID"
// @Success     204 {string} string "No Content"
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products/{id} [delete]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Info("DeleteProduct successful", zap.Int64("product_id", id))
	c.Status(http.StatusNoContent)
}

// ToggleProductStatus godoc
// @Summary     Toggle product status
// @Description Toggles availability status of a product (admin only)
// @Tags        admin/products
// @Produce     json
// @Param       id path int true "Product ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products/{id}/toggle [patch]
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
		httpErrors.HandleError(c, err)
		return
	}

	h.logger.Info("ToggleProductStatus successful", zap.Int64("product_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Product status toggled"})
}

// SaveProductImg godoc
// @Summary     Upload product image
// @Description Uploads an image for a product
// @Tags        admin/products
// @Accept      multipart/form-data
// @Produce     json
// @Param       id    path     int  true "Product ID"
// @Param       image formData file true "Product image"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products/{id}/images [post]
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

// DeleteProductImg godoc
// @Summary     Delete product image
// @Description Removes an image from a product (admin only)
// @Tags        admin/products
// @Produce     json
// @Param       id     path int true "Product ID"
// @Param       img_id path int true "Image ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/products/{id}/images/{img_id} [delete]
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
		httpErrors.HandleError(c, err)
		return
	}
	h.logger.Info("DeleteProductImg successful",
		zap.Int64("product_id", productID),
		zap.Int64("img_id", imgID))

	c.JSON(http.StatusOK, gin.H{"message": "Product image deleted successfully"})
}

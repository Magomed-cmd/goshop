package httpadapter

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	errors2 "goshop/internal/core/domain/errors"
)

type uploadedFile struct {
	Filename string
	Size     int64
	Header   multipart.FileHeader
	Reader   multipart.File
}

func (h *ProductHandler) getProductID(c *gin.Context) (int64, bool) {
	productIDStr := c.Param("id")
	if productIDStr == "" {
		h.logger.Warn("missing product ID in request")
		c.JSON(400, gin.H{"error": "Product ID is required"})
		return 0, false
	}
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		h.logger.Warn("invalid product ID", zap.String("product_id", productIDStr))
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return 0, false
	}
	return productID, true
}

func (h *ProductHandler) getFormFile(c *gin.Context, field string) (*uploadedFile, bool) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		h.logger.Error("failed to get file from request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid file"})
		return nil, false
	}

	reader, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("failed to open uploaded file", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to open file"})
		return nil, false
	}

	return &uploadedFile{
		Filename: fileHeader.Filename,
		Size:     fileHeader.Size,
		Header:   *fileHeader,
		Reader:   reader,
	}, true
}

func (h *ProductHandler) validateFileSize(c *gin.Context, file *uploadedFile) bool {
	if file.Size == 0 {
		h.logger.Warn("file size is zero", zap.String("filename", file.Filename))
		c.JSON(400, gin.H{"error": "File size cannot be zero"})
		return false
	}
	if file.Size > 10*1024*1024 {
		h.logger.Warn("file size exceeds limit", zap.Int64("size", file.Size))
		c.JSON(400, gin.H{"error": "File size exceeds 10 MB limit"})
		return false
	}
	return true
}

func (h *ProductHandler) getFileExtension(c *gin.Context, file *uploadedFile) (string, bool) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		h.logger.Warn("missing file extension", zap.String("filename", file.Filename))
		c.JSON(400, gin.H{"error": "File extension is required"})
		return "", false
	}
	return ext, true
}

func (h *ProductHandler) validateExtension(c *gin.Context, ext string) bool {
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowedExt[ext] {
		h.logger.Warn("unsupported file extension", zap.String("ext", ext))
		c.JSON(400, gin.H{"error": "Unsupported file type"})
		return false
	}
	return true
}

func (h *ProductHandler) detectContentType(c *gin.Context, reader multipart.File, clientHeader string) (string, bool) {
	buf := make([]byte, 512)
	n, _ := reader.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		h.logger.Error("failed to reset file reader", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to reset file reader"})
		return "", false
	}

	h.logger.Debug("File content type checked",
		zap.String("client_header", clientHeader),
		zap.String("detected_type", detectedType),
		zap.Int("bytes_read", n),
	)

	allowedMime := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedMime[detectedType] {
		h.logger.Warn("invalid MIME type", zap.String("detected_type", detectedType))
		c.JSON(400, gin.H{"error": "Invalid image type"})
		return "", false
	}

	return detectedType, true
}

func (h *ProductHandler) mapServiceError(err error) (int, string) {
	switch {
	case errors.Is(err, errors2.ErrInvalidInput):
		return http.StatusBadRequest, "Invalid input data"
	case errors.Is(err, errors2.ErrInvalidProductData):
		return http.StatusUnprocessableEntity, "Invalid product data"
	case errors.Is(err, errors2.ErrInvalidPrice):
		return http.StatusUnprocessableEntity, "Invalid price"
	case errors.Is(err, errors2.ErrInvalidStock):
		return http.StatusUnprocessableEntity, "Invalid stock value"
	case errors.Is(err, errors2.ErrProductNotFound):
		return http.StatusNotFound, "Product not found"
	case errors.Is(err, errors2.ErrCategoryNotFound):
		return http.StatusUnprocessableEntity, "Category not found"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}

func (h *ProductHandler) parseID(c *gin.Context, param string) (int64, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

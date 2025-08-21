package user

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	errors2 "goshop/internal/domain/errors"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*entities.User, string, error)
	GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error)
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error
	UploadAvatar(ctx context.Context, reader io.ReadCloser, size, userID int64, contentType, extension string) (string, error)
	GetAvatar(ctx context.Context, userID int) (string, error)
}

type UserHandler struct {
	service UserService
	logger  *zap.Logger
}

func NewUserHandler(s UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: s,
		logger:  logger,
	}
}

func (h *UserHandler) Register(c *gin.Context) {

	var req dto.RegisterRequest
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, token, err := h.service.Register(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	roleName := ""
	if result.Role != nil {
		roleName = result.Role.Name
	}

	resp := dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			UUID:  result.UUID.String(),
			Email: result.Email,
			Name:  result.Name,
			Phone: result.Phone,
			Role:  roleName,
		},
	}

	c.JSON(201, resp)
}

func (h *UserHandler) Login(c *gin.Context) {

	var req dto.LoginRequest
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, token, err := h.service.Login(ctx, &req)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	roleName := ""
	if result.Role != nil {
		roleName = result.Role.Name
	}

	resp := dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			UUID:  result.UUID.String(),
			Email: result.Email,
			Name:  result.Name,
			Phone: result.Phone,
			Role:  roleName,
		},
	}

	c.JSON(200, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()
	profile, err := h.service.GetUserProfile(ctx, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(200, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.UpdateProfileRequest
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateProfile(ctx, userID, &req); err != nil {
		if errors.Is(err, errors2.ErrInvalidInput) {
			c.JSON(400, gin.H{"error": "No fields to update"})
			return
		}
		if strings.Contains(err.Error(), "no fields to update") {
			c.JSON(400, gin.H{"error": "No fields to update"})
			return
		}
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, gin.H{"message": "Profile updated successfully"})
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Warn("unauthorized upload attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	h.logger.Debug("User authenticated for avatar upload", zap.Int64("user_id", userID))

	file, err := c.FormFile("avatar")
	if err != nil {
		h.logger.Error("failed to get file from request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid file"})
		return
	}
	h.logger.Debug("File received", zap.String("filename", file.Filename), zap.Int64("size", file.Size))

	size := file.Size
	if size == 0 {
		h.logger.Warn("file size is zero", zap.String("filename", file.Filename))
		c.JSON(400, gin.H{"error": "File size cannot be zero"})
		return
	}

	if size > 5*1024*1024 {
		h.logger.Warn("file size exceeds limit", zap.Int64("size", size))
		c.JSON(400, gin.H{"error": "File size exceeds 5 MB limit"})
		return
	}

	if file.Filename == "" {
		h.logger.Warn("empty filename in upload")
		c.JSON(400, gin.H{"error": "File name is required"})
		return
	}

	ctx := c.Request.Context()
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		h.logger.Warn("missing file extension", zap.String("filename", file.Filename))
		c.JSON(400, gin.H{"error": "File extension is required"})
		return
	}

	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowed[ext] {
		h.logger.Warn("unsupported file extension", zap.String("ext", ext))
		c.JSON(400, gin.H{"error": "Unsupported file type"})
		return
	}

	reader, err := file.Open()
	if err != nil {
		h.logger.Error("failed to open uploaded file", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to open file"})
		return
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			h.logger.Warn("failed to close file", zap.Error(cerr))
		}
	}()

	buf := make([]byte, 512)
	n, _ := reader.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		h.logger.Error("failed to reset file reader", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to reset file reader"})
		return
	}

	h.logger.Debug("File content type checked",
		zap.String("client_header", file.Header.Get("Content-Type")),
		zap.String("detected_type", detectedType),
		zap.Int("bytes_read", n),
	)

	allowedMime := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedMime[detectedType] {
		h.logger.Warn("invalid MIME type", zap.String("detected_type", detectedType))
		c.JSON(400, gin.H{"error": "Invalid image type"})
		return
	}

	url, err := h.service.UploadAvatar(ctx, reader, size, userID, detectedType, ext)
	if err != nil {
		h.logger.Error("service failed to upload avatar", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	} else if url == "" {
		h.logger.Error("service returned empty URL for avatar")
		c.JSON(500, gin.H{"error": "Failed to upload avatar"})
		return
	}

	h.logger.Info("Avatar uploaded successfully", zap.Int64("user_id", userID), zap.String("url", url))
	c.JSON(200, gin.H{"message": "Avatar uploaded successfully", "imageURL": url})
}

func (h *UserHandler) GetAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	userIDAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	url, err := h.service.GetAvatar(ctx, int(userID))
	if err != nil {
		h.logger.Error("failed to get avatar", zap.Error(err), zap.Int64("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}

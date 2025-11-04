package httpadapter

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	httpErrors "goshop/internal/adapters/input/http/errors"
	"goshop/internal/core/mappers"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type UserHandler struct {
	service serviceports.UserService
	logger  *zap.Logger
}

func NewUserHandler(s serviceports.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: s,
		logger:  logger,
	}
}

// Register godoc
// @Summary     Register user
// @Description Registers a new user and returns JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dto.RegisterRequest true "Registration payload"
// @Success     201 {object} dto.AuthResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {

	var req dto.RegisterRequest
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, token, err := h.service.Register(ctx, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	roleName := ""
	if result.Role != nil {
		roleName = result.Role.Name
	}

	profile := mappers.ToUserProfile(result, roleName)
	resp := dto.AuthResponse{
		Token: token,
		User:  profile,
	}

	c.JSON(201, resp)
}

// Login godoc
// @Summary     Login user
// @Description Authenticates user credentials and returns JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dto.LoginRequest true "Login payload"
// @Success     200 {object} dto.AuthResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {

	var req dto.LoginRequest
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, token, err := h.service.Login(ctx, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	roleName := ""
	if result.Role != nil {
		roleName = result.Role.Name
	}

	profile := mappers.ToUserProfile(result, roleName)
	resp := dto.AuthResponse{
		Token: token,
		User:  profile,
	}

	c.JSON(200, resp)
}

// GetProfile godoc
// @Summary     Get profile
// @Description Returns the current user's profile
// @Tags        profile
// @Produce     json
// @Success     200 {object} dto.UserProfile
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()
	profile, err := h.service.GetUserProfile(ctx, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}
	c.JSON(200, profile)
}

// UpdateProfile godoc
// @Summary     Update profile
// @Description Updates the authenticated user's profile
// @Tags        profile
// @Accept      json
// @Produce     json
// @Param       request body dto.UpdateProfileRequest true "Profile payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/profile [put]
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
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "Profile updated successfully"})
}

// UploadAvatar godoc
// @Summary     Upload avatar
// @Description Uploads or replaces the authenticated user's avatar
// @Tags        profile
// @Accept      multipart/form-data
// @Produce     json
// @Param       avatar formData file true "Avatar image"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/profile/avatar [put]
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

// GetAvatar godoc
// @Summary     Get avatar
// @Description Returns a signed URL to the authenticated user's avatar
// @Tags        profile
// @Produce     json
// @Success     200 {object} map[string]string
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/avatar [get]
func (h *UserHandler) GetAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	url, err := h.service.GetAvatar(ctx, int(userID))
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}

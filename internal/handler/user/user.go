package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"strings"
)

type UserServiceInterface interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*entities.User, string, error)
	GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error)
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error
}

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(s UserServiceInterface) *UserHandler {
	return &UserHandler{
		service: s,
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
		if strings.Contains(err.Error(), "no fields to update") {
			c.JSON(400, gin.H{"error": "No fields to update"})
			return
		}
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, gin.H{"message": "Profile updated successfully"})
}

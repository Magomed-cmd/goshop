package handler

import (
	"github.com/gin-gonic/gin"
	"goshop/internal/dto"
	"goshop/internal/service/auth"
	"strings"
)

type AuthHandler struct {
	service *auth.AuthService
}

func NewAuthHandler(s *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		service: s,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {

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

func (h *AuthHandler) Login(c *gin.Context) {

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

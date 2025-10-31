package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/oauth/google"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Интерфейс для OAuth провайдеров
type OAuthProvider interface {
	GetAuthURL(state string) string
	GetUserInfo(ctx context.Context, code string) (*google.UserInfo, error)
}

// Сервис для работы с пользователями через OAuth
type AuthService interface {
	OAuthLogin(ctx context.Context, userInfo *google.UserInfo) (*entities.User, string, error)
}

type OAuthHandler struct {
	googleProvider OAuthProvider
	authService    AuthService
	logger         *zap.Logger
	states         map[string]bool // в продакшене лучше Redis
}

func NewOAuthHandler(googleProvider OAuthProvider, authService AuthService, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		googleProvider: googleProvider,
		authService:    authService,
		logger:         logger,
		states:         make(map[string]bool),
	}
}

// GET /auth/google/login
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	state := h.generateState()
	h.states[state] = true // сохраняем state для проверки

	url := h.googleProvider.GetAuthURL(state)
	h.logger.Info("redirecting to google oauth", zap.String("state", state))

	c.Redirect(302, url)
}

// GET /auth/google/callback
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	ctx := c.Request.Context()

	// Проверяем state
	state := c.Query("state")
	if !h.states[state] {
		h.logger.Warn("invalid oauth state", zap.String("state", state))
		c.JSON(400, gin.H{"error": "Invalid state"})
		return
	}
	delete(h.states, state) // удаляем использованный state

	// Получаем код авторизации
	code := c.Query("code")
	if code == "" {
		h.logger.Warn("missing authorization code")
		c.JSON(400, gin.H{"error": "Missing authorization code"})
		return
	}

	// Получаем информацию о пользователе от Google
	userInfo, err := h.googleProvider.GetUserInfo(ctx, code)
	if err != nil {
		h.logger.Error("failed to get user info from google", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to get user info"})
		return
	}

	// Логиним или создаем пользователя
	user, token, err := h.authService.OAuthLogin(ctx, userInfo)
	if err != nil {
		h.logger.Error("failed to oauth login", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Формируем ответ как в обычном логине
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	resp := dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			UUID:  user.UUID.String(),
			Email: user.Email,
			Name:  user.Name,
			Phone: user.Phone,
			Role:  roleName,
		},
	}

	h.logger.Info("oauth login successful", zap.String("email", user.Email), zap.String("provider", "google"))
	c.JSON(200, resp)
}

func (h *OAuthHandler) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

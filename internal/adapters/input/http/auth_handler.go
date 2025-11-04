package httpadapter

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"goshop/internal/core/mappers"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
	"goshop/internal/oauth/google"
	"goshop/internal/utils"
)

type OAuthProvider interface {
	GetAuthURL(state string) string
	GetUserInfo(ctx context.Context, code string) (*google.UserInfo, error)
}

type OAuthHandler struct {
	googleProvider OAuthProvider
	authService    serviceports.UserService
	logger         *zap.Logger
	redis          *redis.Client
}

func NewOAuthHandler(googleProvider OAuthProvider, authService serviceports.UserService, redis *redis.Client, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		googleProvider: googleProvider,
		authService:    authService,
		logger:         logger,
		redis:          redis,
	}
}

// GoogleLogin godoc
// @Summary     Google OAuth login
// @Description Redirects the user to the Google OAuth consent screen
// @Tags        auth
// @Produce     json
// @Success     302 {string} string "Redirect to Google OAuth"
// @Failure     500 {object} map[string]string
// @Router      /auth/google/login [get]
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	ctx := c.Request.Context()

	state := utils.GenerateState()

	err := h.redis.Set(ctx, "oauth_state:"+state, "1", 10*time.Minute).Err()
	if err != nil {
		h.logger.Error("failed to save oauth state to redis", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	url := h.googleProvider.GetAuthURL(state)
	h.logger.Info("redirecting to google oauth", zap.String("state", state))

	c.Redirect(302, url)
}

// GoogleCallback godoc
// @Summary     Google OAuth callback
// @Description Handles Google OAuth callback, exchanges code for user info and returns JWT
// @Tags        auth
// @Produce     json
// @Param       state query string true "OAuth state"
// @Param       code  query string true "Authorization code"
// @Success     200 {object} dto.AuthResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/google/callback [get]
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	ctx := c.Request.Context()

	state := c.Query("state")

	exists := h.redis.GetDel(ctx, "oauth_state:"+state).Val()
	if exists != "1" {
		h.logger.Warn("invalid oauth state", zap.String("state", state))
		c.JSON(400, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		h.logger.Warn("missing authorization code")
		c.JSON(400, gin.H{"error": "Missing authorization code"})
		return
	}

	userInfo, err := h.googleProvider.GetUserInfo(ctx, code)
	if err != nil {
		h.logger.Error("failed to get user info from google", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to get user info"})
		return
	}

	user, token, err := h.authService.OAuthLogin(ctx, userInfo)
	if err != nil {
		h.logger.Error("failed to oauth login", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	resp := dto.AuthResponse{
		Token: token,
		User:  mappers.ToUserProfile(user, roleName),
	}

	h.logger.Info("oauth login successful", zap.String("email", user.Email), zap.String("provider", "google"))
	c.JSON(200, resp)
}

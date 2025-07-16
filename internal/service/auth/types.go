package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"goshop/internal/models"
)

const (
	defaultUserRole = "user"
	bcryptCost      = 14
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type AuthService struct {
	roleRepo     RoleRepository
	userRepo     UserRepository
	jwtSecretKey string
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

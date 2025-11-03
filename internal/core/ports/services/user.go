package services

import (
	"context"
	"io"

	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/oauth/google"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*entities.User, string, error)
	GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error)
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error
	UploadAvatar(ctx context.Context, reader io.ReadCloser, size, userID int64, contentType, extension string) (string, error)
	GetAvatar(ctx context.Context, userID int) (string, error)
	OAuthLogin(ctx context.Context, userInfo *google.UserInfo) (*entities.User, string, error)
}

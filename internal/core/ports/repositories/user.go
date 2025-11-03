package repositories

import (
	"context"

	"goshop/internal/core/domain/entities"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	GetAvatar(ctx context.Context, userID int) (*entities.UserAvatar, error)
	SaveAvatar(ctx context.Context, userAvatarInfo *entities.UserAvatar) (int, error)
	UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error
	DeleteAvatar(ctx context.Context, userID int) error
}

package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	domainerrors "goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/vo"
	"goshop/internal/core/ports/repositories"
	storageports "goshop/internal/core/ports/storage"
	dtx "goshop/internal/core/ports/transaction"
	"goshop/internal/oauth/google"
	"goshop/internal/utils"
)

const (
	defaultUserRole = "user"
	defaultImgURL   = "https://storage.yandexcloud.net/goshop/avatars/default/default-avatar.jpg"
)

type UserService struct {
	roleRepo     repositories.RoleRepository
	userRepo     repositories.UserRepository
	uow          dtx.UnitOfWork
	jwtSecretKey string
	bcryptCost   int
	ImgStorage   storageports.ImgStorage
	logger       *zap.Logger
}

func NewUserService(
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	uow dtx.UnitOfWork,
	jwtSecret string,
	bcryptCost int,
	imgStorage storageports.ImgStorage,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		uow:          uow,
		jwtSecretKey: jwtSecret,
		bcryptCost:   bcryptCost,
		ImgStorage:   imgStorage,
		logger:       logger,
	}
}

func (s *UserService) Register(ctx context.Context, email, password string, name, phone *string) (*entities.User, string, error) {
	emailVO, err := vo.NewEmail(email)
	if err != nil {
		return nil, "", err
	}

	rawPassword, err := vo.NewRawPassword(password)
	if err != nil {
		return nil, "", domainerrors.ErrInvalidInput
	}

	var user *entities.User
	var role *entities.Role
	err = s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		if err := s.checkUserExistsWithRepo(ctx, repos.Users(), emailVO.String()); err != nil {
			return err
		}

		role, err = repos.Roles().GetByName(ctx, defaultUserRole)
		if err != nil {
			return err
		}

		user, err = s.createUserWithRepo(ctx, repos.Users(), emailVO, rawPassword, role.ID, name, phone)
		if err != nil {
			return err
		}

		now := time.Now()
		_, err = repos.Users().SaveAvatar(ctx, &entities.UserAvatar{
			ID:        0,
			UserID:    user.ID,
			ImageURL:  defaultImgURL,
			UUID:      uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return domainerrors.ErrAvatarUploadFail
		}

		return nil
	})
	if err != nil {
		s.logger.Error("User registration failed", zap.Error(err), zap.String("email", email))
		return nil, "", err
	}

	user.AttachRole(role)

	token, err := utils.GenerateJWT(user.ID, user.Email, role.Name, s.jwtSecretKey)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*entities.User, string, error) {
	emailVO, err := vo.NewEmail(email)
	if err != nil {
		return nil, "", err
	}
	rawPassword, err := vo.NewRawPassword(password)
	if err != nil {
		return nil, "", domainerrors.ErrInvalidPassword
	}

	var existingUser *entities.User
	err = s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		existingUser, innerErr = repos.Users().GetUserByEmail(ctx, emailVO.String())
		if innerErr != nil {
			return innerErr
		}

		if existingUser.Role == nil && existingUser.RoleID != nil {
			existingUser.Role, innerErr = repos.Roles().GetByID(ctx, *existingUser.RoleID)
			if innerErr != nil {
				return innerErr
			}
		}

		return nil
	})
	if err != nil {
		return nil, "", err
	}

	if err = existingUser.VerifyPassword(rawPassword); err != nil {
		return nil, "", err
	}

	roleName := ""
	if existingUser.Role != nil {
		roleName = existingUser.Role.Name
	}

	token, err := utils.GenerateJWT(existingUser.ID, existingUser.Email, roleName, s.jwtSecretKey)
	if err != nil {
		return nil, "", err
	}

	return existingUser, token, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*entities.User, error) {
	var user *entities.User
	var userRole *entities.Role
	err := s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		user, innerErr = repos.Users().GetUserByID(ctx, userID)
		if innerErr != nil {
			return innerErr
		}

		userRole, innerErr = repos.Roles().GetByID(ctx, *user.RoleID)
		return innerErr
	})
	if err != nil {
		return nil, err
	}

	user.AttachRole(userRole)
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, name, phone *string) error {
	return s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		user, err := repos.Users().GetUserByID(ctx, userID)
		if err != nil {
			return err
		}

		if err = user.ApplyProfilePatch(name, phone); err != nil {
			return err
		}

		return repos.Users().UpdateUserProfile(ctx, userID, user.Name, user.Phone)
	})
}

func (s *UserService) CheckUserExists(ctx context.Context, email string) error {
	return s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		return s.checkUserExistsWithRepo(ctx, repos.Users(), email)
	})
}

func (s *UserService) CreateUser(ctx context.Context, email, password string, roleID int64, name, phone *string) (*entities.User, error) {
	emailVO, err := vo.NewEmail(email)
	if err != nil {
		return nil, err
	}
	rawPassword, err := vo.NewRawPassword(password)
	if err != nil {
		return nil, domainerrors.ErrInvalidInput
	}

	var user *entities.User
	err = s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		user, innerErr = s.createUserWithRepo(ctx, repos.Users(), emailVO, rawPassword, roleID, name, phone)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UploadAvatar(ctx context.Context, reader io.ReadCloser, size, userID int64, contentType, extension string) (string, error) {
	s.logger.Debug("Start UploadAvatar",
		zap.Int64("user_id", userID),
		zap.Int64("size", size),
		zap.String("content_type", contentType),
		zap.String("extension", extension),
	)

	if userID < 1 {
		s.logger.Error("Invalid user ID", zap.Int64("user_id", userID))
		return "", domainerrors.ErrInvalidUserID
	}

	userAvatarInfo := &entities.UserAvatar{
		UserID:    userID,
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	objectName := fmt.Sprintf("avatars/%d/avatar%s", userID, extension)
	s.logger.Debug("Prepared object name for storage", zap.String("object_name", objectName))

	url, err := s.ImgStorage.UploadImage(ctx, objectName, reader, size, contentType)
	if err != nil {
		s.logger.Error("Failed to upload image to storage", zap.Error(err))
		return "", err
	}

	userAvatarInfo.ImageURL = *url
	s.logger.Info("Image uploaded successfully", zap.String("image_url", userAvatarInfo.ImageURL))

	s.logger.Debug("Saving avatar info to database", zap.Int64("user_id", userID))
	var id int
	err = s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		id, innerErr = repos.Users().SaveAvatar(ctx, userAvatarInfo)
		return innerErr
	})
	if err != nil {
		s.logger.Error("Failed to save avatar info to database", zap.Error(err))
		return "", err
	}
	s.logger.Info("Avatar info saved to database", zap.Int64("avatar_id", int64(id)))

	userAvatarInfo.ID = int64(id)
	return userAvatarInfo.ImageURL, nil
}

func (s *UserService) GetAvatar(ctx context.Context, userID int) (string, error) {
	s.logger.Debug("Getting avatar", zap.Int("user_id", userID))

	var avatar *entities.UserAvatar
	err := s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		avatar, innerErr = repos.Users().GetAvatar(ctx, userID)
		return innerErr
	})
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			s.logger.Info("No avatar found, using default", zap.Int("user_id", userID))
			return s.ImgStorage.GetImageURL("avatars/default/default_avatar.jpg"), nil
		}
		s.logger.Error("Failed to get avatar from repo", zap.Error(err), zap.Int("user_id", userID))
		return "", err
	}

	s.logger.Info("Avatar retrieved", zap.Int("user_id", userID), zap.String("url", avatar.ImageURL))
	return avatar.ImageURL, nil
}

func (s *UserService) OAuthLogin(ctx context.Context, userInfo *google.UserInfo) (*entities.User, string, error) {
	if userInfo == nil {
		return nil, "", domainerrors.ErrInvalidInput
	}

	emailVO, err := vo.NewEmail(userInfo.Email)
	if err != nil {
		return nil, "", err
	}

	var existingUser *entities.User
	err = s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		existingUser, innerErr = repos.Users().GetUserByEmail(ctx, emailVO.String())
		return innerErr
	})
	if err != nil {
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			s.logger.Error("database error during user lookup", zap.Error(err))
			return nil, "", err
		}

		newUser, createErr := s.createOAuthUser(ctx, emailVO, userInfo)
		if createErr != nil {
			return nil, "", createErr
		}

		token, tokenErr := s.generateTokenForUser(newUser)
		if tokenErr != nil {
			return nil, "", tokenErr
		}
		return newUser, token, nil
	}

	token, err := s.generateTokenForUser(existingUser)
	if err != nil {
		return nil, "", err
	}

	return existingUser, token, nil
}

func (s *UserService) generateTokenForUser(user *entities.User) (string, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, roleName, s.jwtSecretKey)
	if err != nil {
		s.logger.Error("failed to generate JWT token", zap.Error(err), zap.Int64("user_id", user.ID))
		return "", domainerrors.ErrInvalidInput
	}
	return token, nil
}

func (s *UserService) createOAuthUser(ctx context.Context, emailVO vo.Email, userInfo *google.UserInfo) (*entities.User, error) {
	var createdUser *entities.User
	var role *entities.Role

	err := s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		var err error
		role, err = repos.Roles().GetByName(ctx, defaultUserRole)
		if err != nil {
			return err
		}

		createdUser, err = entities.NewOAuthUser(emailVO, role.ID, &userInfo.Name)
		if err != nil {
			return err
		}

		if err = repos.Users().CreateUser(ctx, createdUser); err != nil {
			return err
		}

		now := time.Now()
		_, err = repos.Users().SaveAvatar(ctx, &entities.UserAvatar{
			UserID:    createdUser.ID,
			ImageURL:  defaultImgURL,
			UUID:      uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return domainerrors.ErrAvatarUploadFail
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to create oauth user", zap.Error(err))
		if errors.Is(err, domainerrors.ErrEmailExists) {
			return nil, domainerrors.ErrEmailExists
		}
		return nil, err
	}

	createdUser.AttachRole(role)
	return createdUser, nil
}

func (s *UserService) checkUserExistsWithRepo(ctx context.Context, userRepo repositories.UserRepository, email string) error {
	existingUser, err := userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil
		}
		return err
	}
	if existingUser != nil {
		return domainerrors.ErrEmailExists
	}
	return nil
}

func (s *UserService) createUserWithRepo(
	ctx context.Context,
	userRepo repositories.UserRepository,
	emailVO vo.Email,
	rawPassword vo.RawPassword,
	roleID int64,
	name, phone *string,
) (*entities.User, error) {
	hashedPassword, err := utils.HashPasswordWithCost(rawPassword.String(), s.bcryptCost)
	if err != nil {
		return nil, err
	}

	user, err := entities.NewUserForRegistration(emailVO, hashedPassword, roleID, name, phone, time.Now())
	if err != nil {
		return nil, err
	}

	if err = userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) withinWriteUOW(
	ctx context.Context,
	fn func(repos dtx.Repositories) error,
) error {
	if s.uow == nil {
		return fn(&fallbackRepos{users: s.userRepo, roles: s.roleRepo})
	}
	return s.uow.Do(ctx, fn)
}

func (s *UserService) withinReadUOW(
	ctx context.Context,
	fn func(repos dtx.Repositories) error,
) error {
	if s.uow == nil {
		return fn(&fallbackRepos{users: s.userRepo, roles: s.roleRepo})
	}
	return s.uow.DoRead(ctx, fn)
}

type fallbackRepos struct {
	users repositories.UserRepository
	roles repositories.RoleRepository
}

func (f *fallbackRepos) Users() repositories.UserRepository           { return f.users }
func (f *fallbackRepos) Roles() repositories.RoleRepository           { return f.roles }
func (f *fallbackRepos) Addresses() repositories.AddressRepository    { return nil }
func (f *fallbackRepos) Categories() repositories.CategoryRepository  { return nil }
func (f *fallbackRepos) Products() repositories.ProductRepository     { return nil }
func (f *fallbackRepos) Carts() repositories.CartRepository           { return nil }
func (f *fallbackRepos) Orders() repositories.OrderRepository         { return nil }
func (f *fallbackRepos) OrderItems() repositories.OrderItemRepository { return nil }
func (f *fallbackRepos) Reviews() repositories.ReviewRepository       { return nil }

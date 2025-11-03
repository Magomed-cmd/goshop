package services

import (
    "context"
    "errors"
    "fmt"
    "io"
    "strings"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"

    "goshop/internal/core/domain/entities"
    errors2 "goshop/internal/core/domain/errors"
    "goshop/internal/core/ports/repositories"
    storageports "goshop/internal/core/ports/storage"
    "goshop/internal/dto"
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
    jwtSecretKey string
    bcryptCost   int
    ImgStorage   storageports.ImgStorage
    logger       *zap.Logger
}

func NewUserService(roleRepo repositories.RoleRepository, userRepo repositories.UserRepository, jwtSecret string, bcryptCost int, imgStorage storageports.ImgStorage, logger *zap.Logger) *UserService {
	return &UserService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		jwtSecretKey: jwtSecret,
		bcryptCost:   bcryptCost,
		ImgStorage:   imgStorage,
		logger:       logger,
	}
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error) {
	s.logger.Info("UserService Register started", zap.String("email", req.Email))

	s.logger.Debug("Checking if user exists", zap.String("email", req.Email))
	if err := s.CheckUserExists(ctx, req.Email); err != nil {
		s.logger.Error("User exists check failed", zap.Error(err), zap.String("email", req.Email))
		return nil, "", err
	}

	s.logger.Debug("Getting user role", zap.String("role", defaultUserRole))
	role, err := s.roleRepo.GetByName(ctx, defaultUserRole)
	if err != nil {
		s.logger.Error("Failed to get user role", zap.Error(err), zap.String("role", defaultUserRole))
		return nil, "", err
	}

	s.logger.Debug("Creating user", zap.String("email", req.Email), zap.Int64("role_id", role.ID))
	user, err := s.CreateUser(ctx, req, role.ID)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, "", err
	}

	_, err = s.userRepo.SaveAvatar(ctx, &entities.UserAvatar{
		ID:        0,
		UserID:    user.ID,
		ImageURL:  defaultImgURL,
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	user.Role = role

	s.logger.Debug("Generating JWT token", zap.Int64("user_id", user.ID))
	token, err := utils.GenerateJWT(user.ID, user.Email, role.Name, s.jwtSecretKey)
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err), zap.Int64("user_id", user.ID))
		return nil, "", err
	}

	s.logger.Info("User registered successfully", zap.Int64("user_id", user.ID), zap.String("email", req.Email))
	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*entities.User, string, error) {
	s.logger.Info("UserService Login started", zap.String("email", req.Email))

	s.logger.Debug("Getting user by email", zap.String("email", req.Email))
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", req.Email))
		return nil, "", err
	}

	s.logger.Debug("Validating password", zap.String("email", req.Email))
	err = utils.ValidatePassword(existingUser.PasswordHash, req.Password)
	if err != nil {
		s.logger.Warn("Invalid password provided", zap.String("email", req.Email))
		return nil, "", errors2.ErrInvalidPassword
	}

	if existingUser.Role == nil {
		s.logger.Debug("Loading user role", zap.Int64("role_id", *existingUser.RoleID))
		role, err := s.roleRepo.GetByID(ctx, *existingUser.RoleID)
		if err != nil {
			s.logger.Error("Failed to get role by ID", zap.Error(err), zap.Int64("role_id", *existingUser.RoleID))
			return nil, "", err
		}
		existingUser.Role = role
	}

	s.logger.Debug("Generating JWT token", zap.Int64("user_id", existingUser.ID))
	token, err := utils.GenerateJWT(existingUser.ID, existingUser.Email, existingUser.Role.Name, s.jwtSecretKey)
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err), zap.Int64("user_id", existingUser.ID))
		return nil, "", err
	}

	s.logger.Info("User logged in successfully", zap.Int64("user_id", existingUser.ID), zap.String("email", req.Email))
	return existingUser, token, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error) {
	s.logger.Debug("Getting user profile", zap.Int64("user_id", userID))

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	s.logger.Debug("Getting user role", zap.Int64("user_id", userID), zap.Int64("role_id", *user.RoleID))
	userRole, err := s.roleRepo.GetByID(ctx, *user.RoleID)
	if err != nil {
		s.logger.Error("Failed to get role by ID", zap.Error(err), zap.Int64("role_id", *user.RoleID))
		return nil, err
	}

	userResponse := &dto.UserProfile{
		UUID:  user.UUID.String(),
		Email: user.Email,
		Name:  user.Name,
		Phone: user.Phone,
		Role:  userRole.Name,
	}

	s.logger.Debug("User profile retrieved successfully", zap.Int64("user_id", userID), zap.String("role", userRole.Name))
	return userResponse, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error {
	s.logger.Debug("Updating user profile", zap.Int64("user_id", userID), zap.Any("request", req))

	if req.Name == nil && req.Phone == nil {
		s.logger.Warn("No fields provided for update", zap.Int64("user_id", userID))
		return errors2.ErrInvalidInput
	}

	s.logger.Debug("Calling repository to update profile", zap.Int64("user_id", userID))
	err := s.userRepo.UpdateUserProfile(ctx, userID, req.Name, req.Phone)
	if err != nil {
		s.logger.Error("Failed to update user profile", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	s.logger.Info("User profile updated successfully", zap.Int64("user_id", userID))
	return nil
}

func (s *UserService) CheckUserExists(ctx context.Context, email string) error {
	s.logger.Debug("Checking if user exists", zap.String("email", email))

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errors2.ErrUserNotFound) {
			s.logger.Debug("User does not exist - OK to register", zap.String("email", email))
			return nil
		}
		s.logger.Error("Error checking user existence", zap.Error(err), zap.String("email", email))
		return err
	}

	if existingUser != nil {
		s.logger.Warn("User already exists", zap.String("email", email))
		return errors2.ErrEmailExists
	}

	s.logger.Debug("User does not exist - OK to register", zap.String("email", email))
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.RegisterRequest, roleID int64) (*entities.User, error) {
	s.logger.Debug("Creating new user", zap.String("email", req.Email), zap.Int64("role_id", roleID))

	s.logger.Debug("Generating UUID")
	uuidV1, err := uuid.NewUUID()
	if err != nil {
		s.logger.Error("Failed to generate UUID", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Hashing password", zap.Int("bcrypt_cost", s.bcryptCost))
	hashedPassword, err := utils.HashPasswordWithCost(req.Password, s.bcryptCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := &entities.User{
		UUID:         uuidV1,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		Phone:        req.Phone,
		RoleID:       &roleID,
		CreatedAt:    time.Now(),
	}

	s.logger.Debug("Saving user to database", zap.String("email", req.Email))
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to save user to database", zap.Error(err), zap.String("email", req.Email))
		return nil, err
	}

	s.logger.Info("User created successfully", zap.Int64("user_id", user.ID), zap.String("email", req.Email))
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
		return "", errors2.ErrInvalidUserID
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
	id, err := s.userRepo.SaveAvatar(ctx, userAvatarInfo)

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

	avatar, err := s.userRepo.GetAvatar(ctx, userID)
	if err != nil {
		if errors.Is(err, errors2.ErrNotFound) {
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
	existingUser, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		if !errors.Is(err, errors2.ErrUserNotFound) {
			s.logger.Error("database error during user lookup", zap.Error(err))
			return nil, "", err
		}

		newUser, err := s.createOAuthUser(ctx, userInfo)
		if err != nil {
			return nil, "", err
		}

		token, err := s.generateTokenForUser(newUser)
		if err != nil {
			return nil, "", err
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
		return "", errors2.ErrInvalidInput
	}
	return token, nil
}

func (s *UserService) createOAuthUser(ctx context.Context, userInfo *google.UserInfo) (*entities.User, error) {

	role, err := s.roleRepo.GetByName(ctx, defaultUserRole)
	if err != nil {
		s.logger.Error("failed to get default role", zap.Error(err))
		return nil, err
	}

	name := userInfo.Name
	newUser := &entities.User{
		UUID:         uuid.New(),
		Email:        userInfo.Email,
		Name:         &name,
		PasswordHash: "",
		Phone:        nil,
		RoleID:       &role.ID,
	}

	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		s.logger.Error("failed to create oauth user", zap.Error(err))
		if strings.Contains(err.Error(), "already exists") {
			return nil, errors2.ErrEmailExists
		}
		return nil, errors2.ErrInvalidInput
	}

	_, err = s.userRepo.SaveAvatar(ctx, &entities.UserAvatar{
		UserID:    newUser.ID,
		ImageURL:  defaultImgURL,
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		s.logger.Warn("failed to create default avatar for oauth user", zap.Error(err))
	}

	newUser.Role = role
	return newUser, nil
}

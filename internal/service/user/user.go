package user

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	errors2 "goshop/internal/domain/errors"
	"goshop/internal/dto"
	"goshop/internal/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultUserRole = "user"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.Role, error)
	GetByName(ctx context.Context, name string) (*entities.Role, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error
}

type UserService struct {
	roleRepo     RoleRepository
	userRepo     UserRepository
	jwtSecretKey string
	bcryptCost   int
	logger       *zap.Logger
}

func NewUserService(roleRepo RoleRepository, userRepo UserRepository, jwtSecret string, bcryptCost int, logger *zap.Logger) *UserService {
	return &UserService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		jwtSecretKey: jwtSecret,
		bcryptCost:   bcryptCost,
		logger:       logger,
	}
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error) {
	s.logger.Info("UserService Register started", zap.String("email", req.Email))

	s.logger.Debug("Checking if user exists", zap.String("email", req.Email))
	if err := s.checkUserExists(ctx, req.Email); err != nil {
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
	user, err := s.createUser(ctx, req, role.ID)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, "", err
	}

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

func (s *UserService) checkUserExists(ctx context.Context, email string) error {
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

func (s *UserService) createUser(ctx context.Context, req *dto.RegisterRequest, roleID int64) (*entities.User, error) {
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

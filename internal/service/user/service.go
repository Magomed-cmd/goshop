package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"goshop/internal/dto"
	"goshop/internal/models"
	"goshop/internal/utils"
	"time"
)

const (
	defaultUserRole = "user"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error
}

type UserService struct {
	roleRepo     RoleRepository
	userRepo     UserRepository
	jwtSecretKey string
}

func NewUserService(roleRepo RoleRepository, userRepo UserRepository, jwtSecret string) *UserService {
	return &UserService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		jwtSecretKey: jwtSecret,
	}
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, string, error) {
	if err := s.checkUserExists(ctx, req.Email); err != nil {
		return nil, "", err
	}

	role, err := s.roleRepo.GetByName(ctx, defaultUserRole)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get role by name")
		return nil, "", err
	}

	user, err := s.createUser(ctx, req, role.ID)
	if err != nil {
		return nil, "", err
	}

	user.Role = role

	token, err := utils.GenerateJWT(user.ID, user.Email, role.Name, s.jwtSecretKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*models.User, string, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Str("email", req.Email).Msg("User not found")
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		return nil, "", err
	}

	err = utils.ValidatePassword(existingUser.PasswordHash, req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Invalid password")
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if existingUser.Role == nil {
		role, err := s.roleRepo.GetByID(ctx, *existingUser.RoleID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user role")
			return nil, "", err
		}
		existingUser.Role = role
	}

	token, err := utils.GenerateJWT(existingUser.ID, existingUser.Email, existingUser.Role.Name, s.jwtSecretKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, "", err
	}

	return existingUser, token, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by ID")
		return nil, err
	}

	roleName := ""
	userRole, err := s.roleRepo.GetByID(ctx, *user.RoleID)
	if userRole == nil {
		log.Error().Err(err).Msg("User role not found")
		return nil, fmt.Errorf("user role not found")
	}

	roleName = userRole.Name

	if err != nil {
		log.Error().Err(err).Msg("Failed to get user role by ID")
		return nil, err
	}

	userResponse := &dto.UserProfile{
		UUID:  user.UUID.String(),
		Email: user.Email,
		Name:  user.Name,
		Phone: user.Phone,
		Role:  roleName,
	}

	return userResponse, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error {
	if err := s.userRepo.UpdateUserProfile(ctx, userID, req.Name, req.Phone); err != nil {
		log.Error().Err(err).Msg("Failed to update user profile")
		return err
	}

	return nil
}

func (s *UserService) checkUserExists(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		log.Error().Str("email", email).Msg("User already exists with this email")
		return fmt.Errorf("user with this email already exists")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Failed to check existing user")
		return err
	}

	return nil
}

func (s *UserService) createUser(ctx context.Context, req *dto.RegisterRequest, roleID int64) (*models.User, error) {
	uuidV1, err := uuid.NewUUID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate UUID")
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, err
	}

	user := &models.User{
		UUID:         uuidV1,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		Phone:        req.Phone,
		RoleID:       &roleID,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	return user, nil
}

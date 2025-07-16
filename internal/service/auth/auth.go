package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"goshop/internal/dto"
	"goshop/internal/models"
)

func NewAuthService(roleRepo RoleRepository, userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		jwtSecretKey: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, string, error) {
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

	token, err := s.generateTokenJWT(user.ID, user.Email, role.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*models.User, string, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Str("email", req.Email).Msg("User not found")
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		return nil, "", err
	}

	err = ValidatePassword(existingUser.PasswordHash, req.Password)
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

	token, err := s.generateTokenJWT(existingUser.ID, existingUser.Email, existingUser.Role.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, "", err
	}

	return existingUser, token, nil
}

package user

import (
	"context"
	"github.com/google/uuid"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"goshop/internal/utils"
	"time"
)

const (
	defaultUserRole = "user"
)

type RoleRepositoryInterface interface {
	GetByID(ctx context.Context, id int64) (*entities.Role, error)
	GetByName(ctx context.Context, name string) (*entities.Role, error)
}

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error
}

type UserService struct {
	roleRepo     RoleRepositoryInterface
	userRepo     UserRepositoryInterface
	jwtSecretKey string
}

func NewUserService(roleRepo RoleRepositoryInterface, userRepo UserRepositoryInterface, jwtSecret string) *UserService {
	return &UserService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		jwtSecretKey: jwtSecret,
	}
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*entities.User, string, error) {
	if err := s.checkUserExists(ctx, req.Email); err != nil {
		return nil, "", err
	}

	role, err := s.roleRepo.GetByName(ctx, defaultUserRole)
	if err != nil {
		return nil, "", err
	}

	user, err := s.createUser(ctx, req, role.ID)
	if err != nil {
		return nil, "", err
	}

	user.Role = role

	token, err := utils.GenerateJWT(user.ID, user.Email, role.Name, s.jwtSecretKey)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*entities.User, string, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}

	err = utils.ValidatePassword(existingUser.PasswordHash, req.Password)
	if err != nil {
		return nil, "", domain_errors.ErrInvalidPassword
	}

	if existingUser.Role == nil {
		role, err := s.roleRepo.GetByID(ctx, *existingUser.RoleID)
		if err != nil {
			return nil, "", err
		}
		existingUser.Role = role
	}

	token, err := utils.GenerateJWT(existingUser.ID, existingUser.Email, existingUser.Role.Name, s.jwtSecretKey)
	if err != nil {
		return nil, "", err
	}

	return existingUser, token, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserProfile, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userRole, err := s.roleRepo.GetByID(ctx, *user.RoleID)
	if err != nil {
		return nil, err
	}

	userResponse := &dto.UserProfile{
		UUID:  user.UUID.String(),
		Email: user.Email,
		Name:  user.Name,
		Phone: user.Phone,
		Role:  userRole.Name,
	}

	return userResponse, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error {
	if req.Name == nil && req.Phone == nil {
		return domain_errors.ErrInvalidInput
	}

	err := s.userRepo.UpdateUserProfile(ctx, userID, req.Name, req.Phone)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) checkUserExists(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == domain_errors.ErrUserNotFound {
			return nil
		}
		return err
	}

	if existingUser != nil {
		return domain_errors.ErrEmailExists
	}

	return nil
}

func (s *UserService) createUser(ctx context.Context, req *dto.RegisterRequest, roleID int64) (*entities.User, error) {
	uuidV1, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
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

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

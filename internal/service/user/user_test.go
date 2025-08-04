package user_test

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"goshop/internal/service/user"
	"goshop/internal/service/user/mocks"
	"goshop/internal/utils"
)

const bcryptCost = 4

func stringPtr(s string) *string {
	return &s
}

func TestUserService_Register(t *testing.T) {
	userRole := &entities.Role{
		ID:   1,
		Name: "user",
	}

	tests := []struct {
		name          string
		request       *dto.RegisterRequest
		mockUserSetup func(*mocks.MockUserRepository)
		mockRoleSetup func(*mocks.MockRoleRepository)
		wantErr       bool
		expectedError error
		checkResult   bool
	}{
		{
			name: "Success_NewUser",
			request: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     stringPtr("Test User"),
				Phone:    stringPtr("+1234567890"),
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(nil, domain_errors.ErrUserNotFound)
				m.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(u *entities.User) bool {
					return u.Email == "test@example.com" && *u.Name == "Test User"
				})).Run(func(ctx context.Context, u *entities.User) {
					u.ID = 1
				}).Return(nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByName(mock.Anything, "user").Return(userRole, nil)
			},
			wantErr:     false,
			checkResult: true,
		},
		{
			name: "Error_UserAlreadyExists",
			request: &dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Name:     stringPtr("Test User"),
				Phone:    stringPtr("+1234567890"),
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				existingUser := &entities.User{
					ID:    1,
					Email: "existing@example.com",
				}
				m.EXPECT().GetUserByEmail(mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {},
			wantErr:       true,
			expectedError: domain_errors.ErrEmailExists,
		},
		{
			name: "Error_RoleNotFound",
			request: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     stringPtr("Test User"),
				Phone:    stringPtr("+1234567890"),
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(nil, domain_errors.ErrUserNotFound)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByName(mock.Anything, "user").Return(nil, errors.New("role not found"))
			},
			wantErr: true,
		},
		{
			name: "Error_CreateUserFails",
			request: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     stringPtr("Test User"),
				Phone:    stringPtr("+1234567890"),
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(nil, domain_errors.ErrUserNotFound)
				m.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByName(mock.Anything, "user").Return(userRole, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockRoleRepo := mocks.NewMockRoleRepository(t)

			tt.mockUserSetup(mockUserRepo)
			tt.mockRoleSetup(mockRoleRepo)

			service := user.NewUserService(mockRoleRepo, mockUserRepo, "test-secret", bcryptCost, zap.NewNop())
			ctx := context.Background()

			result, token, err := service.Register(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				assert.Nil(t, result)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.checkResult {
					assert.NotNil(t, result)
					assert.NotEmpty(t, token)
					assert.Equal(t, tt.request.Email, result.Email)
					assert.Equal(t, *tt.request.Name, *result.Name)
					assert.NotZero(t, result.UUID)
					assert.NotEmpty(t, result.PasswordHash)
				}
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	roleID := int64(1)
	userRole := &entities.Role{
		ID:   1,
		Name: "user",
	}

	hashedPassword, _ := utils.HashPasswordWithCost("password123", bcryptCost)

	tests := []struct {
		name          string
		request       *dto.LoginRequest
		mockUserSetup func(*mocks.MockUserRepository)
		mockRoleSetup func(*mocks.MockRoleRepository)
		wantErr       bool
		expectedError error
		checkResult   bool
	}{
		{
			name: "Success_ValidCredentials",
			request: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				user := &entities.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
					RoleID:       &roleID,
					Role:         nil,
				}
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(user, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(userRole, nil)
			},
			wantErr:     false,
			checkResult: true,
		},
		{
			name: "Error_UserNotFound",
			request: &dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByEmail(mock.Anything, "nonexistent@example.com").Return(nil, domain_errors.ErrUserNotFound)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {},
			wantErr:       true,
			expectedError: domain_errors.ErrUserNotFound,
		},
		{
			name: "Error_InvalidPassword",
			request: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				user := &entities.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
					RoleID:       &roleID,
				}
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(user, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {},
			wantErr:       true,
			expectedError: domain_errors.ErrInvalidPassword,
		},
		{
			name: "Error_RoleNotFound",
			request: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockUserSetup: func(m *mocks.MockUserRepository) {
				user := &entities.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
					RoleID:       &roleID,
				}
				m.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(user, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, errors.New("role not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockRoleRepo := mocks.NewMockRoleRepository(t)

			tt.mockUserSetup(mockUserRepo)
			tt.mockRoleSetup(mockRoleRepo)

			service := user.NewUserService(mockRoleRepo, mockUserRepo, "test-secret", bcryptCost, zap.NewNop())
			ctx := context.Background()

			result, token, err := service.Login(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				assert.Nil(t, result)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.checkResult {
					assert.NotNil(t, result)
					assert.NotEmpty(t, token)
					assert.Equal(t, tt.request.Email, result.Email)
					assert.NotNil(t, result.Role)
					assert.Equal(t, "user", result.Role.Name)
				}
			}
		})
	}
}

func TestUserService_GetUserProfile(t *testing.T) {
	roleID := int64(1)
	userUUID := uuid.New()

	tests := []struct {
		name          string
		userID        int64
		mockUserSetup func(*mocks.MockUserRepository)
		mockRoleSetup func(*mocks.MockRoleRepository)
		wantErr       bool
		expectedError error
		checkResult   bool
	}{
		{
			name:   "Success_ValidUserID",
			userID: 1,
			mockUserSetup: func(m *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        1,
					UUID:      userUUID,
					Email:     "test@example.com",
					Name:      stringPtr("Test User"),
					Phone:     stringPtr("+1234567890"),
					RoleID:    &roleID,
					CreatedAt: time.Now(),
				}
				m.EXPECT().GetUserByID(mock.Anything, int64(1)).Return(user, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				role := &entities.Role{
					ID:   1,
					Name: "user",
				}
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(role, nil)
			},
			wantErr:     false,
			checkResult: true,
		},
		{
			name:   "Error_UserNotFound",
			userID: 999,
			mockUserSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByID(mock.Anything, int64(999)).Return(nil, domain_errors.ErrUserNotFound)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {},
			wantErr:       true,
			expectedError: domain_errors.ErrUserNotFound,
		},
		{
			name:   "Error_RoleNotFound",
			userID: 1,
			mockUserSetup: func(m *mocks.MockUserRepository) {
				user := &entities.User{
					ID:     1,
					UUID:   userUUID,
					Email:  "test@example.com",
					RoleID: &roleID,
				}
				m.EXPECT().GetUserByID(mock.Anything, int64(1)).Return(user, nil)
			},
			mockRoleSetup: func(m *mocks.MockRoleRepository) {
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, errors.New("role not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockRoleRepo := mocks.NewMockRoleRepository(t)

			tt.mockUserSetup(mockUserRepo)
			tt.mockRoleSetup(mockRoleRepo)

			service := user.NewUserService(mockRoleRepo, mockUserRepo, "test-secret", bcryptCost, zap.NewNop())
			ctx := context.Background()

			result, err := service.GetUserProfile(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult {
					assert.NotNil(t, result)
					assert.Equal(t, userUUID.String(), result.UUID)
					assert.Equal(t, "test@example.com", result.Email)
					assert.Equal(t, "Test User", *result.Name)
					assert.Equal(t, "+1234567890", *result.Phone)
					assert.Equal(t, "user", result.Role)
				}
			}
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		request       *dto.UpdateProfileRequest
		mockSetup     func(*mocks.MockUserRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:   "Success_UpdateName",
			userID: 1,
			request: &dto.UpdateProfileRequest{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().UpdateUserProfile(mock.Anything, int64(1), stringPtr("Updated Name"), (*string)(nil)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Success_UpdatePhone",
			userID: 1,
			request: &dto.UpdateProfileRequest{
				Phone: stringPtr("+9876543210"),
			},
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().UpdateUserProfile(mock.Anything, int64(1), (*string)(nil), stringPtr("+9876543210")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Success_UpdateBoth",
			userID: 1,
			request: &dto.UpdateProfileRequest{
				Name:  stringPtr("Updated Name"),
				Phone: stringPtr("+9876543210"),
			},
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().UpdateUserProfile(mock.Anything, int64(1), stringPtr("Updated Name"), stringPtr("+9876543210")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Error_EmptyRequest",
			userID: 1,
			request: &dto.UpdateProfileRequest{
				Name:  nil,
				Phone: nil,
			},
			mockSetup:     func(m *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: domain_errors.ErrInvalidInput,
		},
		{
			name:   "Error_RepositoryFails",
			userID: 1,
			request: &dto.UpdateProfileRequest{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().UpdateUserProfile(mock.Anything, int64(1), stringPtr("Updated Name"), (*string)(nil)).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockRoleRepo := mocks.NewMockRoleRepository(t)

			tt.mockSetup(mockUserRepo)

			service := user.NewUserService(mockRoleRepo, mockUserRepo, "test-secret", bcryptCost, zap.NewNop())
			ctx := context.Background()

			err := service.UpdateProfile(ctx, tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

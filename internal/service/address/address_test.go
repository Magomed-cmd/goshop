package address_test

import (
	"context"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"testing"

	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/service/address"
	"goshop/internal/service/address/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func stringPtr(s string) *string {
	return &s
}

func TestAddressService_CreateAddress(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		request     *dto.CreateAddressRequest
		mockSetup   func(*mocks.MockAddressRepository)
		wantErr     bool
		errType     error
		checkResult bool
	}{
		{
			name:   "Success_ValidRequest",
			userID: 1,
			request: &dto.CreateAddressRequest{
				Address:    "123 Main St",
				City:       stringPtr("New York"),
				PostalCode: stringPtr("10001"),
				Country:    stringPtr("USA"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().CreateAddress(mock.Anything, mock.MatchedBy(func(addr *entities.UserAddress) bool {
					return addr.UserID == 1 && addr.Address == "123 Main St"
				})).Run(func(ctx context.Context, addr *entities.UserAddress) {
					addr.ID = 1
				}).Return(nil)
			},
			wantErr:     false,
			checkResult: true,
		},
		{
			name:   "Error_InvalidUserID",
			userID: 0,
			request: &dto.CreateAddressRequest{
				Address:    "123 Main St",
				City:       stringPtr("New York"),
				PostalCode: stringPtr("10001"),
				Country:    stringPtr("USA"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:   "Error_RepositoryFails",
			userID: 1,
			request: &dto.CreateAddressRequest{
				Address:    "123 Main St",
				City:       stringPtr("New York"),
				PostalCode: stringPtr("10001"),
				Country:    stringPtr("USA"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().CreateAddress(mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			result, err := service.CreateAddress(ctx, tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult {
					assert.NotNil(t, result)
					assert.Equal(t, tt.userID, result.UserID)
					assert.Equal(t, tt.request.Address, result.Address)
					assert.NotZero(t, result.UUID)
					assert.NotZero(t, result.CreatedAt)
				}
			}
		})
	}
}

func TestAddressService_GetUserAddresses(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockSetup     func(*mocks.MockAddressRepository)
		wantErr       bool
		errType       error
		expectedCount int
	}{
		{
			name:   "Success_ValidUserID",
			userID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				addresses := []*entities.UserAddress{
					{
						ID:      1,
						UserID:  1,
						Address: "123 Main St",
						City:    stringPtr("New York"),
					},
					{
						ID:      2,
						UserID:  1,
						Address: "456 Oak Ave",
						City:    stringPtr("Boston"),
					},
				}
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return(addresses, nil)
			},
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name:   "Success_EmptyAddresses",
			userID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return([]*entities.UserAddress{}, nil)
			},
			wantErr:       false,
			expectedCount: 0,
		},
		{
			name:      "Error_InvalidUserID",
			userID:    0,
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:   "Error_RepositoryFails",
			userID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return([]*entities.UserAddress(nil), errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			result, err := service.GetUserAddresses(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestAddressService_GetAddressByID(t *testing.T) {
	tests := []struct {
		name      string
		addressID int64
		mockSetup func(*mocks.MockAddressRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success_ValidAddressID",
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				address := &entities.UserAddress{
					ID:      1,
					UserID:  1,
					Address: "123 Main St",
					City:    stringPtr("New York"),
				}
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(address, nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidAddressID",
			addressID: 0,
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_AddressNotFound",
			addressID: 999,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(999)).Return(nil, errors2.ErrAddressNotFound)
			},
			wantErr: true,
			errType: errors2.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			result, err := service.GetAddressByID(ctx, tt.addressID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.addressID, result.ID)
			}
		})
	}
}

func TestAddressService_UpdateAddress(t *testing.T) {
	existingAddress := &entities.UserAddress{
		ID:      1,
		UserID:  1,
		Address: "123 Main St",
		City:    stringPtr("New York"),
	}

	tests := []struct {
		name      string
		userID    int64
		addressID int64
		request   *dto.UpdateAddressRequest
		mockSetup func(*mocks.MockAddressRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success_UpdateAddress",
			userID:    1,
			addressID: 1,
			request: &dto.UpdateAddressRequest{
				Address: stringPtr("456 Oak Ave"),
				City:    stringPtr("Boston"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
				m.EXPECT().UpdateAddress(mock.Anything, mock.MatchedBy(func(addr *entities.UserAddress) bool {
					return addr.ID == 1 && addr.Address == "456 Oak Ave"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidIDs",
			userID:    0,
			addressID: 1,
			request: &dto.UpdateAddressRequest{
				Address: stringPtr("456 Oak Ave"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_EmptyRequest",
			userID:    1,
			addressID: 1,
			request: &dto.UpdateAddressRequest{
				Address:    nil,
				City:       nil,
				PostalCode: nil,
				Country:    nil,
			},
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidAddressData,
		},
		{
			name:      "Error_AccessDenied",
			userID:    2,
			addressID: 1,
			request: &dto.UpdateAddressRequest{
				Address: stringPtr("456 Oak Ave"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
			},
			wantErr: true,
			errType: errors2.ErrForbidden,
		},
		{
			name:      "Error_AddressNotFound",
			userID:    1,
			addressID: 999,
			request: &dto.UpdateAddressRequest{
				Address: stringPtr("456 Oak Ave"),
			},
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(999)).Return(nil, errors2.ErrAddressNotFound)
			},
			wantErr: true,
			errType: errors2.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			result, err := service.UpdateAddress(ctx, tt.userID, tt.addressID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.addressID, result.ID)
				assert.Equal(t, tt.userID, result.UserID)
			}
		})
	}
}

func TestAddressService_GetAddressByIDForUser(t *testing.T) {
	existingAddress := &entities.UserAddress{
		ID:      1,
		UserID:  1,
		Address: "123 Main St",
		City:    stringPtr("New York"),
	}

	tests := []struct {
		name      string
		userID    int64
		addressID int64
		mockSetup func(*mocks.MockAddressRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success_ValidAccess",
			userID:    1,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidIDs",
			userID:    0,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_AccessDenied",
			userID:    2,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
			},
			wantErr: true,
			errType: errors2.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			result, err := service.GetAddressByIDForUser(ctx, tt.userID, tt.addressID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.addressID, result.ID)
				assert.Equal(t, tt.userID, result.UserID)
			}
		})
	}
}

func TestAddressService_DeleteAddress(t *testing.T) {
	existingAddress := &entities.UserAddress{
		ID:      1,
		UserID:  1,
		Address: "123 Main St",
		City:    stringPtr("New York"),
	}

	tests := []struct {
		name      string
		userID    int64
		addressID int64
		mockSetup func(*mocks.MockAddressRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success_ValidDelete",
			userID:    1,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
				m.EXPECT().DeleteAddress(mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidIDs",
			userID:    0,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_AccessDenied",
			userID:    2,
			addressID: 1,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(1)).Return(existingAddress, nil)
			},
			wantErr: true,
			errType: errors2.ErrForbidden,
		},
		{
			name:      "Error_AddressNotFound",
			userID:    1,
			addressID: 999,
			mockSetup: func(m *mocks.MockAddressRepository) {
				m.EXPECT().GetAddressByID(mock.Anything, int64(999)).Return(nil, errors2.ErrAddressNotFound)
			},
			wantErr: true,
			errType: errors2.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAddressRepository(t)
			tt.mockSetup(mockRepo)

			service := address.NewAddressService(mockRepo)
			ctx := context.Background()

			err := service.DeleteAddress(ctx, tt.userID, tt.addressID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

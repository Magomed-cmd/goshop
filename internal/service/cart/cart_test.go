package cart

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/errors"
	"goshop/internal/dto"
	"goshop/internal/service/cart/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCartService_GetCart(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockSetup     func(*mocks.MockCartRepository, *mocks.MockProductRepository)
		expectedError error
		checkResponse func(*testing.T, *dto.CartResponse)
	}{
		{
			name:   "Success_ExistingCart",
			userID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cart := &entities.Cart{
					ID:     1,
					UUID:   uuid.New(),
					UserID: 1,
					Items: []entities.CartItem{
						{
							CartID:    1,
							ProductID: 1,
							Quantity:  2,
							Product: &entities.Product{
								ID:    1,
								Name:  "Test Product",
								Price: decimal.NewFromFloat(10.99),
							},
						},
					},
				}
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(cart, nil)
			},
			expectedError: nil,
			checkResponse: func(t *testing.T, response *dto.CartResponse) {
				assert.Equal(t, int64(1), response.ID)
				assert.Len(t, response.Items, 1)
				assert.Equal(t, "Test Product", response.Items[0].ProductName)
				assert.Equal(t, "10.99", response.Items[0].Price)
				assert.Equal(t, "21.98", response.Items[0].Subtotal)
				assert.Equal(t, "21.98", response.TotalPrice)
				assert.Equal(t, 2, response.TotalItems)
			},
		},
		{
			name:   "Success_CreateNewCart",
			userID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
				cartRepo.EXPECT().CreateCart(mock.Anything, mock.AnythingOfType("*entities.Cart")).
					Run(func(ctx context.Context, cart *entities.Cart) {
						cart.ID = 1
					}).Return(nil)
			},
			expectedError: nil,
			checkResponse: func(t *testing.T, response *dto.CartResponse) {
				assert.Equal(t, int64(1), response.ID)
				assert.Len(t, response.Items, 0)
				assert.Equal(t, "0.00", response.TotalPrice)
				assert.Equal(t, 0, response.TotalItems)
			},
		},
		{
			name:   "Error_CreateCartFailed",
			userID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
				cartRepo.EXPECT().CreateCart(mock.Anything, mock.AnythingOfType("*entities.Cart")).
					Return(errors.ErrUserNotFound)
			},
			expectedError: errors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			productRepo := mocks.NewMockProductRepository(t)
			tt.mockSetup(cartRepo, productRepo)

			service := NewCartService(cartRepo, productRepo)
			response, err := service.GetCart(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
			}
		})
	}
}

func TestCartService_AddItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		request       *dto.AddToCartRequest
		mockSetup     func(*mocks.MockCartRepository, *mocks.MockProductRepository)
		expectedError error
	}{
		{
			name:   "Success_AddToExistingCart",
			userID: 1,
			request: &dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  2,
			},
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				product := &entities.Product{
					ID:    1,
					Name:  "Test Product",
					Stock: 10,
					Price: decimal.NewFromFloat(10.99),
				}
				cart := &entities.Cart{
					ID:     1,
					UserID: 1,
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(product, nil)
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(cart, nil)
				cartRepo.EXPECT().AddItem(mock.Anything, int64(1), int64(1), 2).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "Success_CreateCartAndAddItem",
			userID: 1,
			request: &dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  2,
			},
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				product := &entities.Product{
					ID:    1,
					Name:  "Test Product",
					Stock: 10,
					Price: decimal.NewFromFloat(10.99),
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(product, nil)
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
				cartRepo.EXPECT().CreateCart(mock.Anything, mock.AnythingOfType("*entities.Cart")).
					Run(func(ctx context.Context, cart *entities.Cart) {
						cart.ID = 1
					}).Return(nil)
				cartRepo.EXPECT().AddItem(mock.Anything, int64(1), int64(1), 2).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "Error_InvalidQuantity",
			userID: 1,
			request: &dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  0,
			},
			mockSetup:     func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {},
			expectedError: errors.ErrInvalidQuantity,
		},
		{
			name:   "Error_ProductNotFound",
			userID: 1,
			request: &dto.AddToCartRequest{
				ProductID: 999,
				Quantity:  2,
			},
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(999)).Return(nil, errors.ErrProductNotFound)
			},
			expectedError: errors.ErrProductNotFound,
		},
		{
			name:   "Error_InsufficientStock",
			userID: 1,
			request: &dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  10,
			},
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				product := &entities.Product{
					ID:    1,
					Name:  "Test Product",
					Stock: 5,
					Price: decimal.NewFromFloat(10.99),
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(product, nil)
			},
			expectedError: errors.ErrInsufficientStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			productRepo := mocks.NewMockProductRepository(t)
			tt.mockSetup(cartRepo, productRepo)

			service := NewCartService(cartRepo, productRepo)
			err := service.AddItem(context.Background(), tt.userID, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartService_UpdateItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		productID     int64
		quantity      int
		mockSetup     func(*mocks.MockCartRepository, *mocks.MockProductRepository)
		expectedError error
	}{
		{
			name:      "Success_UpdateItem",
			userID:    1,
			productID: 1,
			quantity:  5,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				product := &entities.Product{
					ID:    1,
					Stock: 10,
				}
				cart := &entities.Cart{
					ID:     1,
					UserID: 1,
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(product, nil)
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(cart, nil)
				cartRepo.EXPECT().UpdateItem(mock.Anything, int64(1), int64(1), 5).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error_InvalidQuantity",
			userID:        1,
			productID:     1,
			quantity:      0,
			mockSetup:     func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {},
			expectedError: errors.ErrInvalidQuantity,
		},
		{
			name:      "Error_CartNotFound",
			userID:    1,
			productID: 1,
			quantity:  5,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				product := &entities.Product{
					ID:    1,
					Stock: 10,
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(product, nil)
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
			},
			expectedError: errors.ErrCartNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			productRepo := mocks.NewMockProductRepository(t)
			tt.mockSetup(cartRepo, productRepo)

			service := NewCartService(cartRepo, productRepo)
			err := service.UpdateItem(context.Background(), tt.userID, tt.productID, tt.quantity)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		productID     int64
		mockSetup     func(*mocks.MockCartRepository, *mocks.MockProductRepository)
		expectedError error
	}{
		{
			name:      "Success_RemoveItem",
			userID:    1,
			productID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cart := &entities.Cart{
					ID:     1,
					UserID: 1,
				}
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(cart, nil)
				cartRepo.EXPECT().RemoveItem(mock.Anything, int64(1), int64(1)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "Error_CartNotFound",
			userID:    1,
			productID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
			},
			expectedError: errors.ErrCartNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			productRepo := mocks.NewMockProductRepository(t)
			tt.mockSetup(cartRepo, productRepo)

			service := NewCartService(cartRepo, productRepo)
			err := service.RemoveItem(context.Background(), tt.userID, tt.productID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartService_ClearCart(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockSetup     func(*mocks.MockCartRepository, *mocks.MockProductRepository)
		expectedError error
	}{
		{
			name:   "Success_ClearCart",
			userID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cart := &entities.Cart{
					ID:     1,
					UserID: 1,
				}
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(cart, nil)
				cartRepo.EXPECT().ClearCart(mock.Anything, int64(1)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "Error_CartNotFound",
			userID: 1,
			mockSetup: func(cartRepo *mocks.MockCartRepository, productRepo *mocks.MockProductRepository) {
				cartRepo.EXPECT().GetUserCart(mock.Anything, int64(1)).Return(nil, errors.ErrCartNotFound)
			},
			expectedError: errors.ErrCartNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := mocks.NewMockCartRepository(t)
			productRepo := mocks.NewMockProductRepository(t)
			tt.mockSetup(cartRepo, productRepo)

			service := NewCartService(cartRepo, productRepo)
			err := service.ClearCart(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

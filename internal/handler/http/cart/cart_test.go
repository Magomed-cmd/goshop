package cart

import (
	"bytes"
	"encoding/json"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"net/http/httptest"
	"testing"

	"goshop/internal/dto"
	"goshop/internal/handler/cart/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupAuthContext(c *gin.Context, userID int64) {
	c.Set("user_id", userID)
}

func TestCartHandler_GetCart(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		mockSetup      func(*mocks.MockCartService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Success_GetCart",
			userID: 1,
			mockSetup: func(m *mocks.MockCartService) {
				cartResponse := &dto.CartResponse{
					ID:         1,
					Items:      []dto.CartItemResponse{},
					TotalPrice: "0.00",
					TotalItems: 0,
				}
				m.EXPECT().GetCart(mock.Anything, int64(1)).Return(cartResponse, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.CartResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "0.00", response.TotalPrice)
			},
		},
		{
			name:   "Error_ServiceFailure",
			userID: 1,
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().GetCart(mock.Anything, int64(1)).Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Internal server error", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCartService(t)
			tt.mockSetup(mockService)

			handler := NewCartHandler(mockService)
			router := setupTestRouter()
			router.GET("/cart", func(c *gin.Context) {
				setupAuthContext(c, tt.userID)
				handler.GetCart(c)
			})

			req := httptest.NewRequest("GET", "/cart", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCartHandler_AddItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		requestBody    interface{}
		mockSetup      func(*mocks.MockCartService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Success_AddItem",
			userID: 1,
			requestBody: map[string]interface{}{
				"product_id": 1,
				"quantity":   2,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().AddItem(mock.Anything, int64(1), mock.AnythingOfType("*dto.AddToCartRequest")).Return(nil)
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Item added to cart", response["message"])
			},
		},
		{
			name:           "Error_InvalidJSON",
			userID:         1,
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockCartService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:   "Error_InvalidQuantity_GinValidation",
			userID: 1,
			requestBody: map[string]interface{}{
				"product_id": 1,
				"quantity":   0,
			},
			mockSetup:      func(m *mocks.MockCartService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:   "Error_InvalidQuantity_ServiceLevel",
			userID: 1,
			requestBody: map[string]interface{}{
				"product_id": 1,
				"quantity":   1,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().AddItem(mock.Anything, int64(1), mock.AnythingOfType("*dto.AddToCartRequest")).
					Return(errors2.ErrInvalidQuantity)
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid quantity", response["error"])
			},
		},
		{
			name:   "Error_ProductNotFound",
			userID: 1,
			requestBody: map[string]interface{}{
				"product_id": 999,
				"quantity":   2,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().AddItem(mock.Anything, int64(1), mock.AnythingOfType("*dto.AddToCartRequest")).
					Return(errors2.ErrProductNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Product not found", response["error"])
			},
		},
		{
			name:   "Error_InsufficientStock",
			userID: 1,
			requestBody: map[string]interface{}{
				"product_id": 1,
				"quantity":   10,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().AddItem(mock.Anything, int64(1), mock.AnythingOfType("*dto.AddToCartRequest")).
					Return(errors2.ErrInsufficientStock)
			},
			expectedStatus: 409,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Insufficient stock", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCartService(t)
			tt.mockSetup(mockService)

			handler := NewCartHandler(mockService)
			router := setupTestRouter()
			router.POST("/cart/items", func(c *gin.Context) {
				setupAuthContext(c, tt.userID)
				handler.AddItem(c)
			})

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/cart/items", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCartHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		productID      string
		requestBody    interface{}
		mockSetup      func(*mocks.MockCartService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Success_UpdateItem",
			userID:    1,
			productID: "1",
			requestBody: map[string]interface{}{
				"quantity": 5,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().UpdateItem(mock.Anything, int64(1), int64(1), 5).Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Item updated in cart", response["message"])
			},
		},
		{
			name:           "Error_InvalidProductID",
			userID:         1,
			productID:      "invalid",
			requestBody:    map[string]interface{}{"quantity": 5},
			mockSetup:      func(m *mocks.MockCartService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid product ID", response["error"])
			},
		},
		{
			name:           "Error_InvalidJSON",
			userID:         1,
			productID:      "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockCartService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:      "Error_CartItemNotFound",
			userID:    1,
			productID: "1",
			requestBody: map[string]interface{}{
				"quantity": 5,
			},
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().UpdateItem(mock.Anything, int64(1), int64(1), 5).
					Return(errors2.ErrCartItemNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Item not found in cart", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCartService(t)
			tt.mockSetup(mockService)

			handler := NewCartHandler(mockService)
			router := setupTestRouter()
			router.PUT("/cart/items/:product_id", func(c *gin.Context) {
				setupAuthContext(c, tt.userID)
				handler.UpdateItem(c)
			})

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/cart/items/"+tt.productID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCartHandler_RemoveItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		productID      string
		mockSetup      func(*mocks.MockCartService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Success_RemoveItem",
			userID:    1,
			productID: "1",
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().RemoveItem(mock.Anything, int64(1), int64(1)).Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Item removed from cart", response["message"])
			},
		},
		{
			name:           "Error_InvalidProductID",
			userID:         1,
			productID:      "invalid",
			mockSetup:      func(m *mocks.MockCartService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid product ID", response["error"])
			},
		},
		{
			name:      "Error_CartItemNotFound",
			userID:    1,
			productID: "1",
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().RemoveItem(mock.Anything, int64(1), int64(1)).
					Return(errors2.ErrCartItemNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Item not found in cart", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCartService(t)
			tt.mockSetup(mockService)

			handler := NewCartHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/cart/items/:product_id", func(c *gin.Context) {
				setupAuthContext(c, tt.userID)
				handler.RemoveItem(c)
			})

			req := httptest.NewRequest("DELETE", "/cart/items/"+tt.productID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCartHandler_ClearCart(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		mockSetup      func(*mocks.MockCartService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Success_ClearCart",
			userID: 1,
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().ClearCart(mock.Anything, int64(1)).Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Cart cleared", response["message"])
			},
		},
		{
			name:   "Error_CartNotFound",
			userID: 1,
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().ClearCart(mock.Anything, int64(1)).
					Return(errors2.ErrCartNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Cart not found", response["error"])
			},
		},
		{
			name:   "Error_ServiceFailure",
			userID: 1,
			mockSetup: func(m *mocks.MockCartService) {
				m.EXPECT().ClearCart(mock.Anything, int64(1)).
					Return(errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Internal server error", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCartService(t)
			tt.mockSetup(mockService)

			handler := NewCartHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/cart", func(c *gin.Context) {
				setupAuthContext(c, tt.userID)
				handler.ClearCart(c)
			})

			req := httptest.NewRequest("DELETE", "/cart", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

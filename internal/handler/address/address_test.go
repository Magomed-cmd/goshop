package address_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/handler/address"
	addressMocks "goshop/internal/handler/address/mocks"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func stringPtr(s string) *string {
	return &s
}

func setupContext(c *gin.Context, userID int64) {
	c.Set("user_id", userID)
}

func TestAddressHandler_CreateAddress(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupContext   func(*gin.Context)
		mockSetup      func(*addressMocks.MockAddressService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ValidRequest",
			requestBody: map[string]interface{}{
				"address":     "123 Main Street, Apt 4B",
				"city":        "Moscow",
				"postal_code": "123456",
				"country":     "Russia",
			},
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				createdAddress := &entities.UserAddress{
					ID:         1,
					UUID:       uuid.New(),
					UserID:     1,
					Address:    "123 Main Street, Apt 4B",
					City:       stringPtr("Moscow"),
					PostalCode: stringPtr("123456"),
					Country:    stringPtr("Russia"),
					CreatedAt:  time.Now(),
				}
				m.EXPECT().CreateAddress(mock.Anything, int64(1), mock.AnythingOfType("*dto.CreateAddressRequest")).
					Return(createdAddress, nil)
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "123 Main Street, Apt 4B", response.Address)
				assert.Equal(t, "Moscow", *response.City)
			},
		},
		{
			name: "Error_Unauthorized",
			requestBody: map[string]interface{}{
				"address": "123 Main Street",
			},
			setupContext: func(c *gin.Context) {
				// Не устанавливаем user_id
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name:        "Error_InvalidJSON",
			requestBody: "invalid json",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid character")
			},
		},
		{
			name: "Error_ServiceFailure",
			requestBody: map[string]interface{}{
				"address": "123 Main Street, Apt 4B",
			},
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().CreateAddress(mock.Anything, int64(1), mock.AnythingOfType("*dto.CreateAddressRequest")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := addressMocks.NewMockAddressService(t)
			tt.mockSetup(mockService)

			addressHandler := address.NewAddressHandler(mockService)
			router := setupTestRouter()
			router.POST("/addresses", func(c *gin.Context) {
				tt.setupContext(c)
				addressHandler.CreateAddress(c)
			})

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/addresses", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAddressHandler_GetUserAddresses(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		mockSetup      func(*addressMocks.MockAddressService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ReturnsAddresses",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				addresses := []*entities.UserAddress{
					{
						ID:         1,
						UUID:       uuid.New(),
						UserID:     1,
						Address:    "123 Main Street",
						City:       stringPtr("Moscow"),
						PostalCode: stringPtr("123456"),
						Country:    stringPtr("Russia"),
						CreatedAt:  time.Now(),
					},
					{
						ID:         2,
						UUID:       uuid.New(),
						UserID:     1,
						Address:    "456 Oak Avenue",
						City:       stringPtr("St. Petersburg"),
						PostalCode: stringPtr("654321"),
						Country:    stringPtr("Russia"),
						CreatedAt:  time.Now(),
					},
				}
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return(addresses, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []dto.AddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
				assert.Equal(t, "123 Main Street", response[0].Address)
				assert.Equal(t, "456 Oak Avenue", response[1].Address)
			},
		},
		{
			name: "Success_EmptyList",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return([]*entities.UserAddress{}, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []dto.AddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name: "Error_Unauthorized",
			setupContext: func(c *gin.Context) {
				// Не устанавливаем user_id
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name: "Error_ServiceFailure",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().GetUserAddresses(mock.Anything, int64(1)).Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := addressMocks.NewMockAddressService(t)
			tt.mockSetup(mockService)

			addressHandler := address.NewAddressHandler(mockService)
			router := setupTestRouter()
			router.GET("/addresses", func(c *gin.Context) {
				tt.setupContext(c)
				addressHandler.GetUserAddresses(c)
			})

			req := httptest.NewRequest("GET", "/addresses", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAddressHandler_GetAddressByID(t *testing.T) {
	tests := []struct {
		name           string
		addressID      string
		setupContext   func(*gin.Context)
		mockSetup      func(*addressMocks.MockAddressService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Success_ValidID",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				address := &entities.UserAddress{
					ID:         1,
					UUID:       uuid.New(),
					UserID:     1,
					Address:    "123 Main Street",
					City:       stringPtr("Moscow"),
					PostalCode: stringPtr("123456"),
					Country:    stringPtr("Russia"),
					CreatedAt:  time.Now(),
				}
				m.EXPECT().GetAddressByIDForUser(mock.Anything, int64(1), int64(1)).Return(address, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "123 Main Street", response.Address)
			},
		},
		{
			name:      "Error_Unauthorized",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				// Не устанавливаем user_id
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name:      "Error_InvalidID",
			addressID: "invalid",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid address ID", response["error"])
			},
		},
		{
			name:      "Error_AccessDenied",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				setupContext(c, 2) // Другой пользователь
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().GetAddressByIDForUser(mock.Anything, int64(2), int64(1)).
					Return(nil, errors.New("access denied"))
			},
			expectedStatus: 403,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Access denied", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := addressMocks.NewMockAddressService(t)
			tt.mockSetup(mockService)

			addressHandler := address.NewAddressHandler(mockService)
			router := setupTestRouter()
			router.GET("/addresses/:addressID", func(c *gin.Context) {
				tt.setupContext(c)
				addressHandler.GetAddressByID(c)
			})

			req := httptest.NewRequest("GET", "/addresses/"+tt.addressID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAddressHandler_UpdateAddress(t *testing.T) {
	tests := []struct {
		name           string
		addressID      string
		requestBody    interface{}
		setupContext   func(*gin.Context)
		mockSetup      func(*addressMocks.MockAddressService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Success_ValidUpdate",
			addressID: "1",
			requestBody: map[string]interface{}{
				"address": "Updated Street 123",
				"city":    "Updated City",
			},
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				updatedAddress := &entities.UserAddress{
					ID:         1,
					UUID:       uuid.New(),
					UserID:     1,
					Address:    "Updated Street 123",
					City:       stringPtr("Updated City"),
					PostalCode: stringPtr("123456"),
					Country:    stringPtr("Russia"),
					CreatedAt:  time.Now(),
				}
				m.EXPECT().UpdateAddress(mock.Anything, int64(1), int64(1), mock.AnythingOfType("*dto.UpdateAddressRequest")).
					Return(updatedAddress, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Updated Street 123", response.Address)
				assert.Equal(t, "Updated City", *response.City)
			},
		},
		{
			name:      "Error_Unauthorized",
			addressID: "1",
			requestBody: map[string]interface{}{
				"address": "Updated Street",
			},
			setupContext: func(c *gin.Context) {
				// Не устанавливаем user_id
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name:        "Error_InvalidJSON",
			addressID:   "1",
			requestBody: "invalid json",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid character")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := addressMocks.NewMockAddressService(t)
			tt.mockSetup(mockService)

			addressHandler := address.NewAddressHandler(mockService)
			router := setupTestRouter()
			router.PUT("/addresses/:addressID", func(c *gin.Context) {
				tt.setupContext(c)
				addressHandler.UpdateAddress(c)
			})

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/addresses/"+tt.addressID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAddressHandler_DeleteAddress(t *testing.T) {
	tests := []struct {
		name           string
		addressID      string
		setupContext   func(*gin.Context)
		mockSetup      func(*addressMocks.MockAddressService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Success_ValidDelete",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().DeleteAddress(mock.Anything, int64(1), int64(1)).Return(nil)
			},
			expectedStatus: 204,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Empty(t, w.Body.String())
			},
		},
		{
			name:      "Error_Unauthorized",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				// Не устанавливаем user_id
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name:      "Error_InvalidID",
			addressID: "invalid",
			setupContext: func(c *gin.Context) {
				setupContext(c, 1)
			},
			mockSetup:      func(m *addressMocks.MockAddressService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid address ID", response["error"])
			},
		},
		{
			name:      "Error_AccessDenied",
			addressID: "1",
			setupContext: func(c *gin.Context) {
				setupContext(c, 2)
			},
			mockSetup: func(m *addressMocks.MockAddressService) {
				m.EXPECT().DeleteAddress(mock.Anything, int64(2), int64(1)).
					Return(errors.New("access denied"))
			},
			expectedStatus: 403,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Access denied", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := addressMocks.NewMockAddressService(t)
			tt.mockSetup(mockService)

			addressHandler := address.NewAddressHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/addresses/:addressID", func(c *gin.Context) {
				tt.setupContext(c)
				addressHandler.DeleteAddress(c)
			})

			req := httptest.NewRequest("DELETE", "/addresses/"+tt.addressID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

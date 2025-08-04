package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"goshop/internal/handler/user"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/handler/user/mocks"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func stringPtr(s string) *string {
	return &s
}

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ValidRegistration",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     stringPtr("Test User"),
				Phone:    stringPtr("+1234567890"),
			},
			mockSetup: func(m *mocks.MockUserService) {
				user := &entities.User{
					ID:    1,
					UUID:  uuid.New(),
					Email: "test@example.com",
					Name:  stringPtr("Test User"),
					Phone: stringPtr("+1234567890"),
					Role: &entities.Role{
						ID:   1,
						Name: "user",
					},
				}
				m.EXPECT().Register(mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).
					Return(user, "jwt-token-123", nil)
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "jwt-token-123", response.Token)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "user", response.User.Role)
			},
		},
		{
			name: "Success_MinimalRegistration",
			requestBody: dto.RegisterRequest{
				Email:    "minimal@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserService) {
				user := &entities.User{
					ID:    2,
					UUID:  uuid.New(),
					Email: "minimal@example.com",
					Role: &entities.Role{
						ID:   1,
						Name: "user",
					},
				}
				m.EXPECT().Register(mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).
					Return(user, "jwt-token-456", nil)
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "jwt-token-456", response.Token)
				assert.Equal(t, "minimal@example.com", response.User.Email)
			},
		},
		{
			name: "Error_UserAlreadyExists",
			requestBody: dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().Register(mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).
					Return(nil, "", errors.New("user with this email already exists"))
			},
			expectedStatus: 409,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User already exists", response["error"])
			},
		},
		{
			name: "Error_ServiceFailure",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().Register(mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).
					Return(nil, "", errors.New("database connection failed"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Internal server error", response["error"])
			},
		},
		{
			name:           "Error_InvalidJSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "Error_MissingEmail",
			requestBody: map[string]interface{}{
				"password": "password123",
				"name":     "Test User",
			},
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			tt.mockSetup(mockService)

			userHandler := user.NewUserHandler(mockService)
			router := setupTestRouter()
			router.POST("/register", userHandler.Register)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ValidLogin",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserService) {
				user := &entities.User{
					ID:    1,
					UUID:  uuid.New(),
					Email: "test@example.com",
					Name:  stringPtr("Test User"),
					Role: &entities.Role{
						ID:   1,
						Name: "user",
					},
				}
				m.EXPECT().Login(mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).
					Return(user, "login-token-123", nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "login-token-123", response.Token)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "user", response.User.Role)
			},
		},
		{
			name: "Error_InvalidCredentials",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().Login(mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).
					Return(nil, "", errors.New("invalid email or password"))
			},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid email or password", response["error"])
			},
		},
		{
			name:           "Error_InvalidJSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "Error_MissingPassword",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			tt.mockSetup(mockService)

			userHandler := user.NewUserHandler(mockService)
			router := setupTestRouter()
			router.POST("/login", userHandler.Login)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ValidProfile",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
				c.Set("email", "test@example.com")
				c.Set("role", "user")
			},
			mockSetup: func(m *mocks.MockUserService) {
				profile := &dto.UserProfile{
					UUID:  uuid.New().String(),
					Email: "test@example.com",
					Name:  stringPtr("Test User"),
					Phone: stringPtr("+1234567890"),
					Role:  "user",
				}
				m.EXPECT().GetUserProfile(mock.Anything, int64(1)).Return(profile, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.UserProfile
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test@example.com", response.Email)
				assert.Equal(t, "user", response.Role)
			},
		},
		{
			name: "Error_NoUserID",
			setupContext: func(c *gin.Context) {
			},
			mockSetup:      func(m *mocks.MockUserService) {},
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
				c.Set("user_id", int64(1))
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().GetUserProfile(mock.Anything, int64(1)).
					Return(nil, errors.New("database error"))
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
			mockService := mocks.NewMockUserService(t)
			tt.mockSetup(mockService)

			userHandler := user.NewUserHandler(mockService)

			router := setupTestRouter()
			router.GET("/profile", func(c *gin.Context) {
				tt.setupContext(c)
				userHandler.GetProfile(c)
			})

			req := httptest.NewRequest("GET", "/profile", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupContext   func(*gin.Context)
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_UpdateBothFields",
			requestBody: dto.UpdateProfileRequest{
				Name:  stringPtr("Updated Name"),
				Phone: stringPtr("+9876543210"),
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().UpdateProfile(mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).
					Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Profile updated successfully", response["message"])
			},
		},
		{
			name: "Success_UpdateNameOnly",
			requestBody: dto.UpdateProfileRequest{
				Name: stringPtr("Name Only"),
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().UpdateProfile(mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).
					Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Profile updated successfully", response["message"])
			},
		},
		{
			name: "Error_NoUserID",
			requestBody: dto.UpdateProfileRequest{
				Name: stringPtr("Test"),
			},
			setupContext: func(c *gin.Context) {
			},
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 401,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
		{
			name: "Error_NoFieldsToUpdate",
			requestBody: dto.UpdateProfileRequest{
				Name: stringPtr("Test"),
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().UpdateProfile(mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).
					Return(errors.New("no fields to update"))
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "No fields to update", response["error"])
			},
		},
		{
			name:        "Error_InvalidJSON",
			requestBody: "invalid json",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
			},
			mockSetup:      func(m *mocks.MockUserService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "Error_ServiceFailure",
			requestBody: dto.UpdateProfileRequest{
				Name: stringPtr("Test"),
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", int64(1))
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.EXPECT().UpdateProfile(mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateProfileRequest")).
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
			mockService := mocks.NewMockUserService(t)
			tt.mockSetup(mockService)

			userHandler := user.NewUserHandler(mockService)

			router := setupTestRouter()
			router.PUT("/profile", func(c *gin.Context) {
				tt.setupContext(c)
				userHandler.UpdateProfile(c)
			})

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

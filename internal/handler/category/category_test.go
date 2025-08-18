package category_test

import (
	"bytes"
	"encoding/json"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"net/http/httptest"
	"testing"

	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/handler/category"
	"goshop/internal/handler/category/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func stringPtr(s string) *string {
	return &s
}

func TestCategoryHandler_GetAllCategories(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockCategoryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ReturnsCategoriesList",
			mockSetup: func(m *mocks.MockCategoryService) {
				categories := []*entities.CategoryWithCount{
					{
						Category: entities.Category{
							ID:          1,
							UUID:        uuid.New(),
							Name:        "Electronics",
							Description: stringPtr("Electronic devices"),
						},
						ProductCount: 5,
					},
					{
						Category: entities.Category{
							ID:          2,
							UUID:        uuid.New(),
							Name:        "Books",
							Description: stringPtr("All kinds of books"),
						},
						ProductCount: 3,
					},
				}
				m.EXPECT().GetAllCategories(mock.Anything).Return(categories, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var categories []dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &categories)
				assert.NoError(t, err)
				assert.Len(t, categories, 2)
				assert.Equal(t, "Electronics", categories[0].Name)
				assert.Equal(t, 5, categories[0].ProductCount)
			},
		},
		{
			name: "Success_EmptyList",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().GetAllCategories(mock.Anything).Return([]*entities.CategoryWithCount{}, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var categories []*entities.CategoryWithCount
				err := json.Unmarshal(w.Body.Bytes(), &categories)
				assert.NoError(t, err)
				assert.Len(t, categories, 0)
			},
		},
		{
			name: "Error_ServiceFailure",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().GetAllCategories(mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Failed to fetch categories", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCategoryService(t)
			tt.mockSetup(mockService)

			categoryHandler := category.NewCategoryHandler(mockService)
			router := setupTestRouter()
			router.GET("/categories", categoryHandler.GetAllCategories)

			req := httptest.NewRequest("GET", "/categories", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCategoryHandler_GetCategoryByID(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		mockSetup      func(*mocks.MockCategoryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "Success_ValidID",
			categoryID: "1",
			mockSetup: func(m *mocks.MockCategoryService) {
				category := &dto.CategoryResponse{
					ID:           1,
					UUID:         uuid.New().String(),
					Name:         "Electronics",
					Description:  stringPtr("Electronic devices"),
					ProductCount: 5,
				}
				m.EXPECT().GetCategoryByID(mock.Anything, int64(1)).Return(category, nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var category dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &category)
				assert.NoError(t, err)
				assert.Equal(t, "Electronics", category.Name)
				assert.Equal(t, 5, category.ProductCount)
			},
		},
		{
			name:           "Error_InvalidID",
			categoryID:     "invalid",
			mockSetup:      func(m *mocks.MockCategoryService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid category ID", response["error"])
			},
		},
		{
			name:       "Error_CategoryNotFound",
			categoryID: "999",
			mockSetup: func(m *mocks.MockCategoryService) {
				// Только GetCategoryByID вызывается, UpdateCategory НЕ вызывается при ошибке
				m.EXPECT().GetCategoryByID(mock.Anything, int64(999)).Return(nil, errors2.ErrCategoryNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Category not found", response["error"])
			},
		},
		{
			name:       "Error_ServiceFailure",
			categoryID: "1",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().GetCategoryByID(mock.Anything, int64(1)).Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Failed to fetch category", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCategoryService(t)
			tt.mockSetup(mockService)

			categoryHandler := category.NewCategoryHandler(mockService)
			router := setupTestRouter()
			router.GET("/categories/:id", categoryHandler.GetCategoryByID)

			req := httptest.NewRequest("GET", "/categories/"+tt.categoryID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockCategoryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success_ValidRequest",
			requestBody: map[string]interface{}{
				"name":        "Electronics",
				"description": "Electronic devices",
			},
			mockSetup: func(m *mocks.MockCategoryService) {
				createdCategory := &entities.Category{
					ID:          1,
					UUID:        uuid.New(),
					Name:        "Electronics",
					Description: stringPtr("Electronic devices"),
				}
				m.EXPECT().CreateCategory(mock.Anything, mock.AnythingOfType("*dto.CreateCategoryRequest")).
					Return(createdCategory, nil)
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Electronics", response.Name)
				assert.Equal(t, 0, response.ProductCount)
			},
		},
		{
			name:           "Error_InvalidJSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockCategoryService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name: "Error_ServiceFailure",
			requestBody: map[string]interface{}{
				"name":        "Electronics",
				"description": "Electronic devices",
			},
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().CreateCategory(mock.Anything, mock.AnythingOfType("*dto.CreateCategoryRequest")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Failed to create category", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCategoryService(t)
			tt.mockSetup(mockService)

			categoryHandler := category.NewCategoryHandler(mockService)
			router := setupTestRouter()
			router.POST("/categories", categoryHandler.CreateCategory)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		requestBody    interface{}
		mockSetup      func(*mocks.MockCategoryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "Success_ValidUpdate",
			categoryID: "1",
			requestBody: map[string]interface{}{
				"name":        "Updated Electronics",
				"description": "Updated description",
			},
			mockSetup: func(m *mocks.MockCategoryService) {
				// First, expect GetCategoryByID call
				category := &dto.CategoryResponse{
					ID:           1,
					UUID:         uuid.New().String(),
					Name:         "Updated Electronics",
					Description:  stringPtr("Updated description"),
					ProductCount: 5,
				}
				m.EXPECT().GetCategoryByID(mock.Anything, int64(1)).Return(category, nil)
				m.EXPECT().UpdateCategory(mock.Anything, mock.AnythingOfType("*entities.Category")).Return(nil)
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Updated Electronics", response.Name)
				assert.Equal(t, 5, response.ProductCount)
			},
		},
		{
			name:       "Error_InvalidID",
			categoryID: "invalid",
			requestBody: map[string]interface{}{
				"name": "Updated",
			},
			mockSetup:      func(m *mocks.MockCategoryService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid category ID", response["error"])
			},
		},
		{
			name:       "Error_CategoryNotFound",
			categoryID: "999",
			requestBody: map[string]interface{}{
				"name": "Updated",
			},
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().GetCategoryByID(mock.Anything, int64(999)).Return(nil, errors2.ErrCategoryNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Category not found", response["error"])
			},
		},
		{
			name:           "Error_InvalidJSON",
			categoryID:     "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockCategoryService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCategoryService(t)
			tt.mockSetup(mockService)

			categoryHandler := category.NewCategoryHandler(mockService)
			router := setupTestRouter()
			router.PUT("/categories/:id", categoryHandler.UpdateCategory)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/categories/"+tt.categoryID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		mockSetup      func(*mocks.MockCategoryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "Success_ValidDelete",
			categoryID: "1",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().DeleteCategory(mock.Anything, int64(1)).Return(nil)
			},
			expectedStatus: 204,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Empty(t, w.Body.String())
			},
		},
		{
			name:           "Error_InvalidID",
			categoryID:     "invalid",
			mockSetup:      func(m *mocks.MockCategoryService) {},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid category ID", response["error"])
			},
		},
		{
			name:       "Error_CategoryNotFound",
			categoryID: "999",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().DeleteCategory(mock.Anything, int64(999)).Return(errors2.ErrCategoryNotFound)
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Category not found", response["error"])
			},
		},
		{
			name:       "Error_ServiceFailure",
			categoryID: "1",
			mockSetup: func(m *mocks.MockCategoryService) {
				m.EXPECT().DeleteCategory(mock.Anything, int64(1)).Return(errors.New("database error"))
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Failed to delete category", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockCategoryService(t)
			tt.mockSetup(mockService)

			categoryHandler := category.NewCategoryHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/categories/:id", categoryHandler.DeleteCategory)

			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

package product_test

import (
	"bytes"
	"encoding/json"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"goshop/internal/handler/http/product"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"goshop/internal/domain/types"
	"goshop/internal/dto"
	"goshop/internal/handler/product/mocks"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestProductHandler_GetProducts(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:        "Success_DefaultFilters",
			queryParams: "",
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductCatalogResponse{
					Products: []dto.ProductCatalogItem{
						{ID: 1, Name: "Product 1", Price: "100"},
					},
					Total: 1,
					Page:  1,
					Limit: 20,
				}
				m.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:        "Success_CustomFilters",
			queryParams: "?page=2&limit=10&category_id=5&sort_by=price&sort_order=desc&min_price=50&max_price=200",
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductCatalogResponse{
					Products: []dto.ProductCatalogItem{},
					Total:    0,
					Page:     2,
					Limit:    10,
				}
				m.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return f.Page == 2 && f.Limit == 10 && *f.CategoryID == 5
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:        "Error_ServiceFailure",
			queryParams: "",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().GetProducts(mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.GET("/products", handler.GetProducts)

			req := httptest.NewRequest("GET", "/products"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse && w.Code == http.StatusOK {
				var response dto.ProductCatalogResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductHandler_GetProductByID(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:      "Success_ValidID",
			productID: "1",
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductResponse{
					ID:    1,
					Name:  "Test Product",
					Price: "100",
				}
				m.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Error_InvalidID",
			productID:      "invalid",
			mockSetup:      func(m *mocks.MockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Error_ProductNotFound",
			productID: "999",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().GetProductByID(mock.Anything, int64(999)).Return(nil, errors2.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Error_ServiceFailure",
			productID: "1",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().GetProductByID(mock.Anything, int64(1)).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.GET("/products/:id", handler.GetProductByID)

			req := httptest.NewRequest("GET", "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse && w.Code == http.StatusOK {
				var response dto.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Test Product", response.Name)
			}
		})
	}
}

func TestProductHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "Success_ValidRequest",
			requestBody: dto.CreateProductRequest{
				Name:        "New Product",
				Price:       decimal.NewFromInt(150),
				Stock:       10,
				CategoryIDs: []int64{1, 2},
			},
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductResponse{
					ID:    1,
					Name:  "New Product",
					Price: "150",
					Stock: 10,
				}
				m.EXPECT().CreateProduct(mock.Anything, mock.AnythingOfType("*dto.CreateProductRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "Error_InvalidJSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error_ServiceFailure",
			requestBody: dto.CreateProductRequest{
				Name:        "New Product",
				Price:       decimal.NewFromInt(150),
				Stock:       10,
				CategoryIDs: []int64{1, 2},
			},
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().CreateProduct(mock.Anything, mock.AnythingOfType("*dto.CreateProductRequest")).Return(nil, errors2.ErrInvalidProductData)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.POST("/products", handler.CreateProduct)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("POST", "/products", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse && w.Code == http.StatusCreated {
				var response dto.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "New Product", response.Name)
			}
		})
	}
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	name := "Updated Product"
	tests := []struct {
		name           string
		productID      string
		requestBody    dto.UpdateProductRequest
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:      "Success_ValidUpdate",
			productID: "1",
			requestBody: dto.UpdateProductRequest{
				Name: &name,
			},
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductResponse{
					ID:   1,
					Name: "Updated Product",
				}
				m.EXPECT().UpdateProduct(mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateProductRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:      "Error_InvalidID",
			productID: "invalid",
			requestBody: dto.UpdateProductRequest{
				Name: &name,
			},
			mockSetup:      func(m *mocks.MockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Error_ProductNotFound",
			productID: "999",
			requestBody: dto.UpdateProductRequest{
				Name: &name,
			},
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().UpdateProduct(mock.Anything, int64(999), mock.AnythingOfType("*dto.UpdateProductRequest")).Return(nil, errors2.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.PUT("/products/:id", handler.UpdateProduct)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("PUT", "/products/"+tt.productID, &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse && w.Code == http.StatusOK {
				var response dto.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Updated Product", response.Name)
			}
		})
	}
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
	}{
		{
			name:      "Success_ValidDelete",
			productID: "1",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().DeleteProduct(mock.Anything, int64(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Error_InvalidID",
			productID:      "invalid",
			mockSetup:      func(m *mocks.MockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Error_ProductNotFound",
			productID: "999",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().DeleteProduct(mock.Anything, int64(999)).Return(errors2.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.DELETE("/products/:id", handler.DeleteProduct)

			req := httptest.NewRequest("DELETE", "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_GetProductsByCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		queryParams    string
		mockSetup      func(*mocks.MockProductService)
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:        "Success_ValidCategory",
			categoryID:  "5",
			queryParams: "",
			mockSetup: func(m *mocks.MockProductService) {
				response := &dto.ProductCatalogResponse{
					Products: []dto.ProductCatalogItem{
						{ID: 1, Name: "Product in Category", Price: "100"},
					},
					Total: 1,
					Page:  1,
					Limit: 20,
				}
				m.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return *f.CategoryID == 5
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Error_InvalidCategoryID",
			categoryID:     "invalid",
			queryParams:    "",
			mockSetup:      func(m *mocks.MockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Error_ServiceFailure",
			categoryID:  "5",
			queryParams: "",
			mockSetup: func(m *mocks.MockProductService) {
				m.EXPECT().GetProducts(mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockProductService(t)
			tt.mockSetup(mockService)

			handler := product.NewProductHandler(mockService, zap.NewNop())
			router := setupRouter()
			router.GET("/categories/:id/products", handler.GetProductsByCategory)

			req := httptest.NewRequest("GET", "/categories/"+tt.categoryID+"/products"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse && w.Code == http.StatusOK {
				var response dto.ProductCatalogResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}

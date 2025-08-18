package product_test

import (
	"context"
	"goshop/internal/domain/errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/dto"
	"goshop/internal/service/product"
	"goshop/internal/service/product/mocks"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id int64) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error)
}

// -------------------- Общие хелперы --------------------

func buildService(t *testing.T) (*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache, ProductService) {
	productRepo := mocks.NewMockProductRepository(t)
	categoryRepo := mocks.NewMockCategoryRepository(t)
	productCache := mocks.NewMockProductCache(t)

	svcImpl := product.NewProductService(productRepo, categoryRepo, productCache, zap.NewNop())
	var svc ProductService = svcImpl
	return productRepo, categoryRepo, productCache, svc
}

// Cache: промах при получении одного товара
func expectCacheGetProductMiss(cache *mocks.MockProductCache, id int64) {
	cache.EXPECT().
		GetProduct(mock.Anything, id).
		Return(nil, nil)
}

// Cache: запись одного товара (после create/обновления сущности в БД)
func expectCacheSetProduct(cache *mocks.MockProductCache) {
	cache.EXPECT().
		SetProduct(
			mock.Anything,
			mock.AnythingOfType("*dto.ProductResponse"),
			mock.AnythingOfType("time.Duration"),
		).
		Return(nil)
}

// Cache: инвалидация кэша по товару (после update/delete)
func expectCacheInvalidateProduct(cache *mocks.MockProductCache, id int64) {
	cache.EXPECT().
		InvalidateProduct(mock.Anything, id).
		Return(nil)
}

// Cache: промах при получении списка по фильтрам
func expectCacheGetProductsMiss(cache *mocks.MockProductCache, page, limit int) {
	cache.EXPECT().
		GetProductsWithFilters(
			mock.Anything,
			mock.MatchedBy(func(f types.ProductFilters) bool {
				return f.Page == page && f.Limit == limit
			}),
		).
		Return(nil, nil)
}

// Cache: запись списка по фильтрам
func expectCacheSetProducts(cache *mocks.MockProductCache, page, limit int) {
	cache.EXPECT().
		SetProductsWithFilters(
			mock.Anything,
			mock.MatchedBy(func(f types.ProductFilters) bool {
				return f.Page == page && f.Limit == limit
			}),
			mock.AnythingOfType("*dto.ProductCatalogResponse"),
			mock.AnythingOfType("time.Duration"),
		).
		Return(nil)
}

// -------------------- CreateProduct --------------------

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		request        *dto.CreateProductRequest
		mockSetup      func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache)
		expectedError  error
		validateResult func(*testing.T, *dto.ProductResponse)
	}{
		{
			name: "Success_ValidProduct",
			request: &dto.CreateProductRequest{
				Name:        "iPhone 15",
				Description: stringPtr("Latest iPhone model"),
				Price:       decimal.NewFromFloat(99999.99),
				Stock:       50,
				CategoryIDs: []int64{1, 2},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{1, 2}).Return(true, nil)

				productRepo.EXPECT().CreateProduct(mock.Anything, mock.AnythingOfType("*entities.Product")).
					Run(func(ctx context.Context, p *entities.Product) {
						p.ID = 123
					}).Return(nil)

				productRepo.EXPECT().AddProductToCategories(mock.Anything, int64(123), []int64{1, 2}).Return(nil)

				categories := []*entities.Category{
					{ID: 1, UUID: uuid.New(), Name: "Electronics", Description: stringPtr("Electronic devices")},
					{ID: 2, UUID: uuid.New(), Name: "Smartphones", Description: stringPtr("Mobile phones")},
				}
				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return(categories, nil)

				// сервис кладёт созданный товар в кэш
				expectCacheSetProduct(cache)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *dto.ProductResponse) {
				assert.Equal(t, int64(123), result.ID)
				assert.Equal(t, "iPhone 15", result.Name)
				assert.Equal(t, "Latest iPhone model", *result.Description)
				assert.Equal(t, "99999.99", result.Price)
				assert.Equal(t, 50, result.Stock)
				assert.Len(t, result.Categories, 2)
			},
		},
		{
			name: "Error_EmptyName",
			request: &dto.CreateProductRequest{
				Name:        "",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidProductData,
		},
		{
			name: "Error_WhitespaceOnlyName",
			request: &dto.CreateProductRequest{
				Name:        "   ",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidProductData,
		},
		{
			name: "Error_ZeroPrice",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.Zero,
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidPrice,
		},
		{
			name: "Error_NegativePrice",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(-100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidPrice,
		},
		{
			name: "Error_PriceTooHigh",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.RequireFromString("9999999999.99"),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidPrice,
		},
		{
			name: "Error_PriceTooManyDecimals",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.RequireFromString("100.999"),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidPrice,
		},
		{
			name: "Error_ZeroStock",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       0,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidStock,
		},
		{
			name: "Error_NegativeStock",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       -5,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidStock,
		},
		{
			name: "Error_EmptyCategoryIDs",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name: "Error_DescriptionTooLong",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Description: stringPtr(generateLongString(1001)),
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidProductData,
		},
		{
			name: "Error_CategoriesNotFound",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{999},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{999}).Return(false, nil)
			},
			expectedError: errors.ErrCategoryNotFound,
		},
		{
			name: "Error_CategoryRepositoryError",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{1}).Return(false, assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "Error_ProductRepositoryCreateError",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{1}).Return(true, nil)
				productRepo.EXPECT().CreateProduct(mock.Anything, mock.AnythingOfType("*entities.Product")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "Error_AddCategoriesToProductError",
			request: &dto.CreateProductRequest{
				Name:        "Product",
				Price:       decimal.NewFromFloat(100.00),
				Stock:       10,
				CategoryIDs: []int64{1},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{1}).Return(true, nil)
				productRepo.EXPECT().CreateProduct(mock.Anything, mock.AnythingOfType("*entities.Product")).
					Run(func(ctx context.Context, p *entities.Product) { p.ID = 123 }).
					Return(nil)
				productRepo.EXPECT().AddProductToCategories(mock.Anything, int64(123), []int64{1}).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo, categoryRepo, productCache, svc := buildService(t)
			tt.mockSetup(productRepo, categoryRepo, productCache)

			result, err := svc.CreateProduct(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// -------------------- GetProductByID --------------------

func TestProductService_GetProductByID(t *testing.T) {
	tests := []struct {
		name           string
		productID      int64
		mockSetup      func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache)
		expectedError  error
		validateResult func(*testing.T, *dto.ProductResponse)
	}{
		{
			name:      "Success_ValidProduct",
			productID: 123,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, mockCache *mocks.MockProductCache) {
				// сначала кэш
				expectCacheGetProductMiss(mockCache, 123)

				product := &entities.Product{
					ID:          123,
					UUID:        uuid.New(),
					Name:        "iPhone 15",
					Description: stringPtr("Latest iPhone"),
					Price:       decimal.NewFromFloat(99999.99),
					Stock:       50,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(product, nil)

				categories := []*entities.Category{
					{ID: 1, UUID: uuid.New(), Name: "Electronics"},
				}
				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return(categories, nil)

				// сервис кладёт полученный товар в кэш
				expectCacheSetProduct(mockCache)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *dto.ProductResponse) {
				assert.Equal(t, int64(123), result.ID)
				assert.Equal(t, "iPhone 15", result.Name)
				assert.Len(t, result.Categories, 1)
			},
		},
		{
			name:          "Error_InvalidID",
			productID:     0,
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name:          "Error_NegativeID",
			productID:     -1,
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name:      "Error_ProductNotFound",
			productID: 999,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				expectCacheGetProductMiss(cache, 999)
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(999)).Return(nil, errors.ErrProductNotFound)
			},
			expectedError: errors.ErrProductNotFound,
		},
		{
			name:      "Error_GetCategoriesError",
			productID: 123,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				expectCacheGetProductMiss(cache, 123)
				product := &entities.Product{ID: 123, Name: "Product"}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(product, nil)
				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo, categoryRepo, productCache, svc := buildService(t)
			tt.mockSetup(productRepo, categoryRepo, productCache)

			result, err := svc.GetProductByID(context.Background(), tt.productID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// -------------------- UpdateProduct --------------------

func TestProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name           string
		productID      int64
		request        *dto.UpdateProductRequest
		mockSetup      func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache)
		expectedError  error
		validateResult func(*testing.T, *dto.ProductResponse)
	}{
		{
			name:      "Success_UpdateAllFields",
			productID: 123,
			request: &dto.UpdateProductRequest{
				Name:        stringPtr("Updated iPhone"),
				Description: stringPtr("Updated description"),
				Price:       decimalPtr(decimal.NewFromFloat(89999.99)),
				Stock:       intPtr(75),
				CategoryIDs: []int64{2, 3},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				existingProduct := &entities.Product{
					ID:          123,
					UUID:        uuid.New(),
					Name:        "iPhone 15",
					Description: stringPtr("Old description"),
					Price:       decimal.NewFromFloat(99999.99),
					Stock:       50,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(existingProduct, nil)

				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{2, 3}).Return(true, nil)

				productRepo.EXPECT().UpdateProduct(mock.Anything, mock.AnythingOfType("*entities.Product")).Return(nil)
				productRepo.EXPECT().RemoveProductFromCategories(mock.Anything, int64(123)).Return(nil)
				productRepo.EXPECT().AddProductToCategories(mock.Anything, int64(123), []int64{2, 3}).Return(nil)

				categories := []*entities.Category{
					{ID: 2, UUID: uuid.New(), Name: "Tablets"},
					{ID: 3, UUID: uuid.New(), Name: "Apple"},
				}
				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return(categories, nil)

				// инвалидация кэша по товару
				expectCacheInvalidateProduct(cache, 123)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *dto.ProductResponse) {
				assert.Equal(t, "Updated iPhone", result.Name)
				assert.Equal(t, "Updated description", *result.Description)
				assert.Equal(t, "89999.99", result.Price)
				assert.Equal(t, 75, result.Stock)
				assert.Len(t, result.Categories, 2)
			},
		},
		{
			name:      "Success_UpdateOnlyName",
			productID: 123,
			request: &dto.UpdateProductRequest{
				Name: stringPtr("New Name"),
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				existingProduct := &entities.Product{ID: 123, Name: "Old Name"}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(existingProduct, nil)
				productRepo.EXPECT().UpdateProduct(mock.Anything, mock.AnythingOfType("*entities.Product")).Return(nil)

				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return([]*entities.Category{}, nil)

				expectCacheInvalidateProduct(cache, 123)
			},
			expectedError: nil,
		},
		{
			name:      "Success_UpdateOnlyCategories",
			productID: 123,
			request: &dto.UpdateProductRequest{
				CategoryIDs: []int64{4, 5},
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				existingProduct := &entities.Product{ID: 123}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(existingProduct, nil)

				categoryRepo.EXPECT().CheckCategoriesExist(mock.Anything, []int64{4, 5}).Return(true, nil)
				productRepo.EXPECT().RemoveProductFromCategories(mock.Anything, int64(123)).Return(nil)
				productRepo.EXPECT().AddProductToCategories(mock.Anything, int64(123), []int64{4, 5}).Return(nil)

				productRepo.EXPECT().GetProductCategories(mock.Anything, int64(123)).Return([]*entities.Category{}, nil)

				expectCacheInvalidateProduct(cache, 123)
			},
			expectedError: nil,
		},
		{
			name:          "Error_InvalidID",
			productID:     0,
			request:       &dto.UpdateProductRequest{},
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name:      "Error_EmptyName",
			productID: 123,
			request:   &dto.UpdateProductRequest{Name: stringPtr("")},
			mockSetup: func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {
			},
			expectedError: errors.ErrInvalidProductData,
		},
		{
			name:      "Error_NegativePrice",
			productID: 123,
			request:   &dto.UpdateProductRequest{Price: decimalPtr(decimal.NewFromFloat(-100.00))},
			mockSetup: func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {
			},
			expectedError: errors.ErrInvalidPrice,
		},
		{
			name:      "Error_NegativeStock",
			productID: 123,
			request:   &dto.UpdateProductRequest{Stock: intPtr(-5)},
			mockSetup: func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {
			},
			expectedError: errors.ErrInvalidStock,
		},
		{
			name:      "Error_NoChanges",
			productID: 123,
			request:   &dto.UpdateProductRequest{},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				existingProduct := &entities.Product{ID: 123}
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(123)).Return(existingProduct, nil)
				// кеш не трогаем, т.к. ошибка до изменений
			},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name:      "Error_ProductNotFound",
			productID: 999,
			request:   &dto.UpdateProductRequest{Name: stringPtr("New Name")},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				productRepo.EXPECT().GetProductByID(mock.Anything, int64(999)).Return(nil, errors.ErrProductNotFound)
			},
			expectedError: errors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo, categoryRepo, productCache, svc := buildService(t)
			tt.mockSetup(productRepo, categoryRepo, productCache)

			result, err := svc.UpdateProduct(context.Background(), tt.productID, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// -------------------- DeleteProduct --------------------

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     int64
		mockSetup     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache)
		expectedError error
	}{
		{
			name:      "Success_ValidDelete",
			productID: 123,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				productRepo.EXPECT().RemoveProductFromCategories(mock.Anything, int64(123)).Return(nil)
				productRepo.EXPECT().DeleteProduct(mock.Anything, int64(123)).Return(nil)

				expectCacheInvalidateProduct(cache, 123)
			},
			expectedError: nil,
		},
		{
			name:          "Error_InvalidID",
			productID:     0,
			mockSetup:     func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache) {},
			expectedError: errors.ErrInvalidInput,
		},
		{
			name:      "Error_RemoveCategoriesError",
			productID: 123,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				productRepo.EXPECT().RemoveProductFromCategories(mock.Anything, int64(123)).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name:      "Error_DeleteProductError",
			productID: 123,
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				productRepo.EXPECT().RemoveProductFromCategories(mock.Anything, int64(123)).Return(nil)
				productRepo.EXPECT().DeleteProduct(mock.Anything, int64(123)).Return(errors.ErrProductNotFound)
			},
			expectedError: errors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo, categoryRepo, productCache, svc := buildService(t)
			tt.mockSetup(productRepo, categoryRepo, productCache)

			err := svc.DeleteProduct(context.Background(), tt.productID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -------------------- GetProducts (каталог) --------------------

func TestProductService_GetProducts(t *testing.T) {
	tests := []struct {
		name           string
		filters        types.ProductFilters
		mockSetup      func(*mocks.MockProductRepository, *mocks.MockCategoryRepository, *mocks.MockProductCache)
		expectedError  error
		validateResult func(*testing.T, *dto.ProductCatalogResponse)
	}{
		{
			name: "Success_DefaultPagination",
			filters: types.ProductFilters{
				Page:  0,
				Limit: 0,
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				// сервис нормализует в Page=1, Limit=20
				expectCacheGetProductsMiss(cache, 1, 20)

				products := []*entities.Product{
					{ID: 1, UUID: uuid.New(), Name: "Product 1", Price: decimal.NewFromFloat(100.00), Stock: 10},
					{ID: 2, UUID: uuid.New(), Name: "Product 2", Price: decimal.NewFromFloat(200.00), Stock: 20},
				}
				productRepo.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return(products, 50, nil)

				expectCacheSetProducts(cache, 1, 20)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *dto.ProductCatalogResponse) {
				assert.Len(t, result.Products, 2)
				assert.Equal(t, 50, result.Total)
				assert.Equal(t, 1, result.Page)
				assert.Equal(t, 20, result.Limit)
			},
		},
		{
			name: "Success_CustomPagination",
			filters: types.ProductFilters{
				Page:  2,
				Limit: 10,
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				expectCacheGetProductsMiss(cache, 2, 10)

				products := []*entities.Product{}
				productRepo.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return f.Page == 2 && f.Limit == 10
				})).Return(products, 0, nil)

				expectCacheSetProducts(cache, 2, 10)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *dto.ProductCatalogResponse) {
				assert.Len(t, result.Products, 0)
				assert.Equal(t, 0, result.Total)
				assert.Equal(t, 2, result.Page)
				assert.Equal(t, 10, result.Limit)
			},
		},
		{
			name: "Success_LimitOverMaximum",
			filters: types.ProductFilters{
				Page:  1,
				Limit: 150,
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				// лимит капится до 20
				expectCacheGetProductsMiss(cache, 1, 20)

				products := []*entities.Product{}
				productRepo.EXPECT().GetProducts(mock.Anything, mock.MatchedBy(func(f types.ProductFilters) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return(products, 0, nil)

				expectCacheSetProducts(cache, 1, 20)
			},
			expectedError: nil,
		},
		{
			name: "Error_RepositoryError",
			filters: types.ProductFilters{
				Page:  1,
				Limit: 20,
			},
			mockSetup: func(productRepo *mocks.MockProductRepository, categoryRepo *mocks.MockCategoryRepository, cache *mocks.MockProductCache) {
				expectCacheGetProductsMiss(cache, 1, 20)
				productRepo.EXPECT().GetProducts(mock.Anything, mock.AnythingOfType("types.ProductFilters")).Return(nil, 0, assert.AnError)
				// SetProductsWithFilters НЕ ожидаем, т.к. репозиторий вернул ошибку
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo, categoryRepo, productCache, svc := buildService(t)
			tt.mockSetup(productRepo, categoryRepo, productCache)

			result, err := svc.GetProducts(context.Background(), tt.filters)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// -------------------- утилиты --------------------

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int          { return &i }
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

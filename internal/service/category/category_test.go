package category_test

import (
	"context"
	"errors"
	errors2 "goshop/internal/domain/errors"
	"testing"
	"time"

	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/service/category"
	"goshop/internal/service/category/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func stringPtr(s string) *string {
	return &s
}

func TestCategoryService_GetAllCategories(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mocks.MockCategoryRepository)
		expectedCount int
		wantErr       bool
		errMsg        string
	}{
		{
			name: "Success_ReturnsAllCategories",
			mockSetup: func(m *mocks.MockCategoryRepository) {
				categories := []*entities.CategoryWithCount{
					{
						Category: entities.Category{
							ID:   1,
							Name: "Electronics",
						},
						ProductCount: 5,
					},
					{
						Category: entities.Category{
							ID:   2,
							Name: "Books",
						},
						ProductCount: 3,
					},
				}
				m.EXPECT().GetAllCategories(mock.Anything).Return(categories, nil)
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "Success_EmptyCategories",
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().GetAllCategories(mock.Anything).Return([]*entities.CategoryWithCount{}, nil)
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "Error_RepositoryFails",
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().GetAllCategories(mock.Anything).Return([]*entities.CategoryWithCount(nil), errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to get categories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCategoryRepository(t)
			tt.mockSetup(mockRepo)

			service := category.NewCategoryService(mockRepo, zap.NewNop())
			ctx := context.Background()

			result, err := service.GetAllCategories(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*mocks.MockCategoryRepository)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success_ValidID",
			id:   1,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				categoryWithCount := &entities.CategoryWithCount{
					Category: entities.Category{
						ID:   1,
						Name: "Electronics",
					},
					ProductCount: 5,
				}
				m.EXPECT().GetCategoryByID(mock.Anything, int64(1)).Return(categoryWithCount, nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidID_Zero",
			id:        0,
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_InvalidID_Negative",
			id:        -1,
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name: "Error_CategoryNotFound",
			id:   999,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().GetCategoryByID(mock.Anything, int64(999)).Return(nil, errors2.ErrCategoryNotFound)
			},
			wantErr: true,
			errType: errors2.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCategoryRepository(t)
			tt.mockSetup(mockRepo)

			service := category.NewCategoryService(mockRepo, zap.NewNop())
			ctx := context.Background()

			result, err := service.GetCategoryByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestCategoryService_CreateCategory(t *testing.T) {
	tests := []struct {
		name      string
		request   *dto.CreateCategoryRequest
		mockSetup func(*mocks.MockCategoryRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Success_ValidRequest",
			request: &dto.CreateCategoryRequest{
				Name:        "Electronics",
				Description: stringPtr("Electronic devices and gadgets"),
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().CreateCategory(mock.Anything, mock.MatchedBy(func(cat *entities.Category) bool {
					return cat.Name == "Electronics" && *cat.Description == "Electronic devices and gadgets"
				})).Run(func(ctx context.Context, cat *entities.Category) {
					cat.ID = 1
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Error_RepositoryFails",
			request: &dto.CreateCategoryRequest{
				Name:        "Books",
				Description: stringPtr("All kinds of books"),
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().CreateCategory(mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to create category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCategoryRepository(t)
			tt.mockSetup(mockRepo)

			service := category.NewCategoryService(mockRepo, zap.NewNop())
			ctx := context.Background()

			result, err := service.CreateCategory(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, *tt.request.Description, *result.Description)
				assert.NotZero(t, result.UUID)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			}
		})
	}
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	validCategory := &entities.Category{
		ID:          1,
		Name:        "Updated Electronics",
		Description: stringPtr("Updated description"),
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name      string
		category  *entities.Category
		mockSetup func(*mocks.MockCategoryRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:     "Success_ValidCategory",
			category: validCategory,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().UpdateCategory(mock.Anything, mock.MatchedBy(func(cat *entities.Category) bool {
					return cat.ID == 1 && cat.Name == "Updated Electronics"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Error_InvalidID_Zero",
			category: &entities.Category{
				ID:          0,
				Name:        "Test",
				Description: stringPtr("Test desc"),
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name: "Error_InvalidID_Negative",
			category: &entities.Category{
				ID:          -1,
				Name:        "Test",
				Description: stringPtr("Test desc"),
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_NilCategory",
			category:  nil,
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidCategoryData,
		},
		{
			name: "Error_EmptyName",
			category: &entities.Category{
				ID:          1,
				Name:        "",
				Description: stringPtr("Test desc"),
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidCategoryData,
		},
		{
			name: "Error_NilDescription",
			category: &entities.Category{
				ID:          1,
				Name:        "Test",
				Description: nil,
			},
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidCategoryData,
		},
		{
			name:     "Error_RepositoryFails",
			category: validCategory,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().UpdateCategory(mock.Anything, mock.Anything).Return(errors2.ErrCategoryNotFound)
			},
			wantErr: true,
			errType: errors2.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCategoryRepository(t)
			tt.mockSetup(mockRepo)

			service := category.NewCategoryService(mockRepo, zap.NewNop())
			ctx := context.Background()

			categoryDTO := CategoryEntityToDTO(tt.category, 0)
			categoryEntity := dtoToEntity(categoryDTO)

			err := service.UpdateCategory(ctx, categoryEntity)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.category.UpdatedAt.After(tt.category.CreatedAt))
			}
		})
	}
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*mocks.MockCategoryRepository)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success_ValidID",
			id:   1,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().DeleteCategory(mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Error_InvalidID_Zero",
			id:        0,
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name:      "Error_InvalidID_Negative",
			id:        -1,
			mockSetup: func(m *mocks.MockCategoryRepository) {},
			wantErr:   true,
			errType:   errors2.ErrInvalidInput,
		},
		{
			name: "Error_CategoryNotFound",
			id:   999,
			mockSetup: func(m *mocks.MockCategoryRepository) {
				m.EXPECT().DeleteCategory(mock.Anything, int64(999)).Return(errors2.ErrCategoryNotFound)
			},
			wantErr: true,
			errType: errors2.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCategoryRepository(t)
			tt.mockSetup(mockRepo)

			service := category.NewCategoryService(mockRepo, zap.NewNop())
			ctx := context.Background()

			err := service.DeleteCategory(ctx, tt.id)

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

// --- helpers ---

func CategoryEntityToDTO(entity *entities.Category, productCount int) *dto.CategoryResponse {
	if entity == nil {
		return nil
	}
	return &dto.CategoryResponse{
		ID:           entity.ID,
		UUID:         entity.UUID.String(),
		Name:         entity.Name,
		Description:  entity.Description,
		ProductCount: productCount,
	}
}

func dtoToEntity(d *dto.CategoryResponse) *entities.Category {
	if d == nil {
		return nil
	}
	return &entities.Category{
		ID:          d.ID,
		UUID:        uuid.MustParse(d.UUID),
		Name:        d.Name,
		Description: d.Description,
	}
}

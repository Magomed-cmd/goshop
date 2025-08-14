package category

import (
	"context"
	"fmt"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error)
	CreateCategory(ctx context.Context, category *entities.Category) error
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

type CategoryCache interface {
	GetCategory(ctx context.Context, categoryID int64) (*dto.CategoryResponse, error)
	SetCategory(ctx context.Context, category *dto.CategoryResponse, ttl time.Duration)
}

type CategoryService struct {
	categoryRepo  CategoryRepository
	categoryCache CategoryCache
	logger        *zap.Logger
}

func NewCategoryService(categoryRepo CategoryRepository, logger *zap.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error) {
	categories, err := s.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error) {

	if id <= 0 {
		return nil, domain_errors.ErrInvalidInput
	}

	if s.categoryCache != nil {
		if categoryCached, err := s.categoryCache.GetCategory(ctx, id); err != nil {
			s.logger.Warn(
				"failed to get category from cache",
				zap.Int64("category_id", id),
				zap.Error(err),
			)
		} else if categoryCached != nil {
			return categoryCached, nil
		}
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, domain_errors.ErrCategoryNotFound
	}

	resp := &dto.CategoryResponse{
		ID:           category.Category.ID,
		UUID:         category.Category.UUID.String(),
		Name:         category.Category.Name,
		Description:  category.Category.Description,
		ProductCount: int(category.ProductCount),
	}

	if s.categoryCache != nil {
		s.categoryCache.SetCategory(ctx, resp, 5*time.Minute)
	}

	return resp, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*entities.Category, error) {
	now := time.Now()

	category := &entities.Category{
		UUID:        uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.categoryRepo.CreateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *entities.Category) error {
	if category == nil {
		return domain_errors.ErrInvalidCategoryData
	}
	if category.ID <= 0 {
		return domain_errors.ErrInvalidInput
	}
	if category.Name == "" {
		return domain_errors.ErrInvalidCategoryData
	}
	if category.Description == nil {
		return domain_errors.ErrInvalidCategoryData
	}

	category.UpdatedAt = time.Now()

	categoryEntity := &entities.Category{
		ID:          category.ID,
		UUID:        category.UUID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	if err := s.categoryRepo.UpdateCategory(ctx, categoryEntity); err != nil {
		return err
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain_errors.ErrInvalidInput
	}
	if err := s.categoryRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	return nil
}

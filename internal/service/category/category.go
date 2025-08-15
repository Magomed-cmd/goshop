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

const (
	GetAllCategoriesTTL = 5 * time.Minute
)

type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error)
	CreateCategory(ctx context.Context, category *entities.Category) error
	UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

type CategoryCache interface {
	GetCategory(ctx context.Context, categoryID int64) (*dto.CategoryResponse, error)
	SetCategory(ctx context.Context, category *dto.CategoryResponse, ttl time.Duration) error
	GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error)
	SetAllCategories(ctx context.Context, response *dto.CategoriesListResponse, ttl time.Duration) error
	DeleteCategory(ctx context.Context, categoryID int64) error
	DeleteAllCategories(ctx context.Context) error
}

type CategoryService struct {
	categoryRepo  CategoryRepository
	categoryCache CategoryCache
	logger        *zap.Logger
}

func NewCategoryService(categoryRepo CategoryRepository, categoryCache CategoryCache, logger *zap.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo:  categoryRepo,
		categoryCache: categoryCache,
		logger:        logger,
	}
}

func (s *CategoryService) GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error) {

	categoriesCache, err := s.categoryCache.GetAllCategories(ctx)
	if err != nil {
		s.logger.Warn("Failed to get categories from cache", zap.Error(err))
		return nil, err
	} else if categoriesCache != nil {
		s.logger.Debug("Returning categories from cache")
		return categoriesCache, nil
	}

	categories, err := s.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	resp := &dto.CategoriesListResponse{
		Categories: make([]dto.CategoryResponse, 0, len(categories)),
	}

	for _, categoryEntity := range categories {
		resp.Categories = append(resp.Categories, dto.CategoryResponse{
			ID:           categoryEntity.ID,
			UUID:         categoryEntity.UUID.String(),
			Name:         categoryEntity.Name,
			Description:  categoryEntity.Description,
			CreatedAt:    categoryEntity.CreatedAt,
			UpdatedAt:    categoryEntity.UpdatedAt,
			ProductCount: int(categoryEntity.ProductCount),
		})
	}

	if err = s.categoryCache.SetAllCategories(ctx, resp, GetAllCategoriesTTL); err != nil {
		s.logger.Warn("failed to set all categories in cache", zap.Error(err))
	}

	return resp, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error) {

	if id <= 0 {
		return nil, domain_errors.ErrInvalidInput
	}

	if categoryCached, err := s.categoryCache.GetCategory(ctx, id); err != nil {
		s.logger.Warn(
			"failed to get category from cache",
			zap.Int64("category_id", id),
			zap.Error(err),
		)
	} else if categoryCached != nil {
		return categoryCached, nil
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

	err = s.categoryCache.SetCategory(ctx, resp, 5*time.Minute)
	if err != nil {
		s.logger.Error("failed to set category in cache", zap.Error(err))
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
	if err := s.categoryCache.DeleteAllCategories(ctx); err != nil {
		s.logger.Warn("failed to delete all categories from cache after create", zap.Error(err))
	}
	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	if category == nil {
		return nil, domain_errors.ErrInvalidCategoryData
	}
	if category.ID <= 0 {
		return nil, domain_errors.ErrInvalidInput
	}
	if category.Name == "" {
		return nil, domain_errors.ErrInvalidCategoryData
	}
	if category.Description == nil {
		return nil, domain_errors.ErrInvalidCategoryData
	}

	updated, err := s.categoryRepo.UpdateCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	if err := s.categoryCache.DeleteCategory(ctx, updated.ID); err != nil {
		s.logger.Warn("failed to delete category from cache after update",
			zap.Int64("category_id", updated.ID), zap.Error(err))
	}
	if err := s.categoryCache.DeleteAllCategories(ctx); err != nil {
		s.logger.Warn("failed to delete all categories from cache after update", zap.Error(err))
	}

	return updated, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain_errors.ErrInvalidInput
	}
	if err := s.categoryRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}

	if err := s.categoryCache.DeleteCategory(ctx, id); err != nil {
		s.logger.Error("failed to delete category from cache after update", zap.Int64("category_id", id), zap.Error(err))
	}
	if err := s.categoryCache.DeleteAllCategories(ctx); err != nil {
		s.logger.Error("failed to delete all categories from cache after update", zap.Error(err))
	}
	return nil
}

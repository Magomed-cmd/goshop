package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	cacheports "goshop/internal/core/ports/cache"
	"goshop/internal/core/ports/repositories"
)

type CategoryService struct {
	categoryRepo  repositories.CategoryRepository
	categoryCache cacheports.CategoryCache
	logger        *zap.Logger
}

func NewCategoryService(categoryRepo repositories.CategoryRepository, categoryCache cacheports.CategoryCache, logger *zap.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo:  categoryRepo,
		categoryCache: categoryCache,
		logger:        logger,
	}
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error) {
	return s.categoryRepo.GetAllCategories(ctx)
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error) {

	if id <= 0 {
		return nil, errors.ErrInvalidInput
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.ErrCategoryNotFound
	}
	return category, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, name string, description *string) (*entities.Category, error) {
	if name == "" {
		return nil, errors.ErrInvalidInput
	}

	now := time.Now()

	category := &entities.Category{
		UUID:        uuid.New(),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.categoryRepo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	if err := s.categoryCache.DeleteAllCategories(ctx); err != nil {
		s.logger.Warn("failed to delete all categories from cache after create", zap.Error(err))
	}

	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id int64, name, description *string) (*entities.CategoryWithCount, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidInput
	}
	if name == nil && description == nil {
		return nil, errors.ErrInvalidCategoryData
	}

	current, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.ErrCategoryNotFound
	}

	currentName := current.Name
	if name != nil {
		currentName = *name
	}
	if currentName == "" {
		return nil, errors.ErrInvalidCategoryData
	}

	currentDescription := current.Description
	if description != nil {
		currentDescription = description
	}
	if currentDescription == nil {
		return nil, errors.ErrInvalidCategoryData
	}

	entity := &entities.Category{
		ID:          current.ID,
		UUID:        current.UUID,
		Name:        currentName,
		Description: currentDescription,
		CreatedAt:   current.CreatedAt,
		UpdatedAt:   current.UpdatedAt,
	}

	updated, err := s.categoryRepo.UpdateCategory(ctx, entity)
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

	return &entities.CategoryWithCount{Category: *updated, ProductCount: current.ProductCount}, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.ErrInvalidInput
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

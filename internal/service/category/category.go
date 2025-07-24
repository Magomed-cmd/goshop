package category

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"time"
)

type CategoryRepositoryInterface interface {
	GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error)
	CreateCategory(ctx context.Context, category *entities.Category) error
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

type CategoryService struct {
	categoryRepo CategoryRepositoryInterface
}

func NewCategoryService(categoryRepo CategoryRepositoryInterface) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error) {
	categories, err := s.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error) {
	if id <= 0 {
		return nil, domain_errors.ErrInvalidInput
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
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

	err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
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

	err := s.categoryRepo.UpdateCategory(ctx, category)
	if err != nil {
		return err
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain_errors.ErrInvalidInput
	}

	err := s.categoryRepo.DeleteCategory(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

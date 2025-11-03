package repositories

import (
    "context"

    "goshop/internal/core/domain/entities"
)

type CategoryRepository interface {
    GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
    GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error)
    CreateCategory(ctx context.Context, category *entities.Category) error
    UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error)
    DeleteCategory(ctx context.Context, id int64) error
    CheckCategoriesExist(ctx context.Context, categoryIDs []int64) (bool, error)
}

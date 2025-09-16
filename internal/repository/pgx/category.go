package pgx

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	errors2 "goshop/internal/domain/errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type CategoryRepository struct {
	db   *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewCategoryRepository(db *pgxpool.Pool, logger *zap.Logger) *CategoryRepository {
	return &CategoryRepository{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *CategoryRepository) GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error) {
	query := `SELECT c.id, c.uuid, c.name, c.description, c.created_at, c.updated_at, count(pc.*)
             FROM categories c
               LEFT JOIN product_categories pc ON pc.category_id = c.id
             GROUP BY c.id, c.uuid, c.name, c.description, c.created_at, c.updated_at
             ORDER BY c.name;`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var categories []*entities.CategoryWithCount

	for rows.Next() {
		var category entities.CategoryWithCount

		if err := rows.Scan(
			&category.Category.ID,
			&category.Category.UUID,
			&category.Category.Name,
			&category.Category.Description,
			&category.Category.CreatedAt,
			&category.Category.UpdatedAt,
			&category.ProductCount,
		); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryWithCount, error) {
	query := `
        SELECT 
            c.id, c.uuid, c.name, c.description, c.created_at, c.updated_at,
            COALESCE(COUNT(pc.category_id), 0) as product_count
        FROM categories c
        LEFT JOIN product_categories pc ON pc.category_id = c.id
        WHERE c.id = $1
        GROUP BY c.id, c.uuid, c.name, c.description, c.created_at, c.updated_at`

	var category entities.CategoryWithCount
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.Category.ID,
		&category.Category.UUID,
		&category.Category.Name,
		&category.Category.Description,
		&category.Category.CreatedAt,
		&category.Category.UpdatedAt,
		&category.ProductCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors2.ErrCategoryNotFound
		}
		return nil, err
	}

	return &category, nil
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, category *entities.Category) error {
	query := `
    INSERT INTO categories (uuid, name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	err := r.db.QueryRow(ctx, query,
		category.UUID,
		category.Name,
		category.Description,
		category.CreatedAt,
		category.UpdatedAt).Scan(&category.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	query := r.psql.Update("categories")
	paramsCnt := 0

	if category.Name != "" {
		query = query.Set("name", category.Name)
		paramsCnt++
	}
	if category.Description != nil {
		query = query.Set("description", category.Description)
		paramsCnt++
	}
	if paramsCnt == 0 {
		return nil, errors2.ErrInvalidInput
	}

	now := time.Now()
	query = query.
		Set("updated_at", now).
		Where(squirrel.Eq{"id": category.ID}).
		Suffix("RETURNING id, uuid, name, description, created_at, updated_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var updated entities.Category

	if err := r.db.QueryRow(ctx, sql, args...).Scan(
		&updated.ID,
		&updated.UUID,
		&updated.Name,
		&updated.Description,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors2.ErrCategoryNotFound
		}
		return nil, err
	}

	return &updated, nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	query := "DELETE FROM categories WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors2.ErrCategoryNotFound
	}

	return nil
}

func (r *CategoryRepository) CheckCategoriesExist(ctx context.Context, categoryIDs []int64) (bool, error) {
	query := `SELECT COUNT(*) FROM categories WHERE id = ANY($1)`

	var count int
	err := r.db.QueryRow(ctx, query, categoryIDs).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == len(categoryIDs), nil
}

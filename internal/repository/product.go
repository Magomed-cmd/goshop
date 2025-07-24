package repository

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
)

type ProductRepository struct {
	db   *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *entities.Product) error {

	query := `INSERT INTO products (uuid, name, description, price, stock, created_at, updated_at, is_active) 
			  values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err := r.db.QueryRow(ctx, query,
		product.UUID,
		product.Name,
		product.Description,
		product.Price, product.Stock,
		product.CreatedAt,
		product.UpdatedAt,
		product.IsActive,
	).Scan(&product.ID)
	if err != nil {

		return err
	}
	return nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	query := `SELECT * FROM products WHERE id = $1`
	var product entities.Product

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.UUID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *entities.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, updated_at = $5, is_active = $6 WHERE id = $7`
	result, err := r.db.Exec(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.UpdatedAt,
		product.IsActive,
		product.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain_errors.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain_errors.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error) {

	baseQuery := r.psql.Select().From("products p")

	if filters.CategoryID != nil {
		baseQuery = baseQuery.Join("product_categories pc on p.id = pc.product_id").
			Where(squirrel.Eq{"pc.category_id": *filters.CategoryID})
	}
	if filters.MinPrice != nil {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"p.price": *filters.MinPrice})
	}
	if filters.MaxPrice != nil {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"p.price": *filters.MaxPrice})
	}

	countSql, countArgs, err := baseQuery.Columns("COUNT(*)").ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := baseQuery.Columns("p.id", "p.uuid", "p.name", "p.description", "p.price", "p.stock", "p.is_active", "p.created_at", "p.updated_at")

	if filters.SortBy != nil {
		sortOrder := "ASC"
		if filters.SortOrder != nil {
			sortOrder = *filters.SortOrder
		}
		dataQuery = dataQuery.OrderBy("p." + *filters.SortBy + " " + sortOrder)
	} else {
		dataQuery = dataQuery.OrderBy("p.created_at DESC")
	}

	offset := (filters.Page - 1) * filters.Limit
	dataQuery = dataQuery.Limit(uint64(filters.Limit)).Offset(uint64(offset))

	dataSql, dataArgs, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, dataSql, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		var product entities.Product
		if err := rows.Scan(&product.ID, &product.UUID, &product.Name, &product.Description,
			&product.Price, &product.Stock, &product.IsActive,
			&product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, &product)
	}

	return products, total, rows.Err()
}

func (r *ProductRepository) AddProductToCategories(ctx context.Context, productID int64, categoryIDs []int64) error {

	query := r.psql.Insert("product_categories").
		Columns("product_id", "category_id").
		Suffix("ON CONFLICT DO NOTHING")

	for _, categoryID := range categoryIDs {
		query = query.Values(productID, categoryID)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) RemoveProductFromCategories(ctx context.Context, productID int64) error {
	query := `DELETE FROM product_categories WHERE product_id = $1`

	_, err := r.db.Exec(ctx, query, productID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ProductRepository) GetProductCategories(ctx context.Context, productID int64) ([]*entities.Category, error) {

	query := `SELECT c.* FROM product_categories pc 
           JOIN categories c on c.id = pc.category_id
           WHERE pc.product_id = $1`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []*entities.Category
	for rows.Next() {
		var category entities.Category
		if err := rows.Scan(
			&category.ID,
			&category.UUID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return categories, nil
}

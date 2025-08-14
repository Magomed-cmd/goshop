package repository

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
)

type ProductRepository struct {
	db     *pgxpool.Pool
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewProductRepository(db *pgxpool.Pool, logger *zap.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *entities.Product) error {
	r.logger.Debug("Creating product in database", zap.String("product_name", product.Name))

	query := `INSERT INTO products (uuid, name, description, price, stock, created_at, updated_at) 
			  values ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := r.db.QueryRow(ctx, query,
		product.UUID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID)

	if err != nil {
		r.logger.Error("Failed to create product in database",
			zap.Error(err),
			zap.String("product_name", product.Name),
			zap.String("product_uuid", product.UUID.String()))
		return err
	}

	r.logger.Info("Product created successfully",
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name))

	return nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	r.logger.Debug("Getting product by ID from database", zap.Int64("product_id", id))

	query := `SELECT id, uuid, name, description, price, stock, created_at, updated_at FROM products WHERE id = $1`
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain_errors.ErrProductNotFound
		}
		r.logger.Error("Failed to get product by ID", zap.Error(err), zap.Int64("product_id", id))
		return nil, err
	}

	r.logger.Debug("Product retrieved successfully",
		zap.Int64("product_id", id),
		zap.String("product_name", product.Name))

	return &product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *entities.Product) error {
	r.logger.Debug("Updating product in database",
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name))

	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, updated_at = $5 WHERE id = $6`
	result, err := r.db.Exec(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update product",
			zap.Error(err),
			zap.Int64("product_id", product.ID),
			zap.String("product_name", product.Name))
		return err
	}

	if result.RowsAffected() == 0 {
		r.logger.Warn("Product not found for update", zap.Int64("product_id", product.ID))
		return domain_errors.ErrProductNotFound
	}

	r.logger.Info("Product updated successfully",
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name))

	return nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	r.logger.Debug("Deleting product from database", zap.Int64("product_id", id))

	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		r.logger.Error("Failed to delete product", zap.Error(err), zap.Int64("product_id", id))
		return err
	}

	if result.RowsAffected() == 0 {
		r.logger.Warn("Product not found for deletion", zap.Int64("product_id", id))
		return domain_errors.ErrProductNotFound
	}

	r.logger.Info("Product deleted successfully", zap.Int64("product_id", id))
	return nil
}

func (r *ProductRepository) GetProducts(ctx context.Context, filters types.ProductFilters) ([]*entities.Product, int, error) {
	r.logger.Debug("Getting products with filters", zap.Any("filters", filters))

	baseQuery := r.psql.Select().From("products p")

	if filters.CategoryID != nil {
		r.logger.Debug("Applying category filter", zap.Int64("category_id", *filters.CategoryID))
		baseQuery = baseQuery.Join("product_categories pc on p.id = pc.product_id").
			Where(squirrel.Eq{"pc.category_id": *filters.CategoryID})
	}
	if filters.MinPrice != nil {
		r.logger.Debug("Applying min price filter", zap.Any("min_price", *filters.MinPrice))
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"p.price": *filters.MinPrice})
	}
	if filters.MaxPrice != nil {
		r.logger.Debug("Applying max price filter", zap.Any("max_price", *filters.MaxPrice))
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"p.price": *filters.MaxPrice})
	}

	countSql, countArgs, err := baseQuery.Columns("COUNT(*)").ToSql()
	if err != nil {
		r.logger.Error("Failed to build count query", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("Executing count query", zap.String("count_query", countSql))

	var total int
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to execute count query", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("Got total products count", zap.Int("total_count", total))

	dataQuery := baseQuery.Columns("p.id", "p.uuid", "p.name", "p.description", "p.price", "p.stock", "p.created_at", "p.updated_at")

	if filters.SortBy != nil {
		sortOrder := "ASC"
		if filters.SortOrder != nil {
			sortOrder = *filters.SortOrder
		}
		r.logger.Debug("Applying sorting", zap.String("sort_by", *filters.SortBy), zap.String("sort_order", sortOrder))
		dataQuery = dataQuery.OrderBy("p." + *filters.SortBy + " " + sortOrder)
	} else {
		dataQuery = dataQuery.OrderBy("p.created_at DESC")
	}

	offset := (filters.Page - 1) * filters.Limit
	r.logger.Debug("Applying pagination", zap.Int("limit", filters.Limit), zap.Int("offset", offset))
	dataQuery = dataQuery.Limit(uint64(filters.Limit)).Offset(uint64(offset))

	dataSql, dataArgs, err := dataQuery.ToSql()
	if err != nil {
		r.logger.Error("Failed to build data query", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("Executing data query", zap.String("data_query", dataSql))

	rows, err := r.db.Query(ctx, dataSql, dataArgs...)
	if err != nil {
		r.logger.Error("Failed to execute data query", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		var product entities.Product

		if err := rows.Scan(&product.ID, &product.UUID, &product.Name, &product.Description,
			&product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt); err != nil {
			r.logger.Error("Failed to scan product row", zap.Error(err))
			return nil, 0, err
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Row iteration error", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Info("Products retrieved successfully",
		zap.Int("products_count", len(products)),
		zap.Int("total_count", total),
		zap.Int("page", filters.Page),
		zap.Int("limit", filters.Limit))

	return products, total, nil
}

func (r *ProductRepository) AddProductToCategories(ctx context.Context, productID int64, categoryIDs []int64) error {
	r.logger.Debug("Adding product to categories",
		zap.Int64("product_id", productID),
		zap.Any("category_ids", categoryIDs))

	query := r.psql.Insert("product_categories").
		Columns("product_id", "category_id").
		Suffix("ON CONFLICT DO NOTHING")

	for _, categoryID := range categoryIDs {
		query = query.Values(productID, categoryID)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("Failed to build insert categories query",
			zap.Error(err),
			zap.Int64("product_id", productID))
		return err
	}

	r.logger.Debug("Executing add categories query", zap.String("query", sql))

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to add product to categories",
			zap.Error(err),
			zap.Int64("product_id", productID),
			zap.Any("category_ids", categoryIDs))
		return err
	}

	r.logger.Info("Product added to categories successfully",
		zap.Int64("product_id", productID),
		zap.Any("category_ids", categoryIDs))

	return nil
}

func (r *ProductRepository) RemoveProductFromCategories(ctx context.Context, productID int64) error {
	r.logger.Debug("Removing product from all categories", zap.Int64("product_id", productID))

	query := `DELETE FROM product_categories WHERE product_id = $1`

	_, err := r.db.Exec(ctx, query, productID)
	if err != nil {
		r.logger.Error("Failed to remove product from categories", zap.Error(err), zap.Int64("product_id", productID))
		return err
	}

	r.logger.Info("Product removed from categories successfully", zap.Int64("product_id", productID))
	return nil
}

func (r *ProductRepository) GetProductCategories(ctx context.Context, productID int64) ([]*entities.Category, error) {
	r.logger.Debug("Getting product categories", zap.Int64("product_id", productID))

	query := `SELECT c.id, c.uuid, c.name, c.description, c.created_at, c.updated_at 
			  FROM product_categories pc 
	          JOIN categories c on c.id = pc.category_id
	          WHERE pc.product_id = $1`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		r.logger.Error("Failed to get product categories", zap.Error(err), zap.Int64("product_id", productID))
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
			r.logger.Error("Failed to scan category row",
				zap.Error(err),
				zap.Int64("product_id", productID))
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Row iteration error in categories", zap.Error(err), zap.Int64("product_id", productID))
		return nil, err
	}

	r.logger.Debug("Product categories retrieved successfully",
		zap.Int64("product_id", productID),
		zap.Int("categories_count", len(categories)))

	return categories, nil
}

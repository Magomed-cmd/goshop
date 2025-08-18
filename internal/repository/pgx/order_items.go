package pgx

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	errors2 "goshop/internal/domain/errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type OrderItemRepository struct {
	db     *pgxpool.Pool
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewOrderItemRepository(db *pgxpool.Pool) *OrderItemRepository {
	return &OrderItemRepository{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *OrderItemRepository) Create(ctx context.Context, items []*entities.OrderItem) error {
	r.logger.Debug("Creating order items", zap.Int("items_count", len(items)))

	if len(items) == 0 {
		r.logger.Debug("No items to create, returning early")
		return nil
	}

	query := r.psql.Insert("order_items").
		Columns("order_id", "product_id", "product_name", "price_at_order", "quantity")

	for _, item := range items {
		query = query.Values(
			item.OrderID,
			item.ProductID,
			item.ProductName,
			item.PriceAtOrder,
			item.Quantity,
		)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("Failed to build insert query", zap.Error(err), zap.Int("items_count", len(items)))
		return err
	}

	r.logger.Debug("Executing order items insert", zap.String("query", sql), zap.Int("items_count", len(items)))

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if strings.Contains(pgErr.Detail, "user_id") {
					r.logger.Error("User not found while creating order items", zap.Error(err), zap.Int("items_count", len(items)))
					return errors2.ErrUserNotFound
				}
				if strings.Contains(pgErr.Detail, "address_id") {
					r.logger.Error("Address not found while creating order items", zap.Error(err), zap.Int("items_count", len(items)))
					return errors2.ErrAddressNotFound
				}
			}
		}
		r.logger.Error("Failed to create order items", zap.Error(err), zap.Int("items_count", len(items)))
		return err
	}

	r.logger.Info("Order items created successfully", zap.Int("items_count", len(items)))
	return nil
}

package database

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	errors2 "goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	portrepo "goshop/internal/core/ports/repositories"
)

type OrderRepository struct {
	base   BaseRepository
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewOrderRepository(conn portrepo.DBConn, logger *zap.Logger) *OrderRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &OrderRepository{
		base:   NewBaseRepository(conn),
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *entities.Order) (*int64, error) {

	query := `
			INSERT INTO orders (uuid, user_id, address_id, total_price, status, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7) RETURNING id
			`

	var id int64
	err := r.base.Conn().QueryRow(
		ctx,
		query,
		order.UUID,
		order.UserID,
		order.AddressID,
		order.TotalPrice,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if strings.Contains(pgErr.Detail, "user_id") {
					return nil, errors2.ErrUserNotFound
				}
				if strings.Contains(pgErr.Detail, "address_id") {
					return nil, errors2.ErrAddressNotFound
				}
			}
		}
		return nil, err
	}

	return &id, nil
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) ([]*entities.Order, int64, error) {

	offset := (filters.Page - 1) * filters.Limit

	countQuery := r.psql.Select("COUNT(*)").From("orders").Where(squirrel.Eq{"user_id": userID})
	dataQuery := r.psql.Select("*").From("orders").Where(squirrel.Eq{"user_id": userID})

	dataQuery = dataQuery.Limit(uint64(filters.Limit)).Offset(uint64(offset))

	if filters.Status != nil {
		countQuery = countQuery.Where(squirrel.Eq{"status": filters.Status})
		dataQuery = dataQuery.Where(squirrel.Eq{"status": filters.Status})
	}

	if filters.DateFrom != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filters.DateFrom})
		dataQuery = dataQuery.Where(squirrel.GtOrEq{"created_at": filters.DateFrom})
	}

	if filters.DateTo != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filters.DateTo})
		dataQuery = dataQuery.Where(squirrel.LtOrEq{"created_at": filters.DateTo})
	}

	if filters.MinAmount != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"total_price": filters.MinAmount})
		dataQuery = dataQuery.Where(squirrel.GtOrEq{"total_price": filters.MinAmount})
	}

	if filters.MaxAmount != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"total_price": filters.MaxAmount})
		dataQuery = dataQuery.Where(squirrel.LtOrEq{"total_price": filters.MaxAmount})
	}

	if filters.SortBy != nil && filters.SortOrder != nil {
		dataQuery = dataQuery.OrderBy(*filters.SortBy + " " + *filters.SortOrder)
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var totalCount int64
	err = r.base.Conn().QueryRow(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	sql, args, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.base.Conn().Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*entities.Order

	for rows.Next() {
		order := &entities.Order{}

		if err := rows.Scan(
			&order.ID,
			&order.UUID,
			&order.UserID,
			&order.AddressID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, userID int64, orderID int64) (*entities.Order, error) {

	query := `SELECT * from orders WHERE id = $1 AND user_id = $2`

	order := &entities.Order{}
	if err := r.base.Conn().QueryRow(ctx, query, orderID, userID).
		Scan(
			&order.ID,
			&order.UUID,
			&order.UserID,
			&order.AddressID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors2.ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {

	query := `
			UPDATE orders SET status = $1 WHERE id = $2
			`

	result, err := r.base.Conn().Exec(ctx, query, orderID, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) CancelOrder(ctx context.Context, orderID int64) error {

	query := `UPDATE orders 
              SET status = 'cancelled', updated_at = NOW()
              WHERE id = $1 AND status IN ('pending', 'paid')
              `

	result, err := r.base.Conn().Exec(ctx, query, orderID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrOrderCannotBeCancelled
	}

	return nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) ([]*entities.Order, int64, error) {
	r.logger.Debug("Getting all orders with admin filters", zap.Any("filters", filters))

	offset := (filters.Page - 1) * filters.Limit

	countQuery := r.psql.Select("COUNT(*)").From("orders")
	dataQuery := r.psql.Select("*").From("orders")

	dataQuery = dataQuery.Limit(uint64(filters.Limit)).Offset(uint64(offset))

	if filters.Status != nil {
		countQuery = countQuery.Where(squirrel.Eq{"status": filters.Status})
		dataQuery = dataQuery.Where(squirrel.Eq{"status": filters.Status})
	}

	if filters.DateFrom != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filters.DateFrom})
		dataQuery = dataQuery.Where(squirrel.GtOrEq{"created_at": filters.DateFrom})
	}

	if filters.DateTo != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filters.DateTo})
		dataQuery = dataQuery.Where(squirrel.LtOrEq{"created_at": filters.DateTo})
	}

	if filters.MinAmount != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"total_price": filters.MinAmount})
		dataQuery = dataQuery.Where(squirrel.GtOrEq{"total_price": filters.MinAmount})
	}

	if filters.MaxAmount != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"total_price": filters.MaxAmount})
		dataQuery = dataQuery.Where(squirrel.LtOrEq{"total_price": filters.MaxAmount})
	}

	if filters.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"user_id": filters.UserID})
		dataQuery = dataQuery.Where(squirrel.Eq{"user_id": filters.UserID})
	}

	if filters.SortBy != nil && filters.SortOrder != nil {
		dataQuery = dataQuery.OrderBy(*filters.SortBy + " " + *filters.SortOrder)
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		r.logger.Error("Failed to build count query", zap.Error(err))
		return nil, 0, err
	}

	var totalCount int64
	err = r.base.Conn().QueryRow(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to execute count query", zap.Error(err))
		return nil, 0, err
	}

	sql, args, err := dataQuery.ToSql()
	if err != nil {
		r.logger.Error("Failed to build data query", zap.Error(err))
		return nil, 0, err
	}

	rows, err := r.base.Conn().Query(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to execute data query", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*entities.Order

	for rows.Next() {
		order := &entities.Order{}

		if err := rows.Scan(
			&order.ID,
			&order.UUID,
			&order.UserID,
			&order.AddressID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan order row", zap.Error(err))
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Info("All orders retrieved successfully",
		zap.Int("orders_count", len(orders)),
		zap.Int64("total_count", totalCount))

	return orders, totalCount, nil
}

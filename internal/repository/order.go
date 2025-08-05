package repository

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"strings"
)

type OrderRepository struct {
	db   *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *entities.Order) (*int64, error) {

	query := `
			INSERT INTO orders (uuid, user_id, address_id, total_price, status, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7) RETURNING id
			`

	var id int64
	err := r.db.QueryRow(
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
					return nil, domain_errors.ErrUserNotFound
				}
				if strings.Contains(pgErr.Detail, "address_id") {
					return nil, domain_errors.ErrAddressNotFound
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
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	sql, args, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
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

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int) (*entities.Order, error) {

	query := `SELECT * from orders WHERE id = $1`

	order := &entities.Order{}
	if err := r.db.QueryRow(ctx, query, orderID).
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
			return nil, domain_errors.ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {

	query := `
			UPDATE orders SET status = $1 WHERE id = $2
			`

	result, err := r.db.Exec(ctx, query, orderID, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain_errors.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) CancelOrder(ctx context.Context, orderID int) error {

	query := `UPDATE orders 
              SET status = 'cancelled', updated_at = NOW()
              WHERE id = $1 AND status IN ('pending', 'paid')
              `

	result, err := r.db.Exec(ctx, query, orderID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain_errors.ErrOrderCannotBeCancelled
	}

	return nil
}

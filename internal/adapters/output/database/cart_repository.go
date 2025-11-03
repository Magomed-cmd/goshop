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
	portrepo "goshop/internal/core/ports/repositories"
)

type CartRepository struct {
	base   BaseRepository
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewCartRepository(conn portrepo.DBConn, logger *zap.Logger) *CartRepository {
	return &CartRepository{
		base:   NewBaseRepository(conn),
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *CartRepository) GetUserCart(ctx context.Context, userID int64) (*entities.Cart, error) {
	query := `
		SELECT 
		c.id, c.uuid, c.user_id, c.created_at
		FROM carts c
		WHERE user_id = $1;
		`

	var cart entities.Cart

	if err := r.base.Conn().QueryRow(ctx, query, userID).Scan(
		&cart.ID,
		&cart.UUID,
		&cart.UserID,
		&cart.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors2.ErrCartNotFound
		}
		r.logger.Error("Failed to get user cart", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	query = `
	SELECT 
	    ci.cart_id, ci.product_id, ci.quantity, 
	    p.id, p.uuid, p.name, p.description, p.price, p.stock, 
	    p.created_at, p.updated_at
	FROM cart_items ci
	JOIN products p ON p.id = ci.product_id
	WHERE cart_id = $1
	`

	rows, err := r.base.Conn().Query(ctx, query, cart.ID)
	if err != nil {
		r.logger.Error("Failed to get cart items", zap.Error(err), zap.Int64("cart_id", cart.ID))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cartItem entities.CartItem
		cartItem.Product = &entities.Product{}

		if err := rows.Scan(
			&cartItem.CartID,
			&cartItem.ProductID,
			&cartItem.Quantity,
			&cartItem.Product.ID,
			&cartItem.Product.UUID,
			&cartItem.Product.Name,
			&cartItem.Product.Description,
			&cartItem.Product.Price,
			&cartItem.Product.Stock,
			&cartItem.Product.CreatedAt,
			&cartItem.Product.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan cart item", zap.Error(err), zap.Int64("cart_id", cart.ID))
			return nil, err
		}

		if cartItem.Product.ID == 0 {
			r.logger.Warn("Product not found for cart item", zap.Int64("product_id", cartItem.ProductID))
			return nil, errors2.ErrProductNotFound
		}

		cart.Items = append(cart.Items, cartItem)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error during cart items iteration", zap.Error(err), zap.Int64("cart_id", cart.ID))
		return nil, err
	}

	return &cart, nil
}

func (r *CartRepository) AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
	query := `INSERT INTO cart_items (cart_id, product_id, quantity)
			  VALUES ($1, $2, $3)
			  ON CONFLICT (cart_id, product_id)
			  DO UPDATE SET quantity = cart_items.quantity + excluded.quantity
			  `

	result, err := r.base.Conn().Exec(ctx, query, cartID, productID, quantity)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if strings.Contains(pgErr.Detail, "cart_id") {
					return errors2.ErrCartNotFound
				}
				if strings.Contains(pgErr.Detail, "product_id") {
					return errors2.ErrProductNotFound
				}
			}
		}
		r.logger.Error("Failed to add item to cart", zap.Error(err), zap.Int64("cart_id", cartID), zap.Int64("product_id", productID))
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrCartNotFound
	}

	return nil
}

func (r *CartRepository) UpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
	query := `UPDATE cart_items 
			  SET quantity = $3
			  WHERE cart_id = $1 AND product_id = $2
			  `

	result, err := r.base.Conn().Exec(ctx, query, cartID, productID, quantity)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if strings.Contains(pgErr.Detail, "cart_id") {
					return errors2.ErrCartNotFound
				}
				if strings.Contains(pgErr.Detail, "product_id") {
					return errors2.ErrProductNotFound
				}
			}
		}
		r.logger.Error("Failed to update cart item", zap.Error(err), zap.Int64("cart_id", cartID), zap.Int64("product_id", productID))
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrCartItemNotFound
	}

	return nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, cartID int64, productID int64) error {
	query := `DELETE FROM cart_items 
			  WHERE cart_id = $1 AND product_id = $2
			  `

	result, err := r.base.Conn().Exec(ctx, query, cartID, productID)
	if err != nil {
		r.logger.Error("Failed to remove cart item", zap.Error(err), zap.Int64("cart_id", cartID), zap.Int64("product_id", productID))
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrCartItemNotFound
	}

	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, cartID int64) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`

	_, err := r.base.Conn().Exec(ctx, query, cartID)
	if err != nil {
		r.logger.Error("Failed to clear cart", zap.Error(err), zap.Int64("cart_id", cartID))
		return err
	}

	return nil
}

func (r *CartRepository) CreateCart(ctx context.Context, cart *entities.Cart) error {
	query := `INSERT INTO carts (uuid, user_id, created_at)
			  VALUES ($1, $2, $3)
			  RETURNING id`

	err := r.base.Conn().QueryRow(ctx, query, cart.UUID, cart.UserID, cart.CreatedAt).Scan(&cart.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors2.ErrDuplicate
			case "23503":
				return errors2.ErrUserNotFound
			}
		}
		r.logger.Error("Failed to create cart", zap.Error(err), zap.Int64("user_id", cart.UserID))
		return err
	}

	return nil
}

package main

// Пример использования транзакций (правильная абстракция)

/*
import "goshop/internal/db"

type OrderService struct {
	txManager db.TxManager
	// остальные поля
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Начинаем транзакцию через интерфейс
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	// 1. Создаем заказ
	order := &entities.Order{...}
	orderID, err := s.orderRepo.CreateOrderTx(ctx, tx, order)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// 2. Создаем items
	if err := s.orderItemRepo.CreateTx(ctx, tx, orderItems); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// 3. Очищаем корзину
	if err := s.cartRepo.ClearCartTx(ctx, tx, cart.ID); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	// Коммитим
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &dto.OrderResponse{...}, nil
}
*/

// В репозиториях:
/*
import (
	"github.com/jackc/pgx/v5"
	"goshop/internal/db"
)

type OrderRepository struct {
	pool interface{} // pgxpool.Pool или любая DB
}

// Внутри репозитория можно работать с pgx.Tx напрямую
func (r *OrderRepository) CreateOrderTx(ctx context.Context, transaction db.Transaction, order *entities.Order) (*int64, error) {
	// Приводим к PgxTransaction (только в слое repository!)
	tx := transaction.(*postgres.PgxTransaction).Tx

	query := `INSERT INTO orders (...) VALUES (...) RETURNING id`
	var id int64
	err := tx.QueryRow(ctx, query, ...).Scan(&id)
	return &id, err
}
*/

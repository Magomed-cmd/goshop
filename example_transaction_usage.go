package main

// Пример использования транзакций в CreateOrder сервисе

/*
type OrderService struct {
	txManager     *postgres.TxManager
	// остальные поля
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	var orderID *int64
	var orderItems []*entities.OrderItem

	// Все операции в одной транзакции
	err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		// 1. Создаем заказ
		order := &entities.Order{...}
		id, err := s.orderRepo.CreateOrderTx(ctx, tx, order)
		if err != nil {
			return err
		}
		orderID = id

		// 2. Создаем items
		orderItems = make([]*entities.OrderItem, 0, len(cart.Items))
		for _, item := range cart.Items {
			orderItem := &entities.OrderItem{
				OrderID: *id,
				// ... остальные поля
			}
			orderItems = append(orderItems, orderItem)
		}

		if err := s.orderItemRepo.CreateTx(ctx, tx, orderItems); err != nil {
			return err
		}

		// 3. Очищаем корзину
		if err := s.cartRepo.ClearCartTx(ctx, tx, cart.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.OrderResponse{...}, nil
}
*/

// Для репозиториев добавь методы с Tx:
/*
func (r *OrderRepository) CreateOrderTx(ctx context.Context, tx pgx.Tx, order *entities.Order) (*int64, error) {
	query := `INSERT INTO orders (...) VALUES (...) RETURNING id`

	var id int64
	err := tx.QueryRow(ctx, query, ...).Scan(&id)
	return &id, err
}
*/

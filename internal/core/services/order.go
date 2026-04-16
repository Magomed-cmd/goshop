package services

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	"goshop/internal/core/ports/repositories"
	dtx "goshop/internal/core/ports/transaction"
)

type OrderService struct {
	orderRepo     repositories.OrderRepository
	cartRepo      repositories.CartRepository
	userRepo      repositories.UserRepository
	addressRepo   repositories.AddressRepository
	orderItemRepo repositories.OrderItemRepository
	uow           dtx.UnitOfWork
	logger        *zap.Logger
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	userRepo repositories.UserRepository,
	addressRepo repositories.AddressRepository,
	orderItemRepo repositories.OrderItemRepository,
	uow dtx.UnitOfWork,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		userRepo:      userRepo,
		addressRepo:   addressRepo,
		orderItemRepo: orderItemRepo,
		uow:           uow,
		logger:        logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, addressID *int64) (*entities.Order, error) {
	s.logger.Info("Creating order", zap.Int64("user_id", userID), zap.Int64("address_id", *addressID))

	var orderResult *entities.Order
	err := s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		s.logger.Debug("Getting user cart", zap.Int64("user_id", userID))
		cart, err := repos.Carts().GetUserCart(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get user cart", zap.Error(err), zap.Int64("user_id", userID))
			return err
		}

		if len(cart.Items) == 0 {
			s.logger.Warn("Cart is empty", zap.Int64("user_id", userID))
			return errors.ErrCartEmpty
		}

		s.logger.Debug("Getting user by ID", zap.Int64("user_id", userID))
		user, err := repos.Users().GetUserByID(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get user", zap.Error(err), zap.Int64("user_id", userID))
			return err
		}

		s.logger.Debug("Getting address", zap.Int64("address_id", *addressID))
		userAddress, err := repos.Addresses().GetAddressByID(ctx, *addressID)
		if err != nil {
			s.logger.Error("Failed to get address", zap.Error(err), zap.Int64("address_id", *addressID))
			return err
		}

		s.logger.Debug("Calculating total price", zap.Int("items_count", len(cart.Items)))
		var totalPrice decimal.Decimal
		for _, item := range cart.Items {
			itemTotal := item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			totalPrice = totalPrice.Add(itemTotal)
		}

		order := &entities.Order{
			UUID:       uuid.New(),
			UserID:     userID,
			AddressID:  addressID,
			TotalPrice: totalPrice,
			Status:     entities.OrderStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			User:       user,
			Address:    userAddress,
		}

		s.logger.Debug("Creating order in database", zap.String("order_uuid", order.UUID.String()))
		id, err := repos.Orders().CreateOrder(ctx, order)
		if err != nil {
			s.logger.Error("Failed to create order", zap.Error(err), zap.Int64("user_id", userID))
			return err
		}

		orderItems := make([]*entities.OrderItem, 0, len(cart.Items))
		for _, item := range cart.Items {
			orderItem := &entities.OrderItem{
				OrderID:      *id,
				ProductID:    item.Product.ID,
				ProductName:  item.Product.Name,
				PriceAtOrder: item.Product.Price,
				Quantity:     item.Quantity,
			}
			orderItems = append(orderItems, orderItem)
		}

		s.logger.Debug("Creating order items", zap.Int("items_count", len(orderItems)))
		err = repos.OrderItems().Create(ctx, orderItems)
		if err != nil {
			s.logger.Error("Failed to create order items", zap.Error(err), zap.Int64("order_id", *id))
			return err
		}

		orderItemEntities := make([]entities.OrderItem, len(orderItems))
		for i, item := range orderItems {
			orderItemEntities[i] = *item
		}

		order.ID = *id
		order.Items = orderItemEntities

		s.logger.Debug("Clearing cart", zap.Int64("cart_id", cart.ID))
		err = repos.Carts().ClearCart(ctx, cart.ID)
		if err != nil {
			s.logger.Error("Failed to clear cart", zap.Error(err), zap.Int64("cart_id", cart.ID))
			return err
		}

		orderResult = order
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Order created successfully",
		zap.Int64("order_id", orderResult.ID),
		zap.Int64("user_id", userID),
		zap.String("total_price", orderResult.TotalPrice.String()))

	return orderResult, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) ([]*entities.Order, int64, error) {

	var orders []*entities.Order
	var totalCount int64
	err := s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		orders, totalCount, innerErr = repos.Orders().GetUserOrders(ctx, userID, filters)
		return innerErr
	})
	if err != nil {
		return nil, 0, err
	}
	return orders, totalCount, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID int64, orderID int64) (*entities.Order, error) {

	var order *entities.Order
	err := s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		order, innerErr = repos.Orders().GetOrderByID(ctx, userID, orderID)
		return innerErr
	})
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID int64, orderID int64) error {

	return s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		order, err := repos.Orders().GetOrderByID(ctx, userID, orderID)
		if err != nil {
			return err
		}

		if order.Status == entities.OrderStatusCancelled {
			return errors.ErrOrderAlreadyCancelled
		}

		if (order.Status != entities.OrderStatusPending) && (order.Status != entities.OrderStatusPaid) {
			return errors.ErrOrderCannotBeCancelled
		}

		return repos.Orders().CancelOrder(ctx, orderID)
	})
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	s.logger.Info("Updating order status", zap.Int64("order_id", orderID), zap.String("status", status))

	validStatuses := []string{"pending", "paid", "shipped", "delivered", "cancelled"}
	if !slices.Contains(validStatuses, status) {
		return errors.ErrInvalidOrderStatus
	}

	return s.withinWriteUOW(ctx, func(repos dtx.Repositories) error {
		return repos.Orders().UpdateOrderStatus(ctx, orderID, status)
	})
}

func (s *OrderService) GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) ([]*entities.Order, int64, error) {
	s.logger.Info("Getting all orders for admin", zap.Any("filters", filters))

	var orders []*entities.Order
	var totalCount int64
	err := s.withinReadUOW(ctx, func(repos dtx.Repositories) error {
		var innerErr error
		orders, totalCount, innerErr = repos.Orders().GetAllOrders(ctx, filters)
		return innerErr
	})
	if err != nil {
		s.logger.Error("Failed to get all orders", zap.Error(err))
		return nil, 0, err
	}

	s.logger.Info("All orders retrieved successfully",
		zap.Int("orders_count", len(orders)),
		zap.Int64("total_count", totalCount))

	return orders, totalCount, nil
}

func (s *OrderService) withinWriteUOW(ctx context.Context, fn func(repos dtx.Repositories) error) error {
	if s.uow == nil {
		return fn(&orderFallbackRepos{
			users:      s.userRepo,
			addresses:  s.addressRepo,
			carts:      s.cartRepo,
			orders:     s.orderRepo,
			orderItems: s.orderItemRepo,
		})
	}
	return s.uow.Do(ctx, fn)
}

func (s *OrderService) withinReadUOW(ctx context.Context, fn func(repos dtx.Repositories) error) error {
	if s.uow == nil {
		return fn(&orderFallbackRepos{
			users:      s.userRepo,
			addresses:  s.addressRepo,
			carts:      s.cartRepo,
			orders:     s.orderRepo,
			orderItems: s.orderItemRepo,
		})
	}
	return s.uow.DoRead(ctx, fn)
}

type orderFallbackRepos struct {
	users      repositories.UserRepository
	addresses  repositories.AddressRepository
	carts      repositories.CartRepository
	orders     repositories.OrderRepository
	orderItems repositories.OrderItemRepository
}

func (f *orderFallbackRepos) Users() repositories.UserRepository           { return f.users }
func (f *orderFallbackRepos) Roles() repositories.RoleRepository           { return nil }
func (f *orderFallbackRepos) Addresses() repositories.AddressRepository    { return f.addresses }
func (f *orderFallbackRepos) Categories() repositories.CategoryRepository  { return nil }
func (f *orderFallbackRepos) Products() repositories.ProductRepository     { return nil }
func (f *orderFallbackRepos) Carts() repositories.CartRepository           { return f.carts }
func (f *orderFallbackRepos) Orders() repositories.OrderRepository         { return f.orders }
func (f *orderFallbackRepos) OrderItems() repositories.OrderItemRepository { return f.orderItems }
func (f *orderFallbackRepos) Reviews() repositories.ReviewRepository       { return nil }

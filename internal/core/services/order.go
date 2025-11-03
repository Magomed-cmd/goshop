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
	"goshop/internal/core/mappers"
	"goshop/internal/core/ports/repositories"
	"goshop/internal/dto"
)

type OrderService struct {
	orderRepo     repositories.OrderRepository
	cartRepo      repositories.CartRepository
	userRepo      repositories.UserRepository
	addressRepo   repositories.AddressRepository
	orderItemRepo repositories.OrderItemRepository
	logger        *zap.Logger
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	userRepo repositories.UserRepository,
	addressRepo repositories.AddressRepository,
	orderItemRepo repositories.OrderItemRepository,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		userRepo:      userRepo,
		addressRepo:   addressRepo,
		orderItemRepo: orderItemRepo,
		logger:        logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	s.logger.Info("Creating order", zap.Int64("user_id", userID), zap.Int64("address_id", *req.AddressID))

	s.logger.Debug("Getting user cart", zap.Int64("user_id", userID))
	cart, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user cart", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	if len(cart.Items) == 0 {
		s.logger.Warn("Cart is empty", zap.Int64("user_id", userID))
		return nil, errors.ErrCartEmpty
	}

	s.logger.Debug("Getting user by ID", zap.Int64("user_id", userID))
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	s.logger.Debug("Getting address", zap.Int64("address_id", *req.AddressID))
	userAddress, err := s.addressRepo.GetAddressByID(ctx, *req.AddressID)
	if err != nil {
		s.logger.Error("Failed to get address", zap.Error(err), zap.Int64("address_id", *req.AddressID))
		return nil, err
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
		AddressID:  req.AddressID,
		TotalPrice: totalPrice,
		Status:     entities.OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		User:       user,
		Address:    userAddress,
	}

	s.logger.Debug("Creating order in database", zap.String("order_uuid", order.UUID.String()))
	id, err := s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Error("Failed to create order", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
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
	err = s.orderItemRepo.Create(ctx, orderItems)
	if err != nil {
		s.logger.Error("Failed to create order items", zap.Error(err), zap.Int64("order_id", *id))
		return nil, err
	}

	orderItemEntities := make([]entities.OrderItem, len(orderItems))
	for i, item := range orderItems {
		orderItemEntities[i] = *item
	}

	order.ID = *id
	order.Items = orderItemEntities

	s.logger.Debug("Clearing cart", zap.Int64("cart_id", cart.ID))
	err = s.cartRepo.ClearCart(ctx, cart.ID)
	if err != nil {
		s.logger.Error("Failed to clear cart", zap.Error(err), zap.Int64("cart_id", cart.ID))
		return nil, err
	}

	response := mappers.ToOrderResponse(order)

	s.logger.Info("Order created successfully",
		zap.Int64("order_id", *id),
		zap.Int64("user_id", userID),
		zap.String("total_price", totalPrice.StringFixed(2)))

	return response, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) (*dto.OrdersListResponse, error) {

	orders, totalCount, err := s.orderRepo.GetUserOrders(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	return mappers.ToOrdersListResponse(orders, totalCount, filters.Page, filters.Limit), nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID int64, orderID int64) (*dto.OrderResponse, error) {

	order, err := s.orderRepo.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	return mappers.ToOrderResponse(order), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID int64, orderID int64) error {

	order, err := s.orderRepo.GetOrderByID(ctx, userID, orderID)
	if err != nil {
		return err
	}

	if order.Status == entities.OrderStatusCancelled {
		return errors.ErrOrderAlreadyCancelled
	}

	if (order.Status != entities.OrderStatusPending) && (order.Status != entities.OrderStatusPaid) {
		return errors.ErrOrderCannotBeCancelled
	}

	err = s.orderRepo.CancelOrder(ctx, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	s.logger.Info("Updating order status", zap.Int64("order_id", orderID), zap.String("status", status))

	validStatuses := []string{"pending", "paid", "shipped", "delivered", "cancelled"}
	if !slices.Contains(validStatuses, status) {
		return errors.ErrInvalidOrderStatus
	}

	return s.orderRepo.UpdateOrderStatus(ctx, orderID, status)
}

func (s *OrderService) GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) (*dto.OrdersListResponse, error) {
	s.logger.Info("Getting all orders for admin", zap.Any("filters", filters))

	orders, totalCount, err := s.orderRepo.GetAllOrders(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to get all orders", zap.Error(err))
		return nil, err
	}

	s.logger.Info("All orders retrieved successfully",
		zap.Int("orders_count", len(orders)),
		zap.Int64("total_count", totalCount))

	return mappers.ToOrdersListResponse(orders, totalCount, filters.Page, filters.Limit), nil
}

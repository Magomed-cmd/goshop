package order

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"time"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entities.Order) (*int, error)
	GetUserOrders(ctx context.Context, userID int) ([]*entities.Order, error)
	GetOrderByID(ctx context.Context, orderID int) (*entities.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	CancelOrder(ctx context.Context, orderID int) error
	ClearCart(ctx context.Context, cartID int64) error
}

type CartRepository interface {
	GetUserCart(ctx context.Context, userID int64) (*entities.Cart, error)
	ClearCart(ctx context.Context, cartID int64) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
}

type AddressRepository interface {
	GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error)
}

type OrderItemRepository interface {
	Create(ctx context.Context, items []*entities.OrderItem) error
}

type OrderService struct {
	orderRepo     OrderRepository
	cartRepo      CartRepository
	userRepo      UserRepository
	addressRepo   AddressRepository
	orderItemRepo OrderItemRepository
	logger        *zap.Logger
}

func NewOrderService(orderRepo OrderRepository, logger *zap.Logger) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
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
		return nil, domain_errors.ErrCartEmpty
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
			OrderID:      int64(*id),
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
		s.logger.Error("Failed to create order items", zap.Error(err), zap.Int64("order_id", int64(*id)))
		return nil, err
	}

	s.logger.Debug("Clearing cart", zap.Int64("cart_id", cart.ID))
	err = s.cartRepo.ClearCart(ctx, cart.ID)
	if err != nil {
		s.logger.Error("Failed to clear cart", zap.Error(err), zap.Int64("cart_id", cart.ID))
		return nil, err
	}

	itemResponses := make([]dto.OrderItemResponse, len(orderItems))
	for i, item := range orderItems {
		itemResponses[i] = dto.OrderItemResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			PriceAtOrder: item.PriceAtOrder.StringFixed(2),
			Quantity:     item.Quantity,
		}
	}

	response := &dto.OrderResponse{
		ID:         int64(*id),
		UUID:       order.UUID.String(),
		UserID:     userID,
		AddressID:  req.AddressID,
		TotalPrice: order.TotalPrice.StringFixed(2),
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		Items:      itemResponses,
		Address: &dto.AddressResponse{
			ID:         userAddress.ID,
			UUID:       userAddress.UUID.String(),
			Address:    userAddress.Address,
			City:       userAddress.City,
			PostalCode: userAddress.PostalCode,
			Country:    userAddress.Country,
			CreatedAt:  userAddress.CreatedAt.Format(time.RFC3339),
		},
	}

	s.logger.Info("Order created successfully",
		zap.Int64("order_id", int64(*id)),
		zap.Int64("user_id", userID),
		zap.String("total_price", totalPrice.StringFixed(2)))

	return response, nil
}

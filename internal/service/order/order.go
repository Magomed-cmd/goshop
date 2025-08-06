package order

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"slices"
	"time"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entities.Order) (*int64, error)
	GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) ([]*entities.Order, int64, error)
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (*entities.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) ([]*entities.Order, int64, error)
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

func NewOrderService(
	orderRepo OrderRepository,
	cartRepo CartRepository,
	userRepo UserRepository,
	addressRepo AddressRepository,
	orderItemRepo OrderItemRepository,
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
		ID:         *id,
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

	resp := &dto.OrdersListResponse{}
	orderResponses := make([]dto.OrderResponse, 0, len(orders))
	totalAmount := decimal.Zero

	for _, order := range orders {
		items := make([]dto.OrderItemResponse, 0, len(order.Items))
		for _, orderItem := range order.Items {

			subTotal := orderItem.PriceAtOrder.Mul(decimal.NewFromInt(int64(orderItem.Quantity)))

			totalAmount = totalAmount.Add(subTotal)
			item := dto.OrderItemResponse{
				ProductID:    orderItem.ProductID,
				ProductName:  orderItem.ProductName,
				PriceAtOrder: orderItem.PriceAtOrder.StringFixed(2),
				Quantity:     orderItem.Quantity,
				Subtotal:     subTotal.StringFixed(2),
			}
			items = append(items, item)
		}

		Address := &dto.AddressResponse{
			ID:         order.Address.ID,
			UUID:       order.Address.UUID.String(),
			Address:    order.Address.Address,
			City:       order.Address.City,
			PostalCode: order.Address.PostalCode,
			Country:    order.Address.Country,
			CreatedAt:  order.Address.CreatedAt.Format(time.RFC3339),
		}

		orderResponse := dto.OrderResponse{
			ID:         order.ID,
			UUID:       order.UUID.String(),
			UserID:     order.UserID,
			AddressID:  order.AddressID,
			TotalPrice: order.TotalPrice.StringFixed(2),
			Status:     string(order.Status),
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
			Items:      items,
			Address:    Address,
		}
		orderResponses = append(orderResponses, orderResponse)
	}

	resp.Orders = orderResponses
	resp.TotalCount = totalCount
	resp.TotalAmount = totalAmount.StringFixed(2)
	resp.Page = filters.Page
	resp.Limit = filters.Limit

	return resp, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID int64, orderID int64) (*dto.OrderResponse, error) {

	order, err := s.orderRepo.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	orderItems := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		subTotal := item.PriceAtOrder.Mul(decimal.NewFromInt(int64(item.Quantity)))

		orderItem := dto.OrderItemResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			PriceAtOrder: item.PriceAtOrder.StringFixed(2),
			Quantity:     item.Quantity,
			Subtotal:     subTotal.StringFixed(2),
		}
		orderItems = append(orderItems, orderItem)
	}

	address := &dto.AddressResponse{
		ID:         order.Address.ID,
		UUID:       order.Address.UUID.String(),
		Address:    order.Address.Address,
		City:       order.Address.City,
		PostalCode: order.Address.PostalCode,
		Country:    order.Address.Country,
		CreatedAt:  order.Address.CreatedAt.Format(time.RFC3339),
	}

	resp := &dto.OrderResponse{
		ID:         order.ID,
		UUID:       order.UUID.String(),
		UserID:     order.UserID,
		AddressID:  order.AddressID,
		TotalPrice: order.TotalPrice.StringFixed(2),
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		Items:      orderItems,
		Address:    address,
	}

	return resp, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID int64, orderID int64) error {

	order, err := s.orderRepo.GetOrderByID(ctx, userID, orderID)
	if err != nil {
		return err
	}

	if order.Status == entities.OrderStatusCancelled {
		return domain_errors.ErrOrderAlreadyCancelled
	}

	if (order.Status != entities.OrderStatusPending) && (order.Status != entities.OrderStatusPaid) {
		return domain_errors.ErrOrderCannotBeCancelled
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
		return domain_errors.ErrInvalidOrderStatus
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

	orderResponses := make([]dto.OrderResponse, 0, len(orders))
	var totalAmount decimal.Decimal

	for _, order := range orders {
		items := make([]dto.OrderItemResponse, 0, len(order.Items))

		for _, orderItem := range order.Items {
			subTotal := orderItem.PriceAtOrder.Mul(decimal.NewFromInt(int64(orderItem.Quantity)))

			item := dto.OrderItemResponse{
				ProductID:    orderItem.ProductID,
				ProductName:  orderItem.ProductName,
				PriceAtOrder: orderItem.PriceAtOrder.StringFixed(2),
				Quantity:     orderItem.Quantity,
				Subtotal:     subTotal.StringFixed(2),
			}
			items = append(items, item)
		}

		address := &dto.AddressResponse{
			ID:         order.Address.ID,
			UUID:       order.Address.UUID.String(),
			Address:    order.Address.Address,
			City:       order.Address.City,
			PostalCode: order.Address.PostalCode,
			Country:    order.Address.Country,
			CreatedAt:  order.Address.CreatedAt.Format(time.RFC3339),
		}

		orderResponse := dto.OrderResponse{
			ID:         order.ID,
			UUID:       order.UUID.String(),
			UserID:     order.UserID,
			AddressID:  order.AddressID,
			TotalPrice: order.TotalPrice.StringFixed(2),
			Status:     string(order.Status),
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
			Items:      items,
			Address:    address,
		}

		orderResponses = append(orderResponses, orderResponse)
		totalAmount = totalAmount.Add(order.TotalPrice)
	}

	s.logger.Info("All orders retrieved successfully",
		zap.Int("orders_count", len(orders)),
		zap.Int64("total_count", totalCount))

	return &dto.OrdersListResponse{
		Orders:      orderResponses,
		TotalCount:  totalCount,
		TotalAmount: totalAmount.StringFixed(2),
		Page:        filters.Page,
		Limit:       filters.Limit,
	}, nil
}

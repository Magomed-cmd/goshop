package mappers

import (
	"github.com/shopspring/decimal"

	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

func ToOrderResponse(order *entities.Order) *dto.OrderResponse {
	if order == nil {
		return nil
	}

	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, toOrderItemResponse(item))
	}

	var address *dto.AddressResponse
	if order.Address != nil {
		addr := ToAddressResponse(order.Address)
		address = addr
	}

	return &dto.OrderResponse{
		ID:         order.ID,
		UUID:       order.UUID.String(),
		UserID:     order.UserID,
		AddressID:  order.AddressID,
		TotalPrice: decimalToString(order.TotalPrice),
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		Items:      items,
		Address:    address,
	}
}

func ToOrdersListResponse(orders []*entities.Order, totalCount int64, page, limit int) *dto.OrdersListResponse {
	orderResponses := make([]dto.OrderResponse, 0, len(orders))
	totalAmount := decimal.Zero

	for _, order := range orders {
		if order == nil {
			continue
		}

		response := ToOrderResponse(order)
		orderResponses = append(orderResponses, *response)

		totalAmount = totalAmount.Add(order.TotalPrice)
	}

	return &dto.OrdersListResponse{
		Orders:      orderResponses,
		TotalCount:  totalCount,
		TotalAmount: decimalToString(totalAmount),
		Page:        page,
		Limit:       limit,
	}
}

func toOrderItemResponse(item entities.OrderItem) dto.OrderItemResponse {
	subTotal := item.PriceAtOrder.Mul(decimal.NewFromInt(int64(item.Quantity)))

	return dto.OrderItemResponse{
		ProductID:    item.ProductID,
		ProductName:  item.ProductName,
		PriceAtOrder: decimalToString(item.PriceAtOrder),
		Quantity:     item.Quantity,
		Subtotal:     decimalToString(subTotal),
	}
}

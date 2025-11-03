package mappers

import (
	"github.com/shopspring/decimal"

	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

func ToCartResponse(cart *entities.Cart) *dto.CartResponse {
	if cart == nil {
		return nil
	}

	items := make([]dto.CartItemResponse, 0, len(cart.Items))
	totalPrice := decimal.Zero
	totalItems := 0

	for _, item := range cart.Items {
		var (
			price decimal.Decimal
			name  string
		)
		if item.Product != nil {
			price = item.Product.Price
			name = item.Product.Name
		}
		subtotal := price.Mul(decimal.NewFromInt(int64(item.Quantity)))

		items = append(items, dto.CartItemResponse{
			ProductID:   item.ProductID,
			ProductName: name,
			Quantity:    item.Quantity,
			Price:       decimalToString(price),
			Subtotal:    decimalToString(subtotal),
		})

		totalPrice = totalPrice.Add(subtotal)
		totalItems += item.Quantity
	}

	return &dto.CartResponse{
		ID:         cart.ID,
		Items:      items,
		TotalPrice: decimalToString(totalPrice),
		TotalItems: totalItems,
	}
}

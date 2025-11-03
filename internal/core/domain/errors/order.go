package errors

import "errors"

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrOrderItemNotFound      = errors.New("order item not found")
	ErrEmptyCart              = errors.New("cart is empty")
	ErrInvalidOrderStatus     = errors.New("invalid order status")
	ErrOrderAlreadyCancelled  = errors.New("order is already cancelled")
	ErrOrderCannotBeCancelled = errors.New("order cannot be cancelled")
	ErrOrderNotOwnedByUser    = errors.New("order does not belong to user")
	ErrAddressRequired        = errors.New("address is required for this order")
	ErrInvalidOrderData       = errors.New("invalid order data")
)

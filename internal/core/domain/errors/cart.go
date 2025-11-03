package errors

import "errors"

var (
	ErrCartNotFound        = errors.New("cart not found")
	ErrCartItemNotFound    = errors.New("cart item not found")
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrInsufficientStock   = errors.New("insufficient product stock")
	ErrProductNotAvailable = errors.New("product is not available")
	ErrCartItemExists      = errors.New("product already in cart")
	ErrCartEmpty           = errors.New("cart is empty")
)

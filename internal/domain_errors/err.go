package domain_errors

import "errors"

var (
	// General errors
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("access forbidden")
	ErrDuplicate    = errors.New("duplicate entry")
	ErrInvalidInput = errors.New("invalid input data")

	// User errors
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email format")

	// Address errors
	ErrAddressNotFound    = errors.New("address not found")
	ErrInvalidAddressData = errors.New("invalid address data")

	// Category errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryNameExists  = errors.New("category name already exists")
	ErrInvalidCategoryData = errors.New("invalid category data")

	// Product errors
	ErrProductNotFound    = errors.New("product not found")
	ErrProductNameExists  = errors.New("product name already exists")
	ErrInvalidProductData = errors.New("invalid product data")
	ErrInvalidPrice       = errors.New("price must be greater than 0")
	ErrInvalidStock       = errors.New("stock cannot be negative")

	// Cart errors
	ErrCartNotFound        = errors.New("cart not found")
	ErrCartItemNotFound    = errors.New("cart item not found")
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrInsufficientStock   = errors.New("insufficient product stock")
	ErrProductNotAvailable = errors.New("product is not available")
	ErrCartItemExists      = errors.New("product already in cart")
	ErrCartEmpty           = errors.New("cart is empty")

	// Order errors
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

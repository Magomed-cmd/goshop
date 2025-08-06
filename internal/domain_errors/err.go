package domain_errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func HandleError(c *gin.Context, err error) {
	switch {
	// General errors
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	case errors.Is(err, ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden"})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "Duplicate entry"})
	case errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})

	// User errors
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
	case errors.Is(err, ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
	case errors.Is(err, ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})

	// Address errors
	case errors.Is(err, ErrAddressNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
	case errors.Is(err, ErrInvalidAddressData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address data"})

	// Category errors
	case errors.Is(err, ErrCategoryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
	case errors.Is(err, ErrCategoryNameExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Category name already exists"})
	case errors.Is(err, ErrInvalidCategoryData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})

	// Product errors
	case errors.Is(err, ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	case errors.Is(err, ErrProductNameExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product name already exists"})
	case errors.Is(err, ErrInvalidProductData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
	case errors.Is(err, ErrInvalidPrice):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
	case errors.Is(err, ErrInvalidStock):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock cannot be negative"})

	// Cart errors
	case errors.Is(err, ErrCartNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
	case errors.Is(err, ErrCartItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
	case errors.Is(err, ErrInvalidQuantity):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than 0"})
	case errors.Is(err, ErrInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"error": "Insufficient product stock"})
	case errors.Is(err, ErrProductNotAvailable):
		c.JSON(http.StatusConflict, gin.H{"error": "Product is not available"})
	case errors.Is(err, ErrCartItemExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product already in cart"})
	case errors.Is(err, ErrCartEmpty):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})

	// Order errors
	case errors.Is(err, ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
	case errors.Is(err, ErrOrderItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
	case errors.Is(err, ErrEmptyCart):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
	case errors.Is(err, ErrInvalidOrderStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
	case errors.Is(err, ErrOrderAlreadyCancelled):
		c.JSON(http.StatusConflict, gin.H{"error": "Order is already cancelled"})
	case errors.Is(err, ErrOrderCannotBeCancelled):
		c.JSON(http.StatusConflict, gin.H{"error": "Order cannot be cancelled"})
	case errors.Is(err, ErrOrderNotOwnedByUser):
		c.JSON(http.StatusForbidden, gin.H{"error": "Order does not belong to user"})
	case errors.Is(err, ErrAddressRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Address is required for this order"})
	case errors.Is(err, ErrInvalidOrderData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})

	// Default case для неизвестных ошибок
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

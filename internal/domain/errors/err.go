package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Все твои ошибки
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
	ErrInvalidUserID   = errors.New("invalid id of user")

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
	ErrInvalidProductID   = errors.New("invalid product id")

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

	// Review errors
	ErrInvalidRating          = errors.New("invalid rating of review")
	ErrInvalidReviewSortBy    = errors.New("invalid sort_by for review filter")
	ErrInvalidReviewSortOrder = errors.New("invalid sort_order for review filter")
	ErrInvalidReviewID        = errors.New("invalid id of review")
	ErrNothingToUpdate        = errors.New("not enough parameters for update")
	ErrInvalidComment         = errors.New("comment too long (max 1000 characters)")
	ErrReviewNotOwnedByUser   = errors.New("review does not belong to user")

	// user_avatars errors
	ErrAvatarNotFound    = errors.New("avatar not found")
	ErrInvalidAvatarData = errors.New("invalid avatar data")
	ErrAvatarUploadFail  = errors.New("failed to upload avatar")

	// product_images errors
	ErrProductImageNotFound = errors.New("product image not found")

	// Message errors
	ErrMessageNotFound       = errors.New("message not found")
	ErrMessageNotOwnedByUser = errors.New("message does not belong to user")
	ErrInvalidMessageData    = errors.New("invalid message data")
	ErrInvalidMessageID      = errors.New("invalid message id")
	ErrMessageContentEmpty   = errors.New("message content cannot be empty")
	ErrMessageTooLong        = errors.New("message too long (max 5000 characters)")
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
	case errors.Is(err, ErrInvalidUserID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id of user"})

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
	case errors.Is(err, ErrInvalidProductID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})

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

	// Review errors
	case errors.Is(err, ErrInvalidRating):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating of review"})
	case errors.Is(err, ErrInvalidReviewSortBy):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_by for review filter"})
	case errors.Is(err, ErrInvalidReviewSortOrder):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_order for review filter"})
	case errors.Is(err, ErrInvalidReviewID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id of review"})
	case errors.Is(err, ErrNothingToUpdate):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough parameters for update"})
	case errors.Is(err, ErrInvalidComment):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment too long (max 1000 characters)"})
	case errors.Is(err, ErrReviewNotOwnedByUser):
		c.JSON(http.StatusForbidden, gin.H{"error": "Review does not belong to user"})

		// Message errors
	case errors.Is(err, ErrMessageNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
	case errors.Is(err, ErrMessageNotOwnedByUser):
		c.JSON(http.StatusForbidden, gin.H{"error": "Message does not belong to user"})
	case errors.Is(err, ErrInvalidMessageData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message data"})
	case errors.Is(err, ErrInvalidMessageID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message id"})
	case errors.Is(err, ErrMessageContentEmpty):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content cannot be empty"})
	case errors.Is(err, ErrMessageTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message too long (max 5000 characters)"})

	// Default case для неизвестных ошибок
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

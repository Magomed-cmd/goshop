package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleError maps domain errors to HTTP responses.
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
	case errors.Is(err, ErrProductImageUploadFail):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload product image"})
	case errors.Is(err, ErrProductImageDeleteFail):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product image"})
	case errors.Is(err, ErrProductImageNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product image not found"})

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

	// Avatar errors
	case errors.Is(err, ErrAvatarNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar not found"})
	case errors.Is(err, ErrInvalidAvatarData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid avatar data"})
	case errors.Is(err, ErrAvatarUploadFail):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	status  int
	message string
}

func newHandler(status int, message string) handler {
	return handler{status: status, message: message}
}

var errorHandlers = map[error]handler{
	// General
	ErrNotFound:     newHandler(http.StatusNotFound, "Resource not found"),
	ErrUnauthorized: newHandler(http.StatusUnauthorized, "Unauthorized access"),
	ErrForbidden:    newHandler(http.StatusForbidden, "Access forbidden"),
	ErrDuplicate:    newHandler(http.StatusConflict, "Duplicate entry"),
	ErrInvalidInput: newHandler(http.StatusBadRequest, "Invalid input data"),

	// User
	ErrUserNotFound:    newHandler(http.StatusNotFound, "User not found"),
	ErrEmailExists:     newHandler(http.StatusConflict, "Email already exists"),
	ErrInvalidPassword: newHandler(http.StatusBadRequest, "Invalid password"),
	ErrInvalidEmail:    newHandler(http.StatusBadRequest, "Invalid email format"),
	ErrInvalidUserID:   newHandler(http.StatusBadRequest, "Invalid id of user"),

	// Address
	ErrAddressNotFound:    newHandler(http.StatusNotFound, "Address not found"),
	ErrInvalidAddressData: newHandler(http.StatusBadRequest, "Invalid address data"),

	// Category
	ErrCategoryNotFound:    newHandler(http.StatusNotFound, "Category not found"),
	ErrCategoryNameExists:  newHandler(http.StatusConflict, "Category name already exists"),
	ErrInvalidCategoryData: newHandler(http.StatusBadRequest, "Invalid category data"),

	// Product
	ErrProductNotFound:        newHandler(http.StatusNotFound, "Product not found"),
	ErrProductNameExists:      newHandler(http.StatusConflict, "Product name already exists"),
	ErrInvalidProductData:     newHandler(http.StatusBadRequest, "Invalid product data"),
	ErrInvalidPrice:           newHandler(http.StatusBadRequest, "Price must be greater than 0"),
	ErrInvalidStock:           newHandler(http.StatusBadRequest, "Stock cannot be negative"),
	ErrInvalidProductID:       newHandler(http.StatusBadRequest, "Invalid product id"),
	ErrProductImageUploadFail: newHandler(http.StatusInternalServerError, "Failed to upload product image"),
	ErrProductImageDeleteFail: newHandler(http.StatusInternalServerError, "Failed to delete product image"),
	ErrProductImageNotFound:   newHandler(http.StatusNotFound, "Product image not found"),

	// Cart
	ErrCartNotFound:        newHandler(http.StatusNotFound, "Cart not found"),
	ErrCartItemNotFound:    newHandler(http.StatusNotFound, "Cart item not found"),
	ErrInvalidQuantity:     newHandler(http.StatusBadRequest, "Quantity must be greater than 0"),
	ErrInsufficientStock:   newHandler(http.StatusConflict, "Insufficient product stock"),
	ErrProductNotAvailable: newHandler(http.StatusConflict, "Product is not available"),
	ErrCartItemExists:      newHandler(http.StatusConflict, "Product already in cart"),
	ErrCartEmpty:           newHandler(http.StatusBadRequest, "Cart is empty"),

	// Order
	ErrOrderNotFound:          newHandler(http.StatusNotFound, "Order not found"),
	ErrOrderItemNotFound:      newHandler(http.StatusNotFound, "Order item not found"),
	ErrEmptyCart:              newHandler(http.StatusBadRequest, "Cart is empty"),
	ErrInvalidOrderStatus:     newHandler(http.StatusBadRequest, "Invalid order status"),
	ErrOrderAlreadyCancelled:  newHandler(http.StatusConflict, "Order is already cancelled"),
	ErrOrderCannotBeCancelled: newHandler(http.StatusConflict, "Order cannot be cancelled"),
	ErrOrderNotOwnedByUser:    newHandler(http.StatusForbidden, "Order does not belong to user"),
	ErrAddressRequired:        newHandler(http.StatusBadRequest, "Address is required for this order"),
	ErrInvalidOrderData:       newHandler(http.StatusBadRequest, "Invalid order data"),

	// Review
	ErrInvalidRating:          newHandler(http.StatusBadRequest, "Invalid rating of review"),
	ErrInvalidReviewSortBy:    newHandler(http.StatusBadRequest, "Invalid sort_by for review filter"),
	ErrInvalidReviewSortOrder: newHandler(http.StatusBadRequest, "Invalid sort_order for review filter"),
	ErrInvalidReviewID:        newHandler(http.StatusBadRequest, "Invalid id of review"),
	ErrNothingToUpdate:        newHandler(http.StatusBadRequest, "Not enough parameters for update"),
	ErrInvalidComment:         newHandler(http.StatusBadRequest, "Comment too long (max 1000 characters)"),
	ErrReviewNotOwnedByUser:   newHandler(http.StatusForbidden, "Review does not belong to user"),

	// Avatar
	ErrAvatarNotFound:    newHandler(http.StatusNotFound, "Avatar not found"),
	ErrInvalidAvatarData: newHandler(http.StatusBadRequest, "Invalid avatar data"),
	ErrAvatarUploadFail:  newHandler(http.StatusInternalServerError, "Failed to upload avatar"),
}

const defaultErrorMessage = "Internal server error"

func HandleError(c *gin.Context, err error) {
	for target, handler := range errorHandlers {
		if errors.Is(err, target) {
			c.JSON(handler.status, gin.H{"error": handler.message})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": defaultErrorMessage})
}

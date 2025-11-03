package httpadapter

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"

    errors2 "goshop/internal/core/domain/errors"
)

// HandleDomainError maps domain/service errors to HTTP responses.
func HandleDomainError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, errors2.ErrNotFound),
        errors.Is(err, errors2.ErrUserNotFound),
        errors.Is(err, errors2.ErrAddressNotFound),
        errors.Is(err, errors2.ErrCategoryNotFound),
        errors.Is(err, errors2.ErrProductNotFound),
        errors.Is(err, errors2.ErrCartNotFound),
        errors.Is(err, errors2.ErrCartItemNotFound),
        errors.Is(err, errors2.ErrOrderNotFound),
        errors.Is(err, errors2.ErrOrderItemNotFound),
        errors.Is(err, errors2.ErrReviewNotOwnedByUser),
        errors.Is(err, errors2.ErrAvatarNotFound),
        errors.Is(err, errors2.ErrProductImageNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case errors.Is(err, errors2.ErrUnauthorized),
        errors.Is(err, errors2.ErrInvalidPassword):
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
    case errors.Is(err, errors2.ErrForbidden),
        errors.Is(err, errors2.ErrOrderNotOwnedByUser),
        errors.Is(err, errors2.ErrReviewNotOwnedByUser):
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
    case errors.Is(err, errors2.ErrDuplicate),
        errors.Is(err, errors2.ErrEmailExists),
        errors.Is(err, errors2.ErrCategoryNameExists),
        errors.Is(err, errors2.ErrProductNameExists),
        errors.Is(err, errors2.ErrCartItemExists),
        errors.Is(err, errors2.ErrOrderAlreadyCancelled):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    case errors.Is(err, errors2.ErrInvalidInput),
        errors.Is(err, errors2.ErrInvalidAddressData),
        errors.Is(err, errors2.ErrInvalidCategoryData),
        errors.Is(err, errors2.ErrInvalidProductData),
        errors.Is(err, errors2.ErrInvalidQuantity),
        errors.Is(err, errors2.ErrInvalidPrice),
        errors.Is(err, errors2.ErrInvalidStock),
        errors.Is(err, errors2.ErrInvalidOrderStatus),
        errors.Is(err, errors2.ErrInvalidRating),
        errors.Is(err, errors2.ErrInvalidReviewID),
        errors.Is(err, errors2.ErrInvalidProductID),
        errors.Is(err, errors2.ErrInvalidUserID),
        errors.Is(err, errors2.ErrNothingToUpdate),
        errors.Is(err, errors2.ErrInvalidComment),
        errors.Is(err, errors2.ErrInvalidEmail):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    case errors.Is(err, errors2.ErrInsufficientStock),
        errors.Is(err, errors2.ErrProductNotAvailable),
        errors.Is(err, errors2.ErrCartEmpty),
        errors.Is(err, errors2.ErrEmptyCart),
        errors.Is(err, errors2.ErrOrderCannotBeCancelled):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
    }
}

package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "goshop/internal/core/domain/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var errorStatusMap = map[error]int{
	// 404 Not Found
	domainErrors.ErrNotFound:             http.StatusNotFound,
	domainErrors.ErrUserNotFound:         http.StatusNotFound,
	domainErrors.ErrAddressNotFound:      http.StatusNotFound,
	domainErrors.ErrCategoryNotFound:     http.StatusNotFound,
	domainErrors.ErrProductNotFound:      http.StatusNotFound,
	domainErrors.ErrCartNotFound:         http.StatusNotFound,
	domainErrors.ErrCartItemNotFound:     http.StatusNotFound,
	domainErrors.ErrOrderNotFound:        http.StatusNotFound,
	domainErrors.ErrOrderItemNotFound:    http.StatusNotFound,
	domainErrors.ErrAvatarNotFound:       http.StatusNotFound,
	domainErrors.ErrProductImageNotFound: http.StatusNotFound,

	// 401 Unauthorized
	domainErrors.ErrUnauthorized:    http.StatusUnauthorized,
	domainErrors.ErrInvalidPassword: http.StatusUnauthorized,

	// 403 Forbidden
	domainErrors.ErrForbidden:            http.StatusForbidden,
	domainErrors.ErrOrderNotOwnedByUser:  http.StatusForbidden,
	domainErrors.ErrReviewNotOwnedByUser: http.StatusForbidden,

	// 400 Bad Request
	domainErrors.ErrInvalidInput:        http.StatusBadRequest,
	domainErrors.ErrInvalidAddressData:  http.StatusBadRequest,
	domainErrors.ErrInvalidCategoryData: http.StatusBadRequest,
	domainErrors.ErrInvalidProductData:  http.StatusBadRequest,
	domainErrors.ErrInvalidQuantity:     http.StatusBadRequest,
	domainErrors.ErrInvalidPrice:        http.StatusBadRequest,
	domainErrors.ErrInvalidStock:        http.StatusBadRequest,
	domainErrors.ErrInvalidOrderStatus:  http.StatusBadRequest,
	domainErrors.ErrInvalidRating:       http.StatusBadRequest,
	domainErrors.ErrInvalidReviewID:     http.StatusBadRequest,
	domainErrors.ErrInvalidProductID:    http.StatusBadRequest,
	domainErrors.ErrInvalidUserID:       http.StatusBadRequest,
	domainErrors.ErrNothingToUpdate:     http.StatusBadRequest,
	domainErrors.ErrInvalidComment:      http.StatusBadRequest,
	domainErrors.ErrInvalidEmail:        http.StatusBadRequest,

	// 409 Conflict
	domainErrors.ErrDuplicate:              http.StatusConflict,
	domainErrors.ErrEmailExists:            http.StatusConflict,
	domainErrors.ErrCategoryNameExists:     http.StatusConflict,
	domainErrors.ErrProductNameExists:      http.StatusConflict,
	domainErrors.ErrCartItemExists:         http.StatusConflict,
	domainErrors.ErrOrderAlreadyCancelled:  http.StatusConflict,
	domainErrors.ErrInsufficientStock:      http.StatusConflict,
	domainErrors.ErrProductNotAvailable:    http.StatusConflict,
	domainErrors.ErrCartEmpty:              http.StatusConflict,
	domainErrors.ErrEmptyCart:              http.StatusConflict,
	domainErrors.ErrOrderCannotBeCancelled: http.StatusConflict,
}


func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}


	for domainErr, statusCode := range errorStatusMap {
		if errors.Is(err, domainErr) {
			c.JSON(statusCode, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

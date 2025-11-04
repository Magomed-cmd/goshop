package httpadapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
	errors2 "goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type ReviewHandler struct {
	reviewService serviceports.ReviewService
}

func NewReviewHandler(reviewService serviceports.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview godoc
// @Summary     Create review
// @Description Creates a review for a product by the authenticated user
// @Tags        reviews
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateReviewRequest true "Review payload"
// @Success     201 {object} dto.ReviewResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	req := dto.CreateReviewRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	resp, err := h.reviewService.CreateReview(ctx, &req, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetReviews godoc
// @Summary     List reviews
// @Description Returns reviews with optional filters
// @Tags        reviews
// @Produce     json
// @Param       page       query int    false "Page number"
// @Param       limit      query int    false "Page size"
// @Param       product_id query int    false "Filter by product ID"
// @Param       user_id    query int    false "Filter by user ID"
// @Param       rating     query int    false "Filter by rating"
// @Param       sort_by    query string false "Sort field"
// @Param       sort_order query string false "Sort order"
// @Success     200 {object} dto.ReviewsListResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /reviews [get]
func (h *ReviewHandler) GetReviews(c *gin.Context) {
	ctx := c.Request.Context()

	var filters types.ReviewFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	resp, err := h.reviewService.GetReviewsWithFilters(ctx, filters)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetReviewByID godoc
// @Summary     Get review
// @Description Returns review details by ID
// @Tags        reviews
// @Produce     json
// @Param       id path int true "Review ID"
// @Success     200 {object} dto.ReviewResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /reviews/{id} [get]
func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	ctx := c.Request.Context()

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil || reviewID <= 0 {
		httpErrors.HandleError(c, errors2.ErrInvalidReviewID)
		return
	}

	resp, err := h.reviewService.GetReviewByID(ctx, reviewID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateReview godoc
// @Summary     Update review
// @Description Updates a review created by the authenticated user
// @Tags        reviews
// @Accept      json
// @Produce     json
// @Param       id      path int                    true "Review ID"
// @Param       request body dto.UpdateReviewRequest true "Review payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil || reviewID <= 0 {
		httpErrors.HandleError(c, errors2.ErrInvalidReviewID)
		return
	}

	req := dto.UpdateReviewRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err = h.reviewService.UpdateReview(ctx, userID, reviewID, req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review updated successfully"})
}

// DeleteReview godoc
// @Summary     Delete review
// @Description Deletes a review created by the authenticated user
// @Tags        reviews
// @Produce     json
// @Param       id path int true "Review ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil || reviewID <= 0 {
		httpErrors.HandleError(c, errors2.ErrInvalidReviewID)
		return
	}

	err = h.reviewService.DeleteReview(ctx, userID, reviewID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

// GetProductReviewStats godoc
// @Summary     Review statistics
// @Description Returns aggregate review statistics for a product
// @Tags        reviews
// @Produce     json
// @Param       productId path int true "Product ID"
// @Success     200 {object} dto.ReviewStatsResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /reviews/stats/{productId} [get]
func (h *ReviewHandler) GetProductReviewStats(c *gin.Context) {
	ctx := c.Request.Context()

	productIDStr := c.Param("productId")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		httpErrors.HandleError(c, errors2.ErrInvalidProductID)
		return
	}

	resp, err := h.reviewService.GetReviewStats(ctx, productID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

package httpadapter

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	errors2 "goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type ReviewService interface {
	CreateReview(ctx context.Context, req *dto.CreateReviewRequest, userID int64) (*dto.ReviewResponse, error)
	GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) (*dto.ReviewsListResponse, error)
	GetReviewByID(ctx context.Context, reviewID int64) (*dto.ReviewResponse, error)
	UpdateReview(ctx context.Context, userID int64, reviewID int64, req dto.UpdateReviewRequest) error
	DeleteReview(ctx context.Context, userID int64, reviewID int64) error
	GetReviewStats(ctx context.Context, productID int64) (*dto.ReviewStatsResponse, error)
}

type ReviewHandler struct {
	reviewService ReviewService
}

func NewReviewHandler(reviewService ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

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
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ReviewHandler) GetReviews(c *gin.Context) {
	ctx := c.Request.Context()

	var filters types.ReviewFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	resp, err := h.reviewService.GetReviewsWithFilters(ctx, filters)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	ctx := c.Request.Context()

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil || reviewID <= 0 {
		HandleDomainError(c, errors2.ErrInvalidReviewID)
		return
	}

	resp, err := h.reviewService.GetReviewByID(ctx, reviewID)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

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
		HandleDomainError(c, errors2.ErrInvalidReviewID)
		return
	}

	req := dto.UpdateReviewRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err = h.reviewService.UpdateReview(ctx, userID, reviewID, req)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review updated successfully"})
}

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
		HandleDomainError(c, errors2.ErrInvalidReviewID)
		return
	}

	err = h.reviewService.DeleteReview(ctx, userID, reviewID)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

func (h *ReviewHandler) GetProductReviewStats(c *gin.Context) {
	ctx := c.Request.Context()

	productIDStr := c.Param("productId")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		HandleDomainError(c, errors2.ErrInvalidProductID)
		return
	}

	resp, err := h.reviewService.GetReviewStats(ctx, productID)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

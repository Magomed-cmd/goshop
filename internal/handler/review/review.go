package handler

import (
	"context"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID int64, req *dto.CreateReviewRequest) (*dto.ReviewResponse, error)
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

	resp, err := h.reviewService.CreateReview(ctx, userID, &req)
	if err != nil {
		domain_errors.HandleError(c, err)
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
		domain_errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	ctx := c.Request.Context()

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil || reviewID <= 0 {
		domain_errors.HandleError(c, domain_errors.ErrInvalidReviewID)
		return
	}

	resp, err := h.reviewService.GetReviewByID(ctx, reviewID)
	if err != nil {
		domain_errors.HandleError(c, err)
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
		domain_errors.HandleError(c, domain_errors.ErrInvalidReviewID)
		return
	}

	req := dto.UpdateReviewRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	err = h.reviewService.UpdateReview(ctx, userID, reviewID, req)
	if err != nil {
		domain_errors.HandleError(c, err)
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
		domain_errors.HandleError(c, domain_errors.ErrInvalidReviewID)
		return
	}

	err = h.reviewService.DeleteReview(ctx, userID, reviewID)
	if err != nil {
		domain_errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

func (h *ReviewHandler) GetProductReviewStats(c *gin.Context) {
	ctx := c.Request.Context()

	productIDStr := c.Param("productId")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		domain_errors.HandleError(c, domain_errors.ErrInvalidProductID)
		return
	}

	resp, err := h.reviewService.GetReviewStats(ctx, productID)
	if err != nil {
		domain_errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

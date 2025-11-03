package services

import (
	"context"

	"goshop/internal/core/domain/types"
	"goshop/internal/dto"
)

type ReviewService interface {
	CreateReview(ctx context.Context, req *dto.CreateReviewRequest, userID int64) (*dto.ReviewResponse, error)
	GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) (*dto.ReviewsListResponse, error)
	GetReviewByID(ctx context.Context, reviewID int64) (*dto.ReviewResponse, error)
	UpdateReview(ctx context.Context, userID int64, reviewID int64, req dto.UpdateReviewRequest) error
	DeleteReview(ctx context.Context, userID int64, reviewID int64) error
	GetReviewStats(ctx context.Context, productID int64) (*dto.ReviewStatsResponse, error)
}

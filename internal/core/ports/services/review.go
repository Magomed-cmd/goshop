package services

import (
	"context"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/types"
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID int64, productID int64, rating int, comment *string) (*entities.Review, error)
	GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) ([]*entities.Review, int64, error)
	GetReviewByID(ctx context.Context, reviewID int64) (*entities.Review, error)
	UpdateReview(ctx context.Context, userID int64, reviewID int64, rating *int, comment *string) error
	DeleteReview(ctx context.Context, userID int64, reviewID int64) error
	GetReviewStats(ctx context.Context, productID int64) (int64, float64, map[int]int64, error)
}

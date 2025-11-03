package repositories

import (
    "context"

    "goshop/internal/core/domain/entities"
    "goshop/internal/core/domain/types"
)

type ReviewRepository interface {
    CreateReview(ctx context.Context, review *entities.Review) (*int64, error)
    GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) ([]*entities.Review, int64, error)
    GetReviewByID(ctx context.Context, reviewID int64) (*entities.Review, error)
    UpdateReview(ctx context.Context, reviewID int64, rating *int, comment *string) error
    DeleteReview(ctx context.Context, reviewID int64) error
    CheckUserReviewExists(ctx context.Context, userID, productID int64) (bool, error)
    GetReviewStats(ctx context.Context, productID int64) (totalReviews int64, averageRating float64, ratingCounts map[int]int64, err error)
}


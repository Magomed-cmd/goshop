package cache

import (
    "context"
    "time"

    "goshop/internal/dto"
)

type ReviewCache interface {
    SetReviewByID(ctx context.Context, reviewID int64, reviewResponse *dto.ReviewResponse, ttl time.Duration) error
    GetReviewByID(ctx context.Context, reviewID int64) (*dto.ReviewResponse, error)
    InvalidateReview(ctx context.Context, reviewID int64) error
}


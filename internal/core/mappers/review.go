package mappers

import (
    "goshop/internal/core/domain/entities"
    "goshop/internal/dto"
)

func ToReviewResponse(review *entities.Review) dto.ReviewResponse {
    if review == nil {
        return dto.ReviewResponse{}
    }

    resp := dto.ReviewResponse{
        ID:        review.ID,
        UUID:      review.UUID.String(),
        ProductID: review.ProductID,
        UserID:    review.UserID,
        Rating:    review.Rating,
        Comment:   review.Comment,
        CreatedAt: review.CreatedAt,
    }

    if review.User != nil {
        resp.User = &dto.UserInfo{
            UUID: review.User.UUID.String(),
            Name: review.User.Name,
        }
    }

    if review.Product != nil {
        resp.Product = &dto.ProductInfo{
            UUID: review.Product.UUID.String(),
            Name: review.Product.Name,
        }
    }

    return resp
}

func ToReviewsListResponse(reviews []*entities.Review, totalCount int64, page, limit int) *dto.ReviewsListResponse {
    responses := make([]dto.ReviewResponse, 0, len(reviews))
    var sum int

    for _, review := range reviews {
        if review == nil {
            continue
        }
        responses = append(responses, ToReviewResponse(review))
        sum += review.Rating
    }

    var avg *float64
    if len(responses) > 0 {
        average := float64(sum) / float64(len(responses))
        avg = &average
    }

    return &dto.ReviewsListResponse{
        Reviews:       responses,
        TotalCount:    totalCount,
        Page:          page,
        Limit:         limit,
        AverageRating: avg,
    }
}

func ToReviewStatsResponse(totalReviews int64, averageRating float64, counts map[int]int64) *dto.ReviewStatsResponse {
    return &dto.ReviewStatsResponse{
        TotalReviews:  totalReviews,
        AverageRating: averageRating,
        RatingCounts:  counts,
    }
}

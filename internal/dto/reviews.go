package dto

import (
	"time"
)

type CreateReviewRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Rating    int     `json:"rating" binding:"required,min=1,max=5"`
	Comment   *string `json:"comment" binding:"omitempty,max=1000"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment *string `json:"comment" binding:"omitempty,max=1000"`
}

type ReviewResponse struct {
	ID        int64     `json:"id"`
	UUID      string    `json:"uuid"`
	ProductID int64     `json:"product_id"`
	UserID    int64     `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`

	User    *UserInfo    `json:"user,omitempty"`
	Product *ProductInfo `json:"product,omitempty"`
}

type UserInfo struct {
	UUID string  `json:"uuid"`
	Name *string `json:"name"`
}

type ProductInfo struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type ReviewsListResponse struct {
	Reviews       []ReviewResponse `json:"reviews"`
	TotalCount    int64            `json:"total_count"`
	Page          int              `json:"page"`
	Limit         int              `json:"limit"`
	AverageRating *float64         `json:"average_rating,omitempty"`
}

type ReviewStatsResponse struct {
	TotalReviews  int64         `json:"total_reviews"`
	AverageRating float64       `json:"average_rating"`
	RatingCounts  map[int]int64 `json:"rating_counts"`
}

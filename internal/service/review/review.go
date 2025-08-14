package review

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"goshop/internal/validation"
	"time"

	"github.com/google/uuid"
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

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
}
type ProductRepository interface {
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
}

type ReviewService struct {
	reviewRepo  ReviewRepository
	userRepo    UserRepository
	productRepo ProductRepository
}

func NewReviewsService(reviewRepo ReviewRepository, userRepository UserRepository, productRepository ProductRepository) *ReviewService {
	return &ReviewService{
		reviewRepo:  reviewRepo,
		userRepo:    userRepository,
		productRepo: productRepository,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, req dto.CreateReviewRequest, userID int64) (*dto.ReviewResponse, error) {

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain_errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	review := &entities.Review{
		UUID:      uuid.New(),
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
		User:      user,
		Product:   product,
	}

	reviewID, err := s.reviewRepo.CreateReview(ctx, review)
	if err != nil {
		return nil, err
	}

	resp := dto.ReviewResponse{
		ID:        *reviewID,
		UUID:      review.UUID.String(),
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: review.CreatedAt,
		User: &dto.UserInfo{
			UUID: user.UUID.String(),
			Name: user.Name,
		},
		Product: &dto.ProductInfo{
			UUID: product.UUID.String(),
			Name: product.Name,
		},
	}

	return &resp, nil
}

func (s *ReviewService) GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) (*dto.ReviewsListResponse, error) {

	if err := validation.ValidateReviewFilters(filters); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, *filters.UserID)
	if err != nil {
		return nil, domain_errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, *filters.ProductID)
	if err != nil {
		return nil, err
	}

	reviews, totalCount, err := s.reviewRepo.GetReviewsWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	var avgRating float64
	respReviews := make([]dto.ReviewResponse, 0, 10)

	for _, review := range reviews {
		respReview := dto.ReviewResponse{
			ID:        review.ID,
			UUID:      review.UUID.String(),
			ProductID: review.ProductID,
			UserID:    review.UserID,
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: time.Now(),
			User: &dto.UserInfo{
				UUID: user.UUID.String(),
				Name: user.Name,
			},
			Product: &dto.ProductInfo{
				UUID: product.UUID.String(),
				Name: product.Name,
			},
		}

		respReviews = append(respReviews, respReview)

		avgRating += float64(review.Rating)
	}
	avgRating = avgRating / float64(len(reviews))

	resp := &dto.ReviewsListResponse{
		Reviews:       respReviews,
		TotalCount:    totalCount,
		Page:          filters.Page,
		Limit:         filters.Limit,
		AverageRating: &avgRating,
	}

	return resp, nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID int64, req *dto.UpdateReviewRequest) (*dto.ReviewResponse, error) {

	if reviewID < 0 {
		return nil, domain_errors.ErrInvalidReviewID
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, review.UserID)
	if err != nil {
		return nil, domain_errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, review.ProductID)
	if err != nil {
		return nil, err
	}

	resp := &dto.ReviewResponse{
		ID:        review.ID,
		UUID:      review.UUID.String(),
		ProductID: review.ProductID,
		UserID:    review.UserID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
		User: &dto.UserInfo{
			UUID: user.UUID.String(),
			Name: user.Name,
		},
		Product: &dto.ProductInfo{
			UUID: product.UUID.String(),
			Name: product.Name,
		},
	}

	return resp, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, reviewID int64, req dto.UpdateReviewRequest) error {

	if req.Rating == nil && req.Comment == nil {
		return domain_errors.ErrNothingToUpdate
	}

	if req.Rating != nil && (*req.Rating > 5 || *req.Rating < 1) {
		return domain_errors.ErrInvalidRating
	}

	if req.Comment != nil && len(*req.Comment) > 1000 {
		return domain_errors.ErrInvalidComment
	}

	err := s.reviewRepo.UpdateReview(ctx, reviewID, req.Rating, req.Comment)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID int64, reviewID int64) error {

	if reviewID < 1 {
		return domain_errors.ErrInvalidReviewID
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		return domain_errors.ErrReviewNotOwnedByUser
	}

	err = s.reviewRepo.DeleteReview(ctx, reviewID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReviewService) CheckUserReviewExists(ctx context.Context, userID, productID int64) (bool, error) {
	if userID < 1 {
		return false, domain_errors.ErrInvalidUserID
	}

	if productID < 1 {
		return false, domain_errors.ErrInvalidProductID
	}

	exists, err := s.reviewRepo.CheckUserReviewExists(ctx, userID, productID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *ReviewService) GetReviewStats(ctx context.Context, productID int64) (*dto.ReviewStatsResponse, error) {
	if productID < 1 {
		return nil, domain_errors.ErrInvalidProductID
	}

	totalReviews, averageRating, ratingCounts, err := s.reviewRepo.GetReviewStats(ctx, productID)
	if err != nil {
		return nil, err
	}

	response := &dto.ReviewStatsResponse{
		TotalReviews:  totalReviews,
		AverageRating: averageRating,
		RatingCounts:  ratingCounts,
	}

	return response, nil
}

package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	cacheports "goshop/internal/core/ports/cache"
	repositories "goshop/internal/core/ports/repositories"
	"goshop/internal/validation"
)

const (
	reviewCacheTTL = 5 * time.Minute
)

type ReviewService struct {
	reviewRepo  repositories.ReviewRepository
	reviewCache cacheports.ReviewCache
	userRepo    repositories.UserRepository
	productRepo repositories.ProductRepository
	logger      *zap.Logger
}

func NewReviewsService(reviewRepo repositories.ReviewRepository, userRepository repositories.UserRepository, productRepository repositories.ProductRepository, cache cacheports.ReviewCache, logger *zap.Logger) *ReviewService {
	return &ReviewService{
		reviewRepo:  reviewRepo,
		reviewCache: cache,
		userRepo:    userRepository,
		productRepo: productRepository,
		logger:      logger,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID int64, productID int64, rating int, comment *string) (*entities.Review, error) {

	if userID < 1 {
		return nil, errors.ErrInvalidUserID
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	review := &entities.Review{
		UUID:      uuid.New(),
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
		User:      user,
		Product:   product,
	}

	reviewID, err := s.reviewRepo.CreateReview(ctx, review)
	if err != nil {
		return nil, err
	}

	review.ID = *reviewID

	return review, nil
}

func (s *ReviewService) GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) ([]*entities.Review, int64, error) {

	if filters.UserID == nil || filters.ProductID == nil {
		return nil, 0, errors.ErrInvalidInput
	}

	if err := validation.ValidateReviewFilters(filters); err != nil {
		return nil, 0, err
	}

	user, err := s.userRepo.GetUserByID(ctx, *filters.UserID)
	if err != nil {
		return nil, 0, errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, *filters.ProductID)
	if err != nil {
		return nil, 0, err
	}

	reviews, totalCount, err := s.reviewRepo.GetReviewsWithFilters(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	for _, review := range reviews {
		if review == nil {
			continue
		}
		if review.User == nil {
			review.User = user
		}
		if review.Product == nil {
			review.Product = product
		}
	}

	return reviews, totalCount, nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID int64) (*entities.Review, error) {

	if reviewID < 1 {
		return nil, errors.ErrInvalidReviewID
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, review.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, review.ProductID)
	if err != nil {
		return nil, err
	}

	review.User = user
	review.Product = product

	return review, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, userID int64, reviewID int64, rating *int, comment *string) error {
	if rating == nil && comment == nil {
		return errors.ErrNothingToUpdate
	}

	if rating != nil && (*rating > 5 || *rating < 1) {
		return errors.ErrInvalidRating
	}

	if comment != nil && len(*comment) > 1000 {
		return errors.ErrInvalidComment
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review.UserID != userID {
		return errors.ErrForbidden
	}

	err = s.reviewRepo.UpdateReview(ctx, reviewID, rating, comment)
	if err != nil {
		return err
	}

	if err := s.reviewCache.InvalidateReview(ctx, reviewID); err != nil {
		s.logger.Warn("failed to invalidate review cache",
			zap.Int64("reviewID", reviewID),
			zap.Error(err))
	}

	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID int64, reviewID int64) error {

	if reviewID < 1 {
		return errors.ErrInvalidReviewID
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		return errors.ErrReviewNotOwnedByUser
	}

	err = s.reviewRepo.DeleteReview(ctx, reviewID)
	if err != nil {
		return err
	}

	err = s.reviewCache.InvalidateReview(ctx, reviewID)
	if err != nil {
		s.logger.Warn("failed to invalidate review cache", zap.Int64("reviewID", reviewID), zap.Error(err))
	}

	return nil
}

func (s *ReviewService) CheckUserReviewExists(ctx context.Context, userID, productID int64) (bool, error) {
	if userID < 1 {
		return false, errors.ErrInvalidUserID
	}

	if productID < 1 {
		return false, errors.ErrInvalidProductID
	}

	exists, err := s.reviewRepo.CheckUserReviewExists(ctx, userID, productID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *ReviewService) GetReviewStats(ctx context.Context, productID int64) (int64, float64, map[int]int64, error) {
	if productID < 1 {
		return 0, 0, nil, errors.ErrInvalidProductID
	}

	totalReviews, averageRating, ratingCounts, err := s.reviewRepo.GetReviewStats(ctx, productID)
	if err != nil {
		return 0, 0, nil, err
	}

	return totalReviews, averageRating, ratingCounts, nil
}

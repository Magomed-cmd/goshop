package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/types"
	"goshop/internal/core/mappers"
	cacheports "goshop/internal/core/ports/cache"
	repositories "goshop/internal/core/ports/repositories"
	"goshop/internal/dto"
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

func (s *ReviewService) CreateReview(ctx context.Context, req *dto.CreateReviewRequest, userID int64) (*dto.ReviewResponse, error) {

	if userID < 1 {
		return nil, errors.ErrInvalidUserID
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
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

	review.ID = *reviewID

	resp := mappers.ToReviewResponse(review)

	return &resp, nil
}

func (s *ReviewService) GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) (*dto.ReviewsListResponse, error) {

	if filters.UserID == nil || filters.ProductID == nil {
		return nil, errors.ErrInvalidInput
	}

	if err := validation.ValidateReviewFilters(filters); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, *filters.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	product, err := s.productRepo.GetProductByID(ctx, *filters.ProductID)
	if err != nil {
		return nil, err
	}

	reviews, totalCount, err := s.reviewRepo.GetReviewsWithFilters(ctx, filters)
	if err != nil {
		return nil, err
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

	return mappers.ToReviewsListResponse(reviews, totalCount, filters.Page, filters.Limit), nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID int64) (*dto.ReviewResponse, error) {

	if reviewID < 1 {
		return nil, errors.ErrInvalidReviewID
	}

	cacheReview, err := s.reviewCache.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.logger.Warn("failed to get review from cache, fallback to repository",
			zap.Int64("reviewID", reviewID), zap.Error(err))
	} else if cacheReview != nil {
		s.logger.Debug("found review in cache", zap.Int64("reviewID", reviewID))
		return cacheReview, nil
	}
	s.logger.Debug("review not found in cache, fetching from repository", zap.Int64("reviewID", reviewID))

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

	resp := mappers.ToReviewResponse(review)

	if err := s.reviewCache.SetReviewByID(ctx, reviewID, &resp, reviewCacheTTL); err != nil {
		s.logger.Warn("failed to set review in cache", zap.Int64("reviewID", reviewID), zap.Error(err))
	}

	return &resp, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, userID int64, reviewID int64, req dto.UpdateReviewRequest) error {
	if req.Rating == nil && req.Comment == nil {
		return errors.ErrNothingToUpdate
	}

	if req.Rating != nil && (*req.Rating > 5 || *req.Rating < 1) {
		return errors.ErrInvalidRating
	}

	if req.Comment != nil && len(*req.Comment) > 1000 {
		return errors.ErrInvalidComment
	}

	review, err := s.reviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review.UserID != userID {
		return errors.ErrForbidden
	}

	err = s.reviewRepo.UpdateReview(ctx, reviewID, req.Rating, req.Comment)
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

func (s *ReviewService) GetReviewStats(ctx context.Context, productID int64) (*dto.ReviewStatsResponse, error) {
	if productID < 1 {
		return nil, errors.ErrInvalidProductID
	}

	totalReviews, averageRating, ratingCounts, err := s.reviewRepo.GetReviewStats(ctx, productID)
	if err != nil {
		return nil, err
	}

	return mappers.ToReviewStatsResponse(totalReviews, averageRating, ratingCounts), nil
}

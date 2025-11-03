package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goshop/internal/dto"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	ReviewCachePatternID = "goshop:review:%d"
)

type ReviewCache struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func NewReviewCache(redisClient *redis.Client, logger *zap.Logger) *ReviewCache {
	return &ReviewCache{
		rdb:    redisClient,
		logger: logger,
	}
}

func ReviewIDKey(reviewID int64) string {
	return fmt.Sprintf(ReviewCachePatternID, reviewID)
}

func (c *ReviewCache) SetReviewByID(ctx context.Context, reviewID int64, reviewResponse *dto.ReviewResponse, ttl time.Duration) error {
	key := ReviewIDKey(reviewID)

	jsonBytes, err := json.Marshal(reviewResponse)
	if err != nil {
		c.logger.Error("failed to marshal review response", zap.Error(err), zap.Int64("reviewID", reviewID))
		return err
	}

	if err := c.rdb.Set(ctx, key, string(jsonBytes), ttl).Err(); err != nil {
		c.logger.Error("failed to set review in cache", zap.Error(err), zap.Int64("reviewID", reviewID))
		return err
	}

	return nil
}

func (c *ReviewCache) GetReviewByID(ctx context.Context, reviewID int64) (*dto.ReviewResponse, error) {

	key := ReviewIDKey(reviewID)
	cmd := c.rdb.Get(ctx, key)

	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Info("review not found in cache", zap.Int64("reviewID", reviewID))
			return nil, nil
		}
		c.logger.Error("failed to get review from cache", zap.Error(err), zap.Int64("reviewID", reviewID))
		return nil, err
	}

	reviewResponse := &dto.ReviewResponse{}
	jsonBytes, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(jsonBytes), reviewResponse); err != nil {
		c.logger.Error("failed to unmarshal review response", zap.Error(err), zap.Int64("reviewID", reviewID))
		return nil, err
	}

	return reviewResponse, nil
}

func (c *ReviewCache) InvalidateReview(ctx context.Context, reviewID int64) error {
	key := ReviewIDKey(reviewID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.logger.Error("failed to invalidate review cache", zap.Error(err), zap.Int64("reviewID", reviewID))
		return err
	}
	return nil
}

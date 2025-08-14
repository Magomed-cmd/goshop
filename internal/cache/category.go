package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goshop/internal/dto"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CategoryCache struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func NewCategoryCache(redisClient *redis.Client, logger *zap.Logger) *CategoryCache {
	return &CategoryCache{
		rdb:    redisClient,
		logger: logger,
	}
}

func CategoryKey(categoryID int64) string {
	// goshop:category:1..2...
	return fmt.Sprintf("goshop:category:%d", categoryID)
}

func (c *CategoryCache) SetCategory(ctx context.Context, category *dto.CategoryResponse, ttl time.Duration) error {

	key := CategoryKey(category.ID)

	jsonBytes, err := json.Marshal(category)
	if err != nil {
		return err
	}

	result := c.rdb.Set(ctx, key, string(jsonBytes), ttl)
	if err := result.Err(); err != nil {
		return result.Err()
	}

	return err
}

func (c *CategoryCache) GetCategory(ctx context.Context, categoryID int64) (*dto.CategoryResponse, error) {

	key := CategoryKey(categoryID)

	categoryData := c.rdb.Get(ctx, key)
	if err := categoryData.Err(); err != nil {
		c.logger.Error("Failed to get category data from Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	categoryStr, err := categoryData.Result()
	if err != nil {
		c.logger.Error("Failed to get category string result from Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	category := &dto.CategoryResponse{}
	err = json.Unmarshal([]byte(categoryStr), category)
	if err != nil {
		c.logger.Error("Failed to unmarshal category from cache",
			zap.String("categoryStr", categoryStr),
			zap.Error(err),
		)
		return nil, err
	}

	return category, nil
}

//func (c *CategoryCache) GetAllCategories(ctx context.Context, )

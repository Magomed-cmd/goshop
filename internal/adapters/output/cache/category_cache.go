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
	keyPatternByID = "goshop:category:%d"
	KeyPatternAll  = "goshop:category:all"
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

func (c *CategoryCache) SetCategory(ctx context.Context, category *dto.CategoryResponse, ttl time.Duration) error {

	key := CategoryKeyID(category.ID)

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

	key := CategoryKeyID(categoryID)

	categoryData := c.rdb.Get(ctx, key)
	if err := categoryData.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		c.logger.Warn("Failed to get category data from Redis",
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

func (c *CategoryCache) SetAllCategories(ctx context.Context, response *dto.CategoriesListResponse, ttl time.Duration) error {

	allCategoriesJsonBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	allCategoriesStr := string(allCategoriesJsonBytes)

	query := c.rdb.Set(ctx, KeyPatternAll, allCategoriesStr, ttl)
	if err = query.Err(); err != nil {
		return err
	}

	return nil
}

func (c *CategoryCache) GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error) {

	categoryData := c.rdb.Get(ctx, KeyPatternAll)
	if err := categoryData.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	jsonStr, err := categoryData.Result()
	if err != nil {
		return nil, err
	}

	categoryResp := &dto.CategoriesListResponse{}
	if err = json.Unmarshal([]byte(jsonStr), categoryResp); err != nil {
		return nil, err
	}

	return categoryResp, nil
}

func (c *CategoryCache) DeleteAllCategories(ctx context.Context) error {
	result := c.rdb.Del(ctx, KeyPatternAll)
	if err := result.Err(); err != nil {
		return err
	}

	c.logger.Info("All categories deleted from cache", zap.String("key", KeyPatternAll))

	return nil
}

func (c *CategoryCache) DeleteCategory(ctx context.Context, categoryID int64) error {
	key := CategoryKeyID(categoryID)

	result := c.rdb.Del(ctx, key)
	if err := result.Err(); err != nil {
		return err
	}
	c.logger.Info("Category deleted from cache",
		zap.String("key", key),
		zap.Int64("categoryID", categoryID),
	)
	return nil
}

func CategoryKeyID(categoryID int64) string {
	return fmt.Sprintf(keyPatternByID, categoryID)
}

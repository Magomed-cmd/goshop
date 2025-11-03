package redisDB

import (
	"context"
	"errors"
	"fmt"
	"goshop/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewConnection(cfg *config.RedisConfig, logger *zap.Logger) (*redis.Client, error) {

	ctx := context.Background()

	logger.Info("Start connecting to Redis")
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
	)

	err := rdb.Ping(ctx).Err()
	if err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}
	logger.Info("Successfully connected to Redis", zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))

	return rdb, nil
}

/*
Get(ctx, client, key) (string, error)
Set(ctx, client, key, value, ttl) error
Del(ctx, client, key) error
*/

func GetData(ctx context.Context, rdb *redis.Client, key string) (*string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func SetData(ctx context.Context, rdb *redis.Client, key, value string, ttl time.Duration) error {
	err := rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func Exists(ctx context.Context, rdb *redis.Client, key string) (bool, error) {
	count, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

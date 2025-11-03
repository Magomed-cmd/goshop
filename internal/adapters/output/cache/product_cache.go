package cache

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"

    "goshop/internal/core/domain/types"
    "goshop/internal/dto"
)

type ProductCache struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func NewProductCache(redisClient *redis.Client, logger *zap.Logger) *ProductCache {
	return &ProductCache{
		rdb:    redisClient,
		logger: logger,
	}
}

func ProductsFilterKey(filters types.ProductFilters) (string, error) {
	// goshop:products:filter:{"category_id":5,"page":1,"limit":20}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("goshop:products:filter:%s", string(filtersJSON)), nil
}

func ProductKey(productID int64) string {
	// goshop:product:123
	return fmt.Sprintf("goshop:product:%d", productID)
}

func (pc *ProductCache) SetProduct(ctx context.Context, product *dto.ProductResponse, ttl time.Duration) error {
	jsonBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}

	productKey := ProductKey(product.ID)
	return pc.rdb.Set(ctx, productKey, string(jsonBytes), ttl).Err()
}

func (pc *ProductCache) GetProduct(ctx context.Context, productID int64) (*dto.ProductResponse, error) {

	key := ProductKey(productID)
	res := pc.rdb.Get(ctx, key)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	jsonString, err := res.Result()
	if err != nil {
		return nil, err
	}

	product := &dto.ProductResponse{}
	if err := json.Unmarshal([]byte(jsonString), product); err != nil {
		return nil, err
	}

	return product, nil
}

func (pc *ProductCache) SetProductsWithFilters(ctx context.Context, filters types.ProductFilters, products *dto.ProductCatalogResponse, ttl time.Duration) error {

	jsonString, err := json.Marshal(products)
	if err != nil {
		return err
	}

	key, err := ProductsFilterKey(filters)
	if err != nil {
		return err
	}

	return pc.rdb.Set(ctx, key, string(jsonString), ttl).Err()
}

func (pc *ProductCache) GetProductsWithFilters(ctx context.Context, filters types.ProductFilters) (*dto.ProductCatalogResponse, error) {

	key, err := ProductsFilterKey(filters)
	if err != nil {
		return nil, err
	}

	res := pc.rdb.Get(ctx, key)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	jsonString, err := res.Result()
	if err != nil {
		return nil, err
	}

	products := &dto.ProductCatalogResponse{}
	if err = json.Unmarshal([]byte(jsonString), products); err != nil {
		return nil, err
	}

	return products, nil
}

func (pc *ProductCache) InvalidateProduct(ctx context.Context, productID int64) error {
	key := ProductKey(productID)
	return pc.rdb.Del(ctx, key).Err()
}

func (pc *ProductCache) InvalidateProductLists(ctx context.Context) error {
	pattern := "goshop:products:filter:*"
	keys, err := pc.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		pc.logger.Error("Failed to get product list keys", zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		err = pc.rdb.Del(ctx, keys...).Err()
		if err != nil {
			pc.logger.Error("Failed to delete product list keys", zap.Error(err))
			return err
		}
		pc.logger.Debug("Product lists cache invalidated", zap.Int("keys_deleted", len(keys)))
	}

	return nil
}

func (pc *ProductCache) InvalidateProductsByCategory(ctx context.Context, categoryID int64) error {
	// TODO: оптимизировать - удалять только списки с categoryID
	return pc.InvalidateProductLists(ctx)
}

package infrastructure

import (
	httpadapter "goshop/internal/adapters/input/http"
	cacheadapter "goshop/internal/adapters/output/cache"
	databaseadapter "goshop/internal/adapters/output/database"
	storageadapter "goshop/internal/adapters/output/storage"
	"goshop/internal/config"
	"goshop/internal/core/services"
	"goshop/internal/oauth/google"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func InitApp(
	cfg *config.Config,
	db *pgxpool.Pool,
	logger *zap.Logger,
	rdb *redis.Client,
	minioClient *minio.Client,
	googleOAuth *google.GoogleOAuth,
) *Handlers {
	if logger == nil {
		logger = zap.NewNop()
	}

	repoFactory := databaseadapter.NewFactory(db, logger)
	repositories := repoFactory.WithPool()

	productCache := cacheadapter.NewProductCache(rdb, logger)
	categoryCache := cacheadapter.NewCategoryCache(rdb, logger)
	reviewCache := cacheadapter.NewReviewCache(rdb, logger)

	imageStorage := storageadapter.NewImgStorage(minioClient, cfg.S3.Bucket, cfg.S3.Region)

	categoryService := services.NewCategoryService(repositories.Categories, categoryCache, logger)
	userService := services.NewUserService(repositories.Roles, repositories.Users, cfg.JWT.Secret, cfg.Security.BcryptCost, imageStorage, logger)
	addressService := services.NewAddressService(repositories.Addresses)
	productService := services.NewProductService(repositories.Products, repositories.Categories, imageStorage, productCache, logger)
	cartService := services.NewCartService(repositories.Carts, repositories.Products)
	orderService := services.NewOrderService(repositories.Orders, repositories.Carts, repositories.Users, repositories.Addresses, repositories.OrderItems, logger)
	reviewService := services.NewReviewsService(repositories.Reviews, repositories.Users, repositories.Products, reviewCache, logger)

	userHandler := httpadapter.NewUserHandler(userService, logger)
	addressHandler := httpadapter.NewAddressHandler(addressService)
	categoryHandler := httpadapter.NewCategoryHandler(categoryService)
	productHandler := httpadapter.NewProductHandler(productService, logger)
	cartHandler := httpadapter.NewCartHandler(cartService)
	orderHandler := httpadapter.NewOrderHandler(orderService)
	reviewHandler := httpadapter.NewReviewHandler(reviewService)
	oauthHandler := httpadapter.NewOAuthHandler(googleOAuth, userService, rdb, logger)

	return &Handlers{
		UserHandler:     userHandler,
		AddressHandler:  addressHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		CartHandler:     cartHandler,
		OrderHandler:    orderHandler,
		ReviewHandler:   reviewHandler,
		OAuthHandler:    oauthHandler,
	}
}

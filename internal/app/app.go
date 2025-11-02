package app

import (
	"goshop/internal/cache"
	"goshop/internal/config"
	"goshop/internal/domain/repository"
	address2 "goshop/internal/handler/address"
	"goshop/internal/handler/auth"
	cart2 "goshop/internal/handler/cart"
	category2 "goshop/internal/handler/category"
	order2 "goshop/internal/handler/order"
	product2 "goshop/internal/handler/product"
	review2 "goshop/internal/handler/review"
	user2 "goshop/internal/handler/user"
	"goshop/internal/oauth/google"
	"goshop/internal/repository/pgx"
	imgStorage "goshop/internal/repository/s3"
	"goshop/internal/routes"
	"goshop/internal/service/address"
	"goshop/internal/service/cart"
	"goshop/internal/service/category"
	"goshop/internal/service/order"
	"goshop/internal/service/product"
	"goshop/internal/service/review"
	"goshop/internal/service/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// createRepositories создает набор репозиториев с указанным соединением
func createRepositories(conn repository.DBConn, logger *zap.Logger) pgx.Set {
	return pgx.Set{
		Users:      pgx.NewUserRepository(conn, logger),
		Roles:      pgx.NewRoleRepository(conn),
		Addresses:  pgx.NewAddressRepository(conn),
		Categories: pgx.NewCategoryRepository(conn),
		Products:   pgx.NewProductRepository(conn, logger),
		Carts:      pgx.NewCartRepository(conn, logger),
		Orders:     pgx.NewOrderRepository(conn, logger),
		OrderItems: pgx.NewOrderItemRepository(conn, logger),
		Reviews:    pgx.NewReviewRepository(conn, logger),
	}
}

func InitApp(cfg *config.Config, db *pgxpool.Pool, logger *zap.Logger, rdb *redis.Client, mcl *minio.Client, googleOAuth *google.GoogleOAuth) *routes.Handlers {

	repos := createRepositories(db, logger)

	productCache := cache.NewProductCache(rdb, logger)
	categoryCache := cache.NewCategoryCache(rdb, logger)
	reviewCache := cache.NewReviewCache(rdb, logger)

	storage := imgStorage.NewImgStorage(mcl, cfg.S3.Bucket, cfg.S3.Region)

	categoryService := category.NewCategoryService(repos.Categories, categoryCache, logger)
	userService := user.NewUserService(repos.Roles, repos.Users, cfg.JWT.Secret, cfg.Security.BcryptCost, storage, logger)
	addressService := address.NewAddressService(repos.Addresses)
	productService := product.NewProductService(repos.Products, repos.Categories, storage, productCache, logger)
	cartService := cart.NewCartService(repos.Carts, repos.Products)
	orderService := order.NewOrderService(repos.Orders, repos.Carts, repos.Users, repos.Addresses, repos.OrderItems, logger)
	reviewService := review.NewReviewsService(repos.Reviews, repos.Users, repos.Products, reviewCache, logger)

	userHandler := user2.NewUserHandler(userService, logger)
	addressHandler := address2.NewAddressHandler(addressService)
	categoryHandler := category2.NewCategoryHandler(categoryService)
	productHandler := product2.NewProductHandler(productService, logger)
	cartHandler := cart2.NewCartHandler(cartService)
	orderHandler := order2.NewOrderHandler(orderService)
	reviewHandler := review2.NewReviewHandler(reviewService)
	oauthHandler := auth.NewOAuthHandler(googleOAuth, userService, rdb, logger)

	return &routes.Handlers{
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

package app

import (
	"goshop/internal/cache"
	"goshop/internal/config"
	address2 "goshop/internal/handler/address"
	cart2 "goshop/internal/handler/cart"
	category2 "goshop/internal/handler/category"
	order2 "goshop/internal/handler/order"
	product2 "goshop/internal/handler/product"
	user2 "goshop/internal/handler/user"
	"goshop/internal/repository"
	"goshop/internal/routes"
	"goshop/internal/service/address"
	"goshop/internal/service/cart"
	"goshop/internal/service/category"
	"goshop/internal/service/order"
	"goshop/internal/service/product"
	"goshop/internal/service/user"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitApp(cfg *config.Config, db *pgxpool.Pool, logger *zap.Logger, rdb *redis.Client) *routes.Handlers {

	userRepo := repository.NewUserRepository(db, logger)
	roleRepo := repository.NewRoleRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db, logger)
	cartRepo := repository.NewCartRepository(db, logger)
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)

	productCache := cache.NewProductCache(rdb, logger)

	categoryService := category.NewCategoryService(categoryRepo, logger)
	userService := user.NewUserService(roleRepo, userRepo, cfg.JWT.Secret, cfg.Security.BcryptCost, logger)
	addressService := address.NewAddressService(addressRepo)
	productService := product.NewProductService(productRepo, categoryRepo, productCache, logger)
	cartService := cart.NewCartService(cartRepo, productRepo)
	orderService := order.NewOrderService(orderRepo, cartRepo, userRepo, addressRepo, orderItemRepo, logger)

	userHandler := user2.NewUserHandler(userService)
	addressHandler := address2.NewAddressHandler(addressService)
	categoryHandler := category2.NewCategoryHandler(categoryService)
	productHandler := product2.NewProductHandler(productService, logger)
	cartHandler := cart2.NewCartHandler(cartService)
	orderHandler := order2.NewOrderHandler(orderService)

	return &routes.Handlers{
		UserHandler:     userHandler,
		AddressHandler:  addressHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		CartHandler:     cartHandler,
		OrderHandler:    orderHandler,
	}
}

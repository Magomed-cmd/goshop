package app

import (
	"go.uber.org/zap"
	"goshop/internal/config"
	address2 "goshop/internal/handler/address"
	cart2 "goshop/internal/handler/cart"
	category2 "goshop/internal/handler/category"
	product2 "goshop/internal/handler/product"
	user2 "goshop/internal/handler/user"
	"goshop/internal/repository"
	"goshop/internal/routes"
	"goshop/internal/service/address"
	"goshop/internal/service/cart"
	"goshop/internal/service/category"
	"goshop/internal/service/product"
	"goshop/internal/service/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitApp(cfg *config.Config, db *pgxpool.Pool, logger *zap.Logger) *routes.Handlers {

	userRepo := repository.NewUserRepository(db, logger)
	roleRepo := repository.NewRoleRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db, logger)
	cartRepo := repository.NewCartRepository(db, logger)

	categoryService := category.NewCategoryService(categoryRepo)
	userService := user.NewUserService(roleRepo, userRepo, cfg.JWT.Secret, cfg.Security.BcryptCost, logger)
	addressService := address.NewAddressService(addressRepo)
	productService := product.NewProductService(productRepo, categoryRepo, logger)
	cartService := cart.NewCartService(cartRepo, productRepo)

	userHandler := user2.NewUserHandler(userService)
	addressHandler := address2.NewAddressHandler(addressService)
	categoryHandler := category2.NewCategoryHandler(categoryService)
	productHandler := product2.NewProductHandler(productService, logger)
	cartHandler := cart2.NewCartHandler(cartService)

	return &routes.Handlers{
		UserHandler:     userHandler,
		AddressHandler:  addressHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		CartHandler:     cartHandler,
	}
}

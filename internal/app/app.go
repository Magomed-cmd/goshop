package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/config"
	address2 "goshop/internal/handler/address"
	category2 "goshop/internal/handler/category"
	user2 "goshop/internal/handler/user"
	"goshop/internal/repository"
	"goshop/internal/routes"
	"goshop/internal/service/address"
	"goshop/internal/service/category"
	"goshop/internal/service/user"
)

func InitApp(cfg *config.Config, db *pgxpool.Pool) *routes.Handlers {

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	categoryService := category.NewCategoryService(categoryRepo)
	userService := user.NewUserService(roleRepo, userRepo, cfg.JWT.Secret)
	addressService := address.NewAddressService(addressRepo)

	userHandler := user2.NewUserHandler(userService)
	addressHandler := address2.NewAddressHandler(addressService)
	categoryHandler := category2.NewCategoryHandler(categoryService)
	return &routes.Handlers{
		UserHandler:     userHandler,
		AddressHandler:  addressHandler,
		CategoryHandler: categoryHandler,
	}
}

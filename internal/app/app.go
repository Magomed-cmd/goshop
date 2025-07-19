package app

import (
	"gorm.io/gorm"
	"goshop/internal/config"
	"goshop/internal/handler"
	"goshop/internal/repository"
	"goshop/internal/routes"
	"goshop/internal/service/address"
	"goshop/internal/service/user"
)

func InitApp(cfg *config.Config, db *gorm.DB) *routes.Handlers {

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	addressRepo := repository.NewAddressRepository(db)

	userService := user.NewUserService(roleRepo, userRepo, cfg.JWT.Secret)
	addressService := address.NewAddressService(addressRepo)

	userHandler := handler.NewUserHandler(userService)
	addressHandler := handler.NewAddressHandler(addressService)

	return &routes.Handlers{
		UserHandler:    userHandler,
		AddressHandler: addressHandler,
	}
}

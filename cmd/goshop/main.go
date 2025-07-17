package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
	"goshop/internal/config"
	"goshop/internal/db/postgres"
	"goshop/internal/handler"
	"goshop/internal/repository"
	"goshop/internal/routes"
	"goshop/internal/service/auth"
	"os"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	db, err := postgres.NewConnection(&cfg.Database.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	db.Logger = db.Logger.LogMode(logger.Silent)

	authRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	authService := auth.NewAuthService(roleRepo, authRepo, cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authService)

	routes.RegisterRoutes(r, authHandler)

	log.Info().Str("address", cfg.Server.GetServerAddr()).Msg("Server starting")
	if err := r.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

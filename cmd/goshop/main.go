package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"goshop/internal/app"
	"goshop/internal/config"
	"goshop/internal/db/postgres"
	"goshop/internal/routes"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	db, err := postgres.NewConnection(&cfg.Database.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	handlers := app.InitApp(cfg, db)
	routes.RegisterRoutes(r, handlers, cfg.JWT.Secret)

	log.Info().Str("address", cfg.Server.GetServerAddr()).Msg("Server starting")
	if err := r.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

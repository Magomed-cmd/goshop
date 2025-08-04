package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goshop/internal/app"
	"goshop/internal/config"
	"goshop/internal/db/postgres"
	"goshop/internal/logger"
	"goshop/internal/routes"
)

func main() {

	log := logger.NewFromGinMode("debug")

	cfg, err := config.LoadConfig(".", log)
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	db, err := postgres.NewConnection(&cfg.Database.Postgres, log)
	if err != nil {
		log.Fatal("Failed to connect to Postgres", zap.Error(err))
	}
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	handlers := app.InitApp(cfg, db, log)
	routes.RegisterRoutes(r, handlers, cfg.JWT.Secret, log)

	log.Info("Server starting", zap.String("address", cfg.Server.GetServerAddr()))
	if err := r.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

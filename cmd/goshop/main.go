package main

import (
	"goshop/internal/app"
	"goshop/internal/config"
	"goshop/internal/db/postgres"
	"goshop/internal/db/redisDB"
	"goshop/internal/logger"
	"goshop/internal/routes"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	rdb, err := redisDB.NewConnection(&cfg.Redis, log)
	if err != nil {
		log.Fatal("failed to connect to Redis", zap.Error(err))
	}

	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error("failed to close Redis connection", zap.Error(err))
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	handlers := app.InitApp(cfg, db, log, rdb)
	routes.RegisterRoutes(r, handlers, cfg.JWT.Secret, log)

	log.Info("Server starting", zap.String("address", cfg.Server.GetServerAddr()))
	if err := r.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

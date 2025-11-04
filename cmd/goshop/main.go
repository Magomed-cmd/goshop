package main

// @title        Goshop API
// @version      1.0
// @description  REST API for the Goshop e-commerce platform.
// @BasePath     /
// @schemes      https http
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization

import (
	_ "goshop/docs"
	storageadapter "goshop/internal/adapters/output/storage"
	"goshop/internal/config"
	"goshop/internal/infrastructure"
	"goshop/internal/infrastructure/database/postgres"
	redisdb "goshop/internal/infrastructure/database/redis"
	"goshop/internal/logger"
	"goshop/internal/oauth/google"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewFromGinMode("debug")

	cfg, err := config.LoadConfig(".", log)
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	s3Client, err := storageadapter.NewS3Connection(cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.Secret, cfg.S3.UseSSL, log)
	if err != nil {
		log.Fatal("Failed to connect to S3", zap.Error(err))
	}

	db, err := postgres.NewConnection(&cfg.Database.Postgres, log)
	if err != nil {
		log.Fatal("Failed to connect to Postgres", zap.Error(err))
	}
	defer db.Close()

	rdb, err := redisdb.NewConnection(&cfg.Redis, log)
	if err != nil {
		log.Fatal("failed to connect to Redis", zap.Error(err))
	}

	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error("failed to close Redis connection", zap.Error(err))
		}
	}()

	googleOAuth := google.New(google.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
	})

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	handlers := infrastructure.InitApp(cfg, db, log, rdb, s3Client, googleOAuth)
	infrastructure.RegisterRoutes(r, handlers, cfg.JWT.Secret, log)

	log.Info("Server starting", zap.String("address", cfg.Server.GetServerAddr()))
	if err := r.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

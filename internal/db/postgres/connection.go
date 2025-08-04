package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"goshop/internal/config"
	"strings"
)

func NewConnection(cfg *config.PostgresConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	dsn := cfg.GetDSN()

	logger.Debug("Connecting to PostgreSQL", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("Failed to connect to Postgres", zap.Error(err))
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		dbpool.Close()
		logger.Error("Failed to ping Postgres", zap.Error(err))
		return nil, err
	}

	safeDSN := strings.Replace(dsn, cfg.Password, "**hidden**", 1)
	logger.Info("PostgreSQL connection established", zap.String("dsn", safeDSN))

	return dbpool, nil
}

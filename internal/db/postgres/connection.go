package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"goshop/internal/config"
	"strings"
)

func NewConnection(cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := cfg.GetDSN()

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to Postgres")
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		dbpool.Close()
		log.Error().Err(err).Msg("Failed to ping Postgres")
		return nil, err
	}

	safeDSN := strings.Replace(dsn, cfg.Password, "**hidden**", 1)
	log.Info().Str("dsn", safeDSN).Msg("PostgreSQL connection established")

	return dbpool, nil
}

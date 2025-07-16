package postgres

import (
	log "github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"goshop/internal/config"
	"strings"
)

func NewConnection(cfg *config.PostgresConfig) (*gorm.DB, error) {

	dsn := cfg.GetDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	log.Info().Str("dsn", strings.Replace(dsn, cfg.Password, "**hidden*", 1)).Msg("PostgresSQL connection established")

	return db, nil
}

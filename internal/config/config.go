package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret    string        `mapstructure:"secret"`
	ExpiresIn time.Duration `mapstructure:"expires_in"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOSHOP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	log.Debug().Str("path", path).Msg("Loading configuration")

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to read config file")
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	log.Info().Str("file", viper.ConfigFileUsed()).Msg("Configuration file loaded")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal config")
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Info().
		Str("db_host", cfg.Database.Postgres.Host).
		Int("db_port", cfg.Database.Postgres.Port).
		Str("server_addr", cfg.Server.GetServerAddr()).
		Msg("Configuration loaded successfully")

	return &cfg, nil
}

func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

func (s *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}

	if c.Database.Postgres.Port < 1 || c.Database.Postgres.Port > 65535 {
		return fmt.Errorf("invalid postgres port: %d", c.Database.Postgres.Port)
	}

	if c.Database.Postgres.User == "" {
		return fmt.Errorf("postgres user is required")
	}

	if c.Database.Postgres.Password == "" {
		return fmt.Errorf("postgres password is required")
	}

	if c.Database.Postgres.DBName == "" {
		return fmt.Errorf("postgres database name is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if c.JWT.ExpiresIn <= 0 {
		return fmt.Errorf("JWT expires_in must be positive")
	}

	if c.Database.Postgres.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive")
	}

	if c.Database.Postgres.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive")
	}

	return nil
}

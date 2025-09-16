package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
	Gin      GinConfig      `mapstructure:"gin"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Redis    RedisConfig    `mapstructure:"redis"`
	S3       S3Config       `mapstructure:"s3"`
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

type SecurityConfig struct {
	BcryptCost int `mapstructure:"bcrypt_cost"`
}

type GinConfig struct {
	Mode               string   `mapstructure:"mode"`
	EnableCORS         bool     `mapstructure:"enable_cors"`
	TrustedProxies     []string `mapstructure:"trusted_proxies"`
	MaxMultipartMemory int      `mapstructure:"max_multipart_memory"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Encoding    string `mapstructure:"encoding"`
	Development bool   `mapstructure:"development"`
}

type S3Config struct {
	AccessKey string `mapstructure:"access_key"`
	Secret    string `mapstructure:"secret"`
	Bucket    string `mapstructure:"bucket"`
	Endpoint  string `mapstructure:"endpoint"`
	Region    string `mapstructure:"region"`
	UseSSL    bool   `mapstructure:"useSSL"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	KeyPrefix    string        `mapstructure:"key_prefix"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	TTL          struct {
		ProductsPopular time.Duration `mapstructure:"products_popular"`
		CategoriesAll   time.Duration `mapstructure:"categories_all"`
		Cart            time.Duration `mapstructure:"cart"`
	} `mapstructure:"ttl"`
}

func LoadConfig(path string, logger *zap.Logger) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOSHOP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	logger.Debug("Loading configuration", zap.String("path", path))

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Failed to read config file", zap.Error(err), zap.String("path", path))
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	logger.Info("Configuration file loaded", zap.String("file", viper.ConfigFileUsed()))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error("Failed to unmarshal config", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(logger); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	logger.Info("Configuration loaded successfully",
		zap.String("db_host", cfg.Database.Postgres.Host),
		zap.Int("db_port", cfg.Database.Postgres.Port),
		zap.String("server_addr", cfg.Server.GetServerAddr()),
		zap.Int("bcrypt_cost", cfg.Security.BcryptCost),
		zap.String("gin_mode", cfg.Gin.Mode),
		zap.String("log_level", cfg.Logger.Level),
		zap.String("s3_bucket", cfg.S3.Bucket),
		zap.String("s3_endpoint", cfg.S3.Endpoint))

	return &cfg, nil
}

func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

func (s *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (c *Config) Validate(logger *zap.Logger) error {
	if c.Server.Host == "" {
		logger.Error("Validation failed", zap.String("field", "server.host"), zap.String("error", "required"))
		return fmt.Errorf("server host is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		logger.Error("Validation failed", zap.String("field", "server.port"), zap.Int("value", c.Server.Port), zap.String("error", "invalid range"))
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Postgres.Host == "" {
		logger.Error("Validation failed", zap.String("field", "database.postgres.host"), zap.String("error", "required"))
		return fmt.Errorf("postgres host is required")
	}

	if c.Database.Postgres.Port < 1 || c.Database.Postgres.Port > 65535 {
		logger.Error("Validation failed", zap.String("field", "database.postgres.port"), zap.Int("value", c.Database.Postgres.Port), zap.String("error", "invalid range"))
		return fmt.Errorf("invalid postgres port: %d", c.Database.Postgres.Port)
	}

	if c.Database.Postgres.User == "" {
		logger.Error("Validation failed", zap.String("field", "database.postgres.user"), zap.String("error", "required"))
		return fmt.Errorf("postgres user is required")
	}

	if c.Database.Postgres.Password == "" {
		logger.Error("Validation failed", zap.String("field", "database.postgres.password"), zap.String("error", "required"))
		return fmt.Errorf("postgres password is required")
	}

	if c.Database.Postgres.DBName == "" {
		logger.Error("Validation failed", zap.String("field", "database.postgres.dbname"), zap.String("error", "required"))
		return fmt.Errorf("postgres database name is required")
	}

	if c.JWT.Secret == "" {
		logger.Error("Validation failed", zap.String("field", "jwt.secret"), zap.String("error", "required"))
		return fmt.Errorf("JWT secret is required")
	}

	if len(c.JWT.Secret) < 32 {
		logger.Error("Validation failed", zap.String("field", "jwt.secret"), zap.Int("length", len(c.JWT.Secret)), zap.String("error", "too short"))
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if c.JWT.ExpiresIn <= 0 {
		logger.Error("Validation failed", zap.String("field", "jwt.expires_in"), zap.Duration("value", c.JWT.ExpiresIn), zap.String("error", "must be positive"))
		return fmt.Errorf("JWT expires_in must be positive")
	}

	if c.Database.Postgres.MaxOpenConns <= 0 {
		logger.Error("Validation failed", zap.String("field", "database.postgres.max_open_conns"), zap.Int("value", c.Database.Postgres.MaxOpenConns), zap.String("error", "must be positive"))
		return fmt.Errorf("max_open_conns must be positive")
	}

	if c.Database.Postgres.MaxIdleConns <= 0 {
		logger.Error("Validation failed", zap.String("field", "database.postgres.max_idle_conns"), zap.Int("value", c.Database.Postgres.MaxIdleConns), zap.String("error", "must be positive"))
		return fmt.Errorf("max_idle_conns must be positive")
	}

	if c.Security.BcryptCost < 4 || c.Security.BcryptCost > 15 {
		logger.Error("Validation failed", zap.String("field", "security.bcrypt_cost"), zap.Int("value", c.Security.BcryptCost), zap.String("error", "invalid range"))
		return fmt.Errorf("bcrypt_cost must be between 4 and 15, got: %d", c.Security.BcryptCost)
	}

	if c.Gin.Mode != "debug" && c.Gin.Mode != "release" && c.Gin.Mode != "test" {
		logger.Error("Validation failed", zap.String("field", "gin.mode"), zap.String("value", c.Gin.Mode), zap.String("error", "invalid value"))
		return fmt.Errorf("gin mode must be debug, release or test, got: %s", c.Gin.Mode)
	}

	if c.Gin.MaxMultipartMemory <= 0 {
		logger.Error("Validation failed", zap.String("field", "gin.max_multipart_memory"), zap.Int("value", c.Gin.MaxMultipartMemory), zap.String("error", "must be positive"))
		return fmt.Errorf("max_multipart_memory must be positive")
	}

	if c.Logger.Level != "debug" && c.Logger.Level != "info" && c.Logger.Level != "warn" && c.Logger.Level != "error" {
		logger.Error("Validation failed", zap.String("field", "logger.level"), zap.String("value", c.Logger.Level), zap.String("error", "invalid value"))
		return fmt.Errorf("logger level must be debug, info, warn or error, got: %s", c.Logger.Level)
	}

	if c.Logger.Encoding != "console" && c.Logger.Encoding != "json" {
		logger.Error("Validation failed", zap.String("field", "logger.encoding"), zap.String("value", c.Logger.Encoding), zap.String("error", "invalid value"))
		return fmt.Errorf("logger encoding must be console or json, got: %s", c.Logger.Encoding)
	}

	if c.Redis.Host == "" {
		logger.Error("Validation failed", zap.String("field", "redis.host"), zap.String("error", "required"))
		return fmt.Errorf("redis host is required")
	}

	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		logger.Error("Validation failed", zap.String("field", "redis.port"), zap.Int("value", c.Redis.Port), zap.String("error", "invalid range"))
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}

	if c.S3.AccessKey == "" {
		logger.Error("Validation failed", zap.String("field", "s3.access_key"), zap.String("error", "required"))
		return fmt.Errorf("s3 access_key is required")
	}
	if c.S3.Secret == "" {
		logger.Error("Validation failed", zap.String("field", "s3.secret"), zap.String("error", "required"))
		return fmt.Errorf("s3 secret is required")
	}
	if c.S3.Bucket == "" {
		logger.Error("Validation failed", zap.String("field", "s3.bucket"), zap.String("error", "required"))
		return fmt.Errorf("s3 bucket is required")
	}
	if c.S3.Endpoint == "" {
		logger.Error("Validation failed", zap.String("field", "s3.endpoint"), zap.String("error", "required"))
		return fmt.Errorf("s3 endpoint is required")
	}
	if c.S3.Region == "" {
		logger.Error("Validation failed", zap.String("field", "s3.region"), zap.String("error", "required"))
		return fmt.Errorf("s3 region is required")
	}

	if !c.S3.UseSSL {
		logger.Warn("S3 connection is not using SSL")
	}

	return nil
}

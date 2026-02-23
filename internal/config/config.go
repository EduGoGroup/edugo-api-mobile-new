package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds the entire application configuration.
type Config struct {
	Environment string          `env:"APP_ENV"     envDefault:"development"`
	Server      ServerConfig    `envPrefix:"SERVER_"`
	Database    DatabaseConfig  `envPrefix:"DATABASE_"`
	Messaging   MessagingConfig `envPrefix:"MESSAGING_"`
	Storage     StorageConfig   `envPrefix:"STORAGE_"`
	Logging     LoggingConfig   `envPrefix:"LOGGING_"`
	Auth        AuthConfig      `envPrefix:"AUTH_"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `env:"PORT"          envDefault:"8080"`
	Host         string        `env:"HOST"          envDefault:"0.0.0.0"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  envDefault:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	SwaggerHost  string        `env:"SWAGGER_HOST"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Postgres PostgresConfig `envPrefix:"POSTGRES_"`
	MongoDB  MongoDBConfig  `envPrefix:"MONGODB_"`
}

// PostgresConfig holds PostgreSQL connection parameters.
type PostgresConfig struct {
	Host         string `env:"HOST"           envDefault:"localhost"`
	Port         int    `env:"PORT"           envDefault:"5432"`
	Database     string `env:"DATABASE"       envDefault:"edugo"`
	User         string `env:"USER"           envDefault:"edugo"`
	Password     string `env:"PASSWORD,required"`
	MaxOpenConns int    `env:"MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns int    `env:"MAX_IDLE_CONNS" envDefault:"10"`
	SSLMode      string `env:"SSL_MODE"       envDefault:"disable"`
}

// DSN returns the PostgreSQL connection string.
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// GormDSN returns the PostgreSQL connection string with search_path for GORM.
func (c *PostgresConfig) GormDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=content,assessment,iam,ui_config,public",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// MongoDBConfig holds MongoDB connection parameters.
type MongoDBConfig struct {
	URI      string        `env:"URI"`
	Database string        `env:"DATABASE" envDefault:"edugo"`
	Timeout  time.Duration `env:"TIMEOUT"  envDefault:"10s"`
}

// MessagingConfig holds RabbitMQ configuration.
type MessagingConfig struct {
	RabbitMQ RabbitMQConfig `envPrefix:"RABBITMQ_"`
}

// RabbitMQConfig holds RabbitMQ connection parameters.
type RabbitMQConfig struct {
	URL           string `env:"URL"`
	Exchange      string `env:"EXCHANGE"`
	PrefetchCount int    `env:"PREFETCH_COUNT" envDefault:"10"`
}

// StorageConfig holds object storage configuration.
type StorageConfig struct {
	S3 S3Config `envPrefix:"S3_"`
}

// S3Config holds AWS S3 configuration.
type S3Config struct {
	Region         string        `env:"REGION"         envDefault:"us-east-1"`
	Bucket         string        `env:"BUCKET"`
	AccessKeyID    string        `env:"ACCESS_KEY_ID"`
	SecretAccessKey string       `env:"SECRET_ACCESS_KEY"`
	Endpoint       string        `env:"ENDPOINT"`
	PresignExpiry  time.Duration `env:"PRESIGN_EXPIRY" envDefault:"15m"`
}

// LoggingConfig holds logger configuration.
type LoggingConfig struct {
	Level  string `env:"LEVEL"  envDefault:"info"`
	Format string `env:"FORMAT" envDefault:"json"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWT      JWTConfig      `envPrefix:"JWT_"`
	APIAdmin APIAdminConfig `envPrefix:"API_ADMIN_"`
}

// JWTConfig holds JWT validation parameters.
type JWTConfig struct {
	Secret string `env:"SECRET,required"`
	Issuer string `env:"ISSUER" envDefault:"edugo-central"`
}

// APIAdminConfig holds remote auth validation parameters.
type APIAdminConfig struct {
	BaseURL         string        `env:"BASE_URL"`
	Timeout         time.Duration `env:"TIMEOUT"          envDefault:"5s"`
	CacheTTL        time.Duration `env:"CACHE_TTL"        envDefault:"60s"`
	CacheEnabled    bool          `env:"CACHE_ENABLED"    envDefault:"true"`
	RemoteEnabled   bool          `env:"REMOTE_ENABLED"   envDefault:"false"`
	FallbackEnabled bool          `env:"FALLBACK_ENABLED" envDefault:"false"`
}

// Load parses configuration from environment variables.
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("error parsing config from environment: %w", err)
	}
	return &cfg, nil
}

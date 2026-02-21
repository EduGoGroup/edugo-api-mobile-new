package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the entire application configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Messaging MessagingConfig `mapstructure:"messaging"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Auth      AuthConfig      `mapstructure:"auth"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
}

// PostgresConfig holds PostgreSQL connection parameters.
type PostgresConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Database     string `mapstructure:"database"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	SSLMode      string `mapstructure:"ssl_mode"`
}

// DSN returns the PostgreSQL connection string.
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// MongoDBConfig holds MongoDB connection parameters.
type MongoDBConfig struct {
	URI      string        `mapstructure:"uri"`
	Database string        `mapstructure:"database"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// MessagingConfig holds RabbitMQ configuration.
type MessagingConfig struct {
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// RabbitMQConfig holds RabbitMQ connection parameters.
type RabbitMQConfig struct {
	URL           string `mapstructure:"url"`
	Exchange      string `mapstructure:"exchange"`
	PrefetchCount int    `mapstructure:"prefetch_count"`
}

// StorageConfig holds object storage configuration.
type StorageConfig struct {
	S3 S3Config `mapstructure:"s3"`
}

// S3Config holds AWS S3 configuration.
type S3Config struct {
	Region         string        `mapstructure:"region"`
	Bucket         string        `mapstructure:"bucket"`
	AccessKeyID    string        `mapstructure:"access_key_id"`
	SecretAccessKey string       `mapstructure:"secret_access_key"`
	Endpoint       string        `mapstructure:"endpoint"`
	PresignExpiry  time.Duration `mapstructure:"presign_expiry"`
}

// LoggingConfig holds logger configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWT      JWTConfig      `mapstructure:"jwt"`
	APIAdmin APIAdminConfig `mapstructure:"api_admin"`
}

// JWTConfig holds JWT validation parameters.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

// APIAdminConfig holds remote auth validation parameters.
type APIAdminConfig struct {
	BaseURL         string        `mapstructure:"base_url"`
	Timeout         time.Duration `mapstructure:"timeout"`
	CacheTTL        time.Duration `mapstructure:"cache_ttl"`
	CacheEnabled    bool          `mapstructure:"cache_enabled"`
	RemoteEnabled   bool          `mapstructure:"remote_enabled"`
	FallbackEnabled bool          `mapstructure:"fallback_enabled"`
}

// Load reads the configuration from config files and environment variables.
func Load() (*Config, error) {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Storage.S3.PresignExpiry == 0 {
		cfg.Storage.S3.PresignExpiry = 15 * time.Minute
	}

	return &cfg, nil
}

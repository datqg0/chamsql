package configs

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	EnvironmentDev  = "development"
	EnvironmentProd = "production"
)

type Config struct {
	// Server
	Environment string `mapstructure:"ENVIRONMENT"`
	HTTPPort    int    `mapstructure:"HTTP_PORT"`

	// System Database (PostgreSQL)
	DatabaseURI string `mapstructure:"DATABASE_URI"`

	// Sandbox Databases (Query Execution)
	SandboxPostgresURI  string `mapstructure:"SANDBOX_POSTGRES_URI"`
	SandboxMySQLURI     string `mapstructure:"SANDBOX_MYSQL_URI"`
	SandboxSQLServerURI string `mapstructure:"SANDBOX_SQLSERVER_URI"`

	// Query Execution Limits
	QueryTimeoutSeconds int `mapstructure:"QUERY_TIMEOUT_SECONDS"`
	QueryMaxRows        int `mapstructure:"QUERY_MAX_ROWS"`

	// Redis
	RedisURI      string `mapstructure:"REDIS_URI"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// RabbitMQ
	RabbitMQURI string `mapstructure:"RABBITMQ_URI"`

	// MinIO
	MinioEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucket    string `mapstructure:"MINIO_BUCKET"`
	MinioBaseURL   string `mapstructure:"MINIO_BASE_URL"`
	MinioUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`

	// JWT
	AuthSecret           string        `mapstructure:"AUTH_SECRET"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

var cfg Config

func LoadConfig() *Config {
	viper.AutomaticEnv()

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		viper.SetConfigType("env")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("Error loading config file")
		}
	}

	cfg = Config{
		Environment:          viper.GetString("ENVIRONMENT"),
		HTTPPort:             viper.GetInt("HTTP_PORT"),
		DatabaseURI:          viper.GetString("DATABASE_URI"),
		SandboxPostgresURI:   viper.GetString("SANDBOX_POSTGRES_URI"),
		SandboxMySQLURI:      viper.GetString("SANDBOX_MYSQL_URI"),
		SandboxSQLServerURI:  viper.GetString("SANDBOX_SQLSERVER_URI"),
		QueryTimeoutSeconds:  viper.GetInt("QUERY_TIMEOUT_SECONDS"),
		QueryMaxRows:         viper.GetInt("QUERY_MAX_ROWS"),
		RedisURI:             viper.GetString("REDIS_URI"),
		RedisPassword:        viper.GetString("REDIS_PASSWORD"),
		RedisDB:              viper.GetInt("REDIS_DB"),
		RabbitMQURI:          viper.GetString("RABBITMQ_URI"),
		MinioEndpoint:        viper.GetString("MINIO_ENDPOINT"),
		MinioAccessKey:       viper.GetString("MINIO_ACCESS_KEY"),
		MinioSecretKey:       viper.GetString("MINIO_SECRET_KEY"),
		MinioBucket:          viper.GetString("MINIO_BUCKET"),
		MinioBaseURL:         viper.GetString("MINIO_BASE_URL"),
		MinioUseSSL:          viper.GetBool("MINIO_USE_SSL"),
		AuthSecret:           viper.GetString("AUTH_SECRET"),
		AccessTokenDuration:  viper.GetDuration("ACCESS_TOKEN_DURATION"),
		RefreshTokenDuration: viper.GetDuration("REFRESH_TOKEN_DURATION"),
	}

	// Defaults
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}
	if cfg.Environment == "" {
		cfg.Environment = EnvironmentDev
	}
	if cfg.QueryTimeoutSeconds == 0 {
		cfg.QueryTimeoutSeconds = 3
	}
	if cfg.QueryMaxRows == 0 {
		cfg.QueryMaxRows = 1000
	}
	if cfg.AccessTokenDuration == 0 {
		cfg.AccessTokenDuration = 15 * time.Minute
	}
	if cfg.RefreshTokenDuration == 0 {
		cfg.RefreshTokenDuration = 168 * time.Hour
	}

	return &cfg
}

func GetConfig() *Config {
	return &cfg
}

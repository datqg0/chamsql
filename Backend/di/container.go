package di

import (
	"fmt"

	"go.uber.org/dig"

	"backend/configs"
	"backend/db"
	httpServer "backend/internals/server/http"
	"backend/pkgs/jwt"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	kafka_config "backend/pkgs/messaging/kafka"
	"backend/pkgs/minio"
	"backend/pkgs/permissions"
	"backend/pkgs/rabbitmq"
	"backend/pkgs/redis"
	"backend/pkgs/runner"
)

// Container wraps dig.Container for dependency injection
type Container struct {
	*dig.Container
}

// NewContainer creates a new DI container with all dependencies registered
func NewContainer() (*Container, error) {
	c := dig.New()

	// Register all providers
	providers := []interface{}{
		// Config
		configs.LoadConfig,

		// Infrastructure
		provideDatabase,
		provideRedis,
		provideRabbitMQ,
		provideKafkaRegistry,
		provideKafka,
		provideMinio,
		provideJWTProvider,
		provideRunner,
		providePermissionService,

		// Server
		httpServer.NewServer,
	}

	for _, provider := range providers {
		if err := c.Provide(provider); err != nil {
			return nil, fmt.Errorf("failed to provide dependency: %w", err)
		}
	}

	// Initialize logger
	if err := c.Invoke(func(cfg *configs.Config) {
		logger.Initialize(cfg.Environment)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return &Container{Container: c}, nil
}

func provideDatabase(cfg *configs.Config) (*db.Database, error) {
	database, err := db.NewDatabase(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}
	return database, nil
}

func provideRedis(cfg *configs.Config) redis.IRedis {
	client := redis.NewRedis(redis.Config{
		Address:  cfg.RedisURI,
		Password: cfg.RedisPassword,
		Database: cfg.RedisDB,
	})
	if client == nil {
		logger.Warn("Redis not connected")
	}
	return client
}

func provideRabbitMQ(cfg *configs.Config) *rabbitmq.RabbitMQ {
	if cfg.RabbitMQURI == "" {
		return nil
	}
	rmq, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURI)
	if err != nil {
		logger.Warn("RabbitMQ not connected: %v", err)
		return nil
	}
	return rmq
}

func provideMinio(cfg *configs.Config) *minio.MinioClient {
	if cfg.MinioEndpoint == "" {
		return nil
	}
	client, err := minio.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioBaseURL,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Warn("MinIO not connected: %v", err)
		return nil
	}
	return client
}

func provideJWTProvider(cfg *configs.Config) jwt.JWTProvider {
	return jwt.NewJWTProvider(cfg.AuthSecret)
}

func provideRunner(cfg *configs.Config) runner.Runner {
	r, err := runner.NewRunner(cfg)
	if err != nil {
		logger.Warn("Query Runner failed to initialize: %v", err)
		return nil
	}
	return r
}

func provideKafkaRegistry() *kafka.Registry {
	registry := kafka.NewRegistry()
	kafka_config.RegisterSystemTopics(registry)
	logger.Info("Kafka topic registry initialized: %d topics", registry.Len())
	return registry
}

func provideKafka(cfg *configs.Config) kafka.IKafka {
	if cfg == nil || !cfg.KafkaEnabled {
		return nil
	}

	client, err := kafka.NewKafka(kafka.Config{
		Enabled:  cfg.KafkaEnabled,
		Brokers:  kafka.ParseBrokers(cfg.KafkaBrokers),
		ClientID: cfg.KafkaClientID,
	})
	if err != nil {
		logger.Warn("Kafka not connected: %v", err)
		return nil
	}

	return client
}

func providePermissionService(database *db.Database) permissions.PermissionService {
	return permissions.NewPermissionService(database)
}

// Invoke runs a function with dependencies injected
func (c *Container) Invoke(fn interface{}) error {
	return c.Container.Invoke(fn)
}

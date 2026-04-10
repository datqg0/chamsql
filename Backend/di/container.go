package di

import (
	"fmt"

	"go.uber.org/dig"

	"backend/configs"
	"backend/db"
	aiUsecase "backend/internals/ai/usecase"
	pdfHttp "backend/internals/pdf/controller/http"
	pdfRepository "backend/internals/pdf/repository"
	pdfUsecase "backend/internals/pdf/usecase"
	httpServer "backend/internals/server/http"
	"backend/pkgs/ai"
	"backend/pkgs/jwt"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	kafka_config "backend/pkgs/messaging/kafka"
	"backend/pkgs/pdf"
	"backend/pkgs/permissions"
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
		provideKafkaRegistry,
		provideKafka,
		provideJWTProvider,
		provideRunner,
		providePermissionService,

		// PDF & AI Services (Phase 4)
		providePDFParser,
		providePatternMatcher,
		provideHuggingFaceClient,
		provideAISolutionGenerator,
		provideAITestCaseGenerator,
		provideAITestCaseValidator,
		provideAIOrchestrator,
		providePDFRepository,
		providePDFUploadManager,
		providePDFHandler,

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

// ===== PHASE 4: PDF & AI Services =====

func providePDFParser() *pdf.PDFParser {
	return pdf.NewPDFParser()
}

func providePatternMatcher() *ai.PatternMatcher {
	return ai.NewPatternMatcher()
}

func provideHuggingFaceClient(cfg *configs.Config) *ai.HuggingFaceClient {
	return ai.NewHuggingFaceClient(ai.HuggingFaceConfig{
		APIKey:  cfg.HuggingFaceAPIKey,
		Timeout: 30, // 30 seconds
	})
}

func provideAISolutionGenerator(
	patternMatcher *ai.PatternMatcher,
	huggingfaceClient *ai.HuggingFaceClient,
) aiUsecase.IAISolutionGenerator {
	return aiUsecase.NewAISolutionGenerator(patternMatcher, huggingfaceClient)
}

func provideAITestCaseGenerator(database *db.Database) aiUsecase.IAITestCaseGenerator {
	return aiUsecase.NewAITestCaseGenerator(database)
}

func provideAITestCaseValidator(database *db.Database) aiUsecase.IAITestCaseValidator {
	return aiUsecase.NewAITestCaseValidator(database)
}

func provideAIOrchestrator(
	solutionGenerator aiUsecase.IAISolutionGenerator,
	testCaseGenerator aiUsecase.IAITestCaseGenerator,
	testCaseValidator aiUsecase.IAITestCaseValidator,
	database *db.Database,
) aiUsecase.IAIOrchestrator {
	return aiUsecase.NewAIOrchestrator(solutionGenerator, testCaseGenerator, testCaseValidator, database)
}

func providePDFRepository(database *db.Database) pdfRepository.IPDFRepository {
	return pdfRepository.NewPDFRepository(database)
}

func providePDFUploadManager(
	pdfRepo pdfRepository.IPDFRepository,
	pdfParser *pdf.PDFParser,
	solutionGenerator aiUsecase.IAISolutionGenerator,
	testCaseGenerator aiUsecase.IAITestCaseGenerator,
	testCaseValidator aiUsecase.IAITestCaseValidator,
	aiOrchestrator aiUsecase.IAIOrchestrator,
) pdfUsecase.IUploadManager {
	return pdfUsecase.NewUploadManager(
		pdfRepo,
		pdfParser,
		solutionGenerator,
		testCaseGenerator,
		testCaseValidator,
		aiOrchestrator,
	)
}

func providePDFHandler(uploadManager pdfUsecase.IUploadManager) *pdfHttp.PDFHandler {
	return pdfHttp.NewPDFHandler(uploadManager)
}

// Invoke runs a function with dependencies injected
func (c *Container) Invoke(fn interface{}) error {
	return c.Container.Invoke(fn)
}

package di

import (
	"fmt"
	"time"

	"go.uber.org/dig"

	"backend/configs"
	"backend/db"
	aiHttp "backend/internals/ai/controller/http"
	aiUsecase "backend/internals/ai/usecase"
	examRepository "backend/internals/exam/repository"
	examUsecase "backend/internals/exam/usecase"
	pdfHttp "backend/internals/pdf/controller/http"
	pdfRepository "backend/internals/pdf/repository"
	pdfUsecase "backend/internals/pdf/usecase"
	problemHttp "backend/internals/problem/controller/http"
	problemRepo "backend/internals/problem/repository"
	problemUsecase "backend/internals/problem/usecase"
	httpServer "backend/internals/server/http"
	submissionRepository "backend/internals/submission/repository"
	submissionUsecase "backend/internals/submission/usecase"
	"backend/pkgs/cronjob"
	"backend/pkgs/ai"
	"backend/pkgs/jwt"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	kafka_config "backend/pkgs/messaging/kafka"
	"backend/pkgs/pdf"
	"backend/pkgs/permissions"
	"backend/pkgs/redis"
	"backend/pkgs/runner"
	chatbotHttp "backend/internals/chatbot/controller/http"
	chatbotTools "backend/internals/chatbot/tools"
	chatbotUsecase "backend/internals/chatbot/usecase"
	miniopkg "backend/pkgs/minio"
	"backend/sql/models"
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
		provideMinio,

		// PDF & AI Services (Phase 4)
		providePDFParser,
		providePatternMatcher,
		provideHuggingFaceClient,
		provideOpenAIClient,
		provideLLMClient,
		provideAISolutionGenerator,
		provideAITestCaseGenerator,
		provideAITestCaseValidator,
		provideAIOrchestrator,
		providePDFRepository,
		providePDFUploadManager,
		providePDFHandler,
		provideAIHandler,

		// Submission & Grading Services (Phase 4)
		provideSubmissionRepository,
		provideProblemRepository,
		provideProblemUseCase,
		provideProblemHandler,
		provideGradingService,

		// Exam Timer & Cronjob (Phase 3)
		provideExamRepository,
		provideExamOutboxRepository,
		provideExamTimerUseCase,
		provideOutboxRelayTask,
		provideCronjobScheduler,

		// Chatbot (Phase 4 Upgrade)
		provideToolExecutor,
		provideChatbotUseCase,
		provideChatbotHandler,

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

func provideMinio(cfg *configs.Config) miniopkg.IUploadService {
	client, err := miniopkg.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioBaseURL,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Error("Failed to initialize MinIO: %v", err)
		return nil
	}
	return client
}

// ===== PHASE 4: PDF & AI Services =====

func providePDFParser() *pdf.PDFParser {
	return pdf.NewPDFParser()
}

func providePatternMatcher() *ai.PatternMatcher {
	return ai.NewPatternMatcher()
}

func provideHuggingFaceClient(cfg *configs.Config) *ai.HuggingFaceClient {
	if cfg.HuggingFaceAPIKey == "" {
		return nil
	}
	return ai.NewHuggingFaceClient(ai.HuggingFaceConfig{
		APIKey:  cfg.HuggingFaceAPIKey,
		Timeout: 30 * time.Second, // 30 seconds
	})
}

func provideOpenAIClient(cfg *configs.Config) *ai.OpenAIClient {
	if cfg.OpenAIAPIKey == "" {
		return nil
	}
	return ai.NewOpenAIClient(ai.OpenAIConfig{
		APIKey:  cfg.OpenAIAPIKey,
		Timeout: 30 * time.Second,
	})
}

func provideLLMClient(
	cfg *configs.Config,
	hfClient *ai.HuggingFaceClient,
	oaClient *ai.OpenAIClient,
) ai.LLMClient {
	switch cfg.AIProvider {
	case "openai":
		if oaClient != nil {
			return oaClient
		}
		// Fallback nếu OpenAI key chưa set
		logger.Warn("AIProvider=openai nhưng OPENAI_API_KEY trống, fallback HuggingFace")
		if hfClient != nil {
			return hfClient
		}
		return nil
	case "huggingface":
		if hfClient != nil {
			return hfClient
		}
		logger.Warn("AIProvider=huggingface nhưng HUGGINGFACE_API_KEY trống")
		return nil
	default:
		logger.Warn("AIProvider không hợp lệ: %s, dùng pattern-only mode", cfg.AIProvider)
		return nil
	}
}

func provideAISolutionGenerator(
	cfg *configs.Config,
	patternMatcher *ai.PatternMatcher,
	llmClient ai.LLMClient,
) aiUsecase.IAISolutionGenerator {
	return aiUsecase.NewAISolutionGenerator(patternMatcher, llmClient, cfg.AIProvider)
}

func provideAITestCaseGenerator(
	cfg *configs.Config,
	database *db.Database,
	llmClient ai.LLMClient,
) aiUsecase.IAITestCaseGenerator {
	return aiUsecase.NewAITestCaseGenerator(database, llmClient, cfg.AIProvider)
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

func provideAIHandler(orchestrator aiUsecase.IAIOrchestrator) *aiHttp.AIHandler {
	return aiHttp.NewAIHandler(orchestrator)
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
	probRepo problemRepo.IProblemRepository,
	storage miniopkg.IUploadService,
) pdfUsecase.IUploadManager {
	return pdfUsecase.NewUploadManager(
		pdfRepo,
		pdfParser,
		solutionGenerator,
		testCaseGenerator,
		testCaseValidator,
		aiOrchestrator,
		probRepo,
		storage,
	)
}

func providePDFHandler(uploadManager pdfUsecase.IUploadManager, storage miniopkg.IUploadService) *pdfHttp.PDFHandler {
	return pdfHttp.NewPDFHandler(uploadManager, storage)
}

// ===== Submission & Grading Services =====

func provideSubmissionRepository(database *db.Database) submissionRepository.ISubmissionRepository {
	return submissionRepository.NewSubmissionRepository(database)
}

func provideProblemRepository(database *db.Database) problemRepo.IProblemRepository {
	return problemRepo.NewProblemRepository(database)
}

func provideProblemUseCase(repo problemRepo.IProblemRepository, cache redis.IRedis) problemUsecase.IProblemUseCase {
	return problemUsecase.NewProblemUseCase(repo, cache)
}

func provideProblemHandler(uc problemUsecase.IProblemUseCase, storage miniopkg.IUploadService) *problemHttp.ProblemHandler {
	return problemHttp.NewProblemHandler(uc, storage)
}

func provideGradingService(
	subRepo submissionRepository.ISubmissionRepository,
	probRepo problemRepo.IProblemRepository,
	queryRunner runner.Runner,
) submissionUsecase.IGradingService {
	return submissionUsecase.NewGradingService(subRepo, probRepo, queryRunner)
}

// ===== Exam Timer & Cronjob Services =====

func provideExamRepository(database *db.Database) examRepository.IExamRepository {
	return examRepository.NewExamRepository(database)
}

func provideExamOutboxRepository(database *db.Database) examRepository.IExamOutboxRepository {
	return examRepository.NewExamOutboxRepository(database)
}

func provideExamTimerUseCase(
	examRepo examRepository.IExamRepository,
	outboxRepo examRepository.IExamOutboxRepository,
) examUsecase.IExamTimerUseCase {
	return examUsecase.NewExamTimerUseCase(examRepo, outboxRepo)
}

func provideOutboxRelayTask(database *db.Database, kafkaClient kafka.IKafka) *cronjob.OutboxRelayTask {
	return cronjob.NewOutboxRelayTask(database, kafkaClient)
}

func provideCronjobScheduler(
	examTimerUseCase examUsecase.IExamTimerUseCase,
	outboxRelay *cronjob.OutboxRelayTask,
) *cronjob.Scheduler {
	scheduler := cronjob.NewScheduler()
	// Register exam timer task to run every 1 minute (as requested)
	scheduler.Register(
		examUsecase.NewExamTimerTask(examTimerUseCase),
		1*time.Minute,
	)
	// Register outbox relay task to run every 5 seconds
	scheduler.Register(outboxRelay, 5*time.Second)
	return scheduler
}

// ===== Chatbot Services (Phase 4 Upgrade) =====

func provideToolExecutor(database *db.Database, runner runner.Runner) *chatbotTools.ToolExecutor {
	queries := models.New(database.GetPool())
	return chatbotTools.NewToolExecutor(queries, runner)
}

func provideChatbotUseCase(cfg *configs.Config, executor *chatbotTools.ToolExecutor, redis redis.IRedis) chatbotUsecase.IChatbotUseCase {
	return chatbotUsecase.NewChatbotUseCase(cfg, executor, redis)
}

func provideChatbotHandler(uc chatbotUsecase.IChatbotUseCase) *chatbotHttp.ChatbotHandler {
	return chatbotHttp.NewChatbotHandler(uc)
}

// Invoke runs a function with dependencies injected
func (c *Container) Invoke(fn interface{}) error {
	return c.Container.Invoke(fn)
}

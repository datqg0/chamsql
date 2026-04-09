package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"backend/configs"
	"backend/db"
	"backend/di"
	exam_consumer "backend/internals/exam/infrastructure/messaging/kafka/consumer"
	httpServer "backend/internals/server/http"
	submission_consumer "backend/internals/submission/infrastructure/messaging/kafka/consumer"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	"backend/pkgs/rabbitmq"
)

func main() {
	container, err := di.NewContainer()
	if err != nil {
		logger.Fatal("Failed to create DI container: ", err)
	}

	err = container.Invoke(func(
		server *httpServer.Server,
		cfg *configs.Config,
		database *db.Database,
		rmq *rabbitmq.RabbitMQ,
		kafkaClient kafka.IKafka,
		kafkaRegistry *kafka.Registry,
	) {
		logger.Info("Starting Exam & Submission Backend...")

		ctx, cancel := context.WithCancel(context.Background())

		// Ensure Kafka topics exist if Kafka is enabled
		if kafkaClient != nil {
			topics := kafkaRegistry.All()
			if err := kafkaClient.EnsureTopics(ctx, topics); err != nil {
				logger.Error("Failed to ensure Kafka topics: %v", err)
			} else {
				logger.Info("Kafka topics ensured successfully (count=%d)", len(topics))
			}
		}

		// Start Kafka consumers for exam and submission domains
		if kafkaClient != nil {
			examEventConsumer := exam_consumer.NewExamEventConsumer(kafkaClient, database)
			go examEventConsumer.Start(ctx)

			submissionEventConsumer := submission_consumer.NewSubmissionEventConsumer(kafkaClient, database)
			go submissionEventConsumer.Start(ctx)
		}

		// Graceful shutdown
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			logger.Info("Shutting down...")
			cancel() // Stop background workers
			if rmq != nil {
				rmq.Close()
			}
			if kafkaClient != nil {
				if err := kafkaClient.Close(); err != nil {
					logger.Warn("Failed to close Kafka client: %v", err)
				}
			}
			database.Close()
			os.Exit(0)
		}()

		// Run server
		if err := server.Run(); err != nil {
			logger.Fatal("Server error: ", err)
		}
	})

	if err != nil {
		logger.Fatal("Startup failed: ", err)
	}
}

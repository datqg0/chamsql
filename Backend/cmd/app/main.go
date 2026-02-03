package main

import (
	"os"
	"os/signal"
	"syscall"

	"backend/db"
	"backend/di"
	httpServer "backend/internals/server/http"
	"backend/pkgs/logger"
)

func main() {
	// Create DI container
	container, err := di.NewContainer()
	if err != nil {
		logger.Fatal("Failed to create DI container: ", err)
	}

	// Run with injected Server
	err = container.Invoke(func(server *httpServer.Server, database *db.Database) {
		logger.Info("Starting Bus Ticketing Backend...")

		// Graceful shutdown
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			logger.Info("Shutting down...")
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

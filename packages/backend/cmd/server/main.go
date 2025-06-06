package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/konflux-ci/kite/internal/config"
	handler_http "github.com/konflux-ci/kite/internal/handlers/http"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variable
	projectEnv := config.GetEnvOrDefault("KITE_PROJECT_ENV", "development")
	fileName := fmt.Sprintf(".env.%s", projectEnv)
	envFile, err := config.GetEnvFileInCwd(fileName)
	if err != nil {
		log.Printf("failed to get env file %s: %v", fileName, err)
	}
	if err := godotenv.Load(envFile); err != nil {
		// It should be fine if the file doesn't exist
		log.Printf("no %s file found, using system environment variables\n", envFile)
	} else {
		log.Printf("successfully loaded env file %s\n", envFile)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	// Initialize logger
	logger := setupLogger()

	logger.WithFields(logrus.Fields{
		"environment": cfg.Server.Environment,
		"version":     getVersion(),
	})

	// Initialize database
	db, err := config.InitDatabase()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}

	// Get database instance for cleanup
	sqlDB, err := db.DB()
	if err != nil {
		logger.WithError(err).Fatal("Failed to get database instance")
	}
	defer func() {
		err := sqlDB.Close()
		if err != nil {
			logger.WithError(err).Fatal("Failed to close database connection")
		}
	}()

	// Setup router
	router, err := handler_http.SetupRouter(db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to setup router")
	}

	// Setup HTTP server with configuration
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Lets start the server in a goroutine.
	// This lets us run the server in this anonymous function concurrently
	// while allowing main() to continue instead of blocking on ListenAndServe().
	go func() {
		logger.WithFields(logrus.Fields{
			"address":     cfg.GetServerAddress(),
			"environment": cfg.Server.Environment,
		}).Info("Starting HTTP Server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	// Create a channel that carries os.Signal values, buffer size 1
	quit := make(chan os.Signal, 1)
	// Notify 'quit' channel whenver the process receives SIGINT (Ctrl+C) or SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Block here (don't run anything after the next line) until one of those signals is received
	// Because the buffer size is one, once the signal is recieved we'll process the rest of the function.
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shut down server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown gracefully")
	}
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()

	// Set log level
	logLevel := config.GetEnvOrDefault("KITE_LOG_LEVEL", "info")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if config.GetEnvOrDefault("KITE_PROJECT_ENV", "development") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	return logger
}

func getVersion() string {
	// This should be set during build time
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}
	return "dev"
}

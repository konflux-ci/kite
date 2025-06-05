package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/konflux-ci/kite/internal/config"
	"github.com/konflux-ci/kite/internal/seed"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Check which environment we're in
	env := os.Getenv("KITE_PROJECT_ENV")
	if env == "production" {
		logger.Fatal("Seeder can only be used in the development environment")
	}

	// Try to load ENV file
	envFile, err := config.GetEnvFileInCwd(".env.development")
	if err := godotenv.Load(envFile); err != nil {
		logger.WithError(err).Info("Could not load env file, using existing environment variables")
	} else {
		logger.Info("Loaded environment from .env.development")
	}

	logger.WithField("environment", env).Info("Starting database seeding")

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
	defer sqlDB.Close()

	// Run seeding
	if err := seed.SeedData(db); err != nil {
		logger.WithError(err).Fatal("Failed to seed database")
	}

	logger.Info("Database seeding completed successfully")
}

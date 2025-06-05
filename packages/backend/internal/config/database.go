package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Returns the database configuration using ENV variables. Uses defaults if ENV variables are not found.
func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("KITE_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("KITE_DB_PORT", "5432"),
		User:     getEnvOrDefault("KITE_DB_USER", "postgres"),
		Password: getEnvOrDefault("KITE_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("KITE_DB_NAME", "issuesdb"),
		SSLMode:  getEnvOrDefault("KITE_DB_SSL_MODE", "disable"),
	}
}

// Initializes the database.
func InitDatabase() (*gorm.DB, error) {
	config := GetDatabaseConfig()

	// Build the connection string
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.User, config.Password, config.Name, config.Port, config.SSLMode)

	// Configure logger based on environment
	var gormLogger logger.Interface
	if os.Getenv("KITE_PROJECT_ENV") == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Connect to the database, setup the logger
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")
	return db, nil
}

// Gets an ENV variable, returns a defaultValue if not found.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

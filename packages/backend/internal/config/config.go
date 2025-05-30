package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	Security SecurityConfig
	Features FeatureFlags
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Environment     string
}

// LoggingConfig holds all logging configuration
type LoggingConfig struct {
	Level  string
	Format string //json or text
}

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	EnableCORS     bool
	AllowedOrigins []string
	RateLimitRPS   int
}

// FeatureFlags holds feature flag configuration
type FeatureFlags struct {
	EnableNamespaceChecking bool
	EnableWebhooks          bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:            GetEnvOrDefault("HOST", "0.0.0.0"),
			Port:            getEnvOrDefault("PORT", "8080"),
			ReadTimeout:     GetEnvDurationOrDefault("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    GetEnvDurationOrDefault("WRITE_TIMEOUT", 39*time.Second),
			IdleTimeout:     GetEnvDurationOrDefault("IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: GetEnvDurationOrDefault("SHUTDOWN_TIMEOUT", 10*time.Second),
			Environment:     getEnvOrDefault("PROJECT_ENV", "production"),
		},
		Database: DatabaseConfig{
			Host:     GetEnvOrDefault("DB_HOST", "localhost"),
			Port:     GetEnvOrDefault("DB_PORT", "5432"),
			User:     GetEnvOrDefault("DB_USER", "kite"),
			Password: GetEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:     GetEnvOrDefault("DB_NAME", "issuesdb"),
			SSLMode:  GetEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		Logging: LoggingConfig{
			Level:  GetEnvOrDefault("LOG_LEVEL", "info"),
			Format: GetEnvOrDefault("LOG_FORMAT", "json"),
		},
		Security: SecurityConfig{
			EnableCORS:     GetEnvBoolOrDefault("ENABLE_CORS", true),
			AllowedOrigins: GetEnvSliceOrDefault("ALLOWED_ORIGINS", []string{"*"}),
			RateLimitRPS:   GetEnvIntOrDefault("RATE_LIMIT_RPS", 100),
		},
		Features: FeatureFlags{
			EnableNamespaceChecking: GetEnvBoolOrDefault("FEATURE_NAMESPACE_CHECKING", true),
			EnableWebhooks:          GetEnvBoolOrDefault("FEATURE_WEBHOOKS", true),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil

}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	portNum, err := strconv.Atoi(c.Server.Port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf("invalid server port: %s", c.Server.Port)
	}

	// Validate project environment
	validEnvs := []string{"development", "staging", "production", "test"}
	if !slices.Contains(validEnvs, c.Server.Environment) {
		return fmt.Errorf("invalid project environment: %s (must be one of: %s)",
			c.Server.Environment, strings.Join(validEnvs, ", "))
	}

	// Validate databse configuration (TODO)
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is requried")
	}

	// Validate logging configuration
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !slices.Contains(validLogLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			c.Logging.Level, strings.Join(validLogLevels, ", "))
	}

	validLogFormats := []string{"json", "text"}
	if !slices.Contains(validLogFormats, c.Logging.Format) {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			c.Logging.Format, strings.Join(validLogFormats, ", "))
	}

	return nil
}

// Helper functions

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper function to get an environment variable. Defaults to the value passed
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get an environment variable.
//
// If the value is found, it's converted into an int.
//
// Defaults to the value passed.
func GetEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get an environment variable.
//
//	If the value is found, its converted into a boolean.
//
// Defaults to the value passed.
func GetEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Helper function to get an environment variable.
//
// If the value is found, it's converted into a type of time.Duration.
//
// Defaults to the value passed.
func GetEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if timeValue, err := time.ParseDuration(value); err != nil {
			return timeValue
		}
	}
	return defaultValue
}

// Helper function to get an environment variable
//
// # If the value is found, it's converted into a slice of strings
//
// Defaults to the value passed.
func GetEnvSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// GetEnvFileInCwd returns the full path to the given filename in project root directory
func GetEnvFileInCwd(filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(cwd, filename), nil
}

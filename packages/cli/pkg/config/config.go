package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Global configuration
type Config struct {
	APIUrl string `mapstructure:"api_url"`
}

// Default configuration values
const (
	DefaultAPIURL = "http://localhost:8080/api/v1"
)

// Initializes the configuration
func InitConfig() error {
	// Find home directory
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(home, ".konflux-issues")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Set configuration file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set default values
	viper.SetDefault("api_url", DefaultAPIURL)

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// Create default config file if it doesn't exist
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("failed to write default config: %w", err)
		}
	}

	// Support environment variables
	viper.SetEnvPrefix("KONFLUX")
	viper.AutomaticEnv()

	return nil
}

// GetConfig returns the current configuration
func GetConfig() Config {
	return Config{
		APIUrl: viper.GetString("api_url"),
	}
}

// SetAPIURL updates the API URL in the configuration
func SetAPIURL(url string) error {
	viper.Set("api_url", url)
	return viper.WriteConfig()
}

// ResetConfig resets the configuration to default values
func ResetConfig() error {
	viper.Set("api_url", DefaultAPIURL)
	return viper.WriteConfig()
}

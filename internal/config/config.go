package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	Host               string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Storage            string
	FilePath           string // used if Storage == file
	LoggingEnabled     bool
}

// LoadConfig loads configuration from ./config.yaml or defaults
func LoadConfig() *Config {
	configFile := "config.yaml"

	// Force viper to read the exact file
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // allow env overrides

	// Default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "127.0.0.1")
	viper.SetDefault("JWT_SECRET", "super-secret-key")
	viper.SetDefault("ACCESS_TOKEN_EXPIRY", "5m")
	viper.SetDefault("REFRESH_TOKEN_EXPIRY", "168h")
	viper.SetDefault("STORAGE", "memory")
	viper.SetDefault("FILE_PATH", filepath.Join("ezauth-data", "users.json"))
	viper.SetDefault("LOGGING_ENABLED", true)

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, using defaults and environment variables")
	} else {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	accessDur, err := time.ParseDuration(viper.GetString("ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_EXPIRY: %v", err)
	}

	refreshDur, err := time.ParseDuration(viper.GetString("REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_EXPIRY: %v", err)
	}

	return &Config{
		Port:               viper.GetString("PORT"),
		Host:               viper.GetString("HOST"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		AccessTokenExpiry:  accessDur,
		RefreshTokenExpiry: refreshDur,
		Storage:            viper.GetString("STORAGE"),
		FilePath:           viper.GetString("FILE_PATH"),
		LoggingEnabled:     viper.GetBool("LOGGING_ENABLED"),
	}
}

// InitConfig bootstraps config.yaml and ezauth-data/users.json in current directory
func InitConfig() error {
	configPath := "config.yaml"
	dataDir := "ezauth-data"
	usersPath := filepath.Join(dataDir, "users.json")

	// Don't overwrite existing config
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config.yaml already exists in current directory")
	}

	// Create data directory
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create ezauth-data directory: %v", err)
	}

	// Set default values
	viper.Set("PORT", "8080")
	viper.Set("HOST", "127.0.0.1")
	viper.Set("JWT_SECRET", "super-secret-key")
	viper.Set("ACCESS_TOKEN_EXPIRY", "5m")
	viper.Set("REFRESH_TOKEN_EXPIRY", "168h")
	viper.Set("STORAGE", "file")
	viper.Set("FILE_PATH", usersPath)
	viper.Set("LOGGING_ENABLED", true)

	// Write config.yaml to root
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config.yaml: %v", err)
	}

	// Create empty users.json
	if _, err := os.Stat(usersPath); os.IsNotExist(err) {
		if err := os.WriteFile(usersPath, []byte("{}"), 0644); err != nil {
			return fmt.Errorf("failed to create users.json: %v", err)
		}
	}

	log.Println("✅ ezauth initialized successfully")
	return nil
}

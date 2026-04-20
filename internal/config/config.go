package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate JWT secret: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

type Config struct {
	Port               string
	Host               string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Storage            string
	FilePath           string // used if Storage == file
	DatabasePath       string // used if Storage == sqlite
	DatabaseURL        string // used if Storage == postgres
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
	viper.SetDefault("DATABASE_PATH", filepath.Join("ezauth-data", "ezauth.db"))
	viper.SetDefault("DATABASE_URL", "for postgres")
	viper.SetDefault("LOGGING_ENABLED", true)

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		viper.Set("JWT_SECRET", randomSecret())
		log.Println("⚠  No config.yaml found — running in ephemeral mode")
		log.Println("   Storage: memory (users lost on restart)")
		log.Println("   JWT secret: randomly generated (sessions lost on restart)")
		log.Println("   Run `ezauth init` to set up persistent storage")
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
		DatabasePath:       viper.GetString("DATABASE_PATH"),
		DatabaseURL:        viper.GetString("DATABASE_URL"),
		LoggingEnabled:     viper.GetBool("LOGGING_ENABLED"),
	}
}

// InitConfig bootstraps config.yaml and ezauth-data/users.json in current directory
func InitConfig() error {
	configPath := "config.yaml"
	dataDir := "ezauth-data"
	usersPath := filepath.Join(dataDir, "users.json")

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config.yaml already exists in current directory")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create ezauth-data directory: %v", err)
	}

	content := fmt.Sprintf(`host: 127.0.0.1
port: "8080"
jwt_secret: %s
access_token_expiry: 5m
refresh_token_expiry: 168h
storage: file
file_path: %s
database_path: %s
database_url: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
logging_enabled: true
`, randomSecret(), usersPath, filepath.Join(dataDir, "ezauth.db"))

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %v", err)
	}

	if _, err := os.Stat(usersPath); os.IsNotExist(err) {
		if err := os.WriteFile(usersPath, []byte("{}"), 0644); err != nil {
			return fmt.Errorf("failed to create users.json: %v", err)
		}
	}

	log.Println("✅ ezauth initialized successfully")
	return nil
}

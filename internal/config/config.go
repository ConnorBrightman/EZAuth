package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port         string
	JWTSecret    []byte
	TokenExpiry  time.Duration
	StorageType  string // "memory" or "file"
	UserFilePath string // used if StorageType == "file"
}

// LoadConfig reads environment variables or .env file
func LoadConfig() *Config {
	// Load .env if exists
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key")
	storageType := getEnv("STORAGE_TYPE", "memory")
	userFilePath := getEnv("USER_FILE_PATH", "./data/users.json")

	tokenExpiryMinutes, _ := strconv.Atoi(getEnv("TOKEN_EXPIRY_MINUTES", "60"))
	tokenExpiry := time.Duration(tokenExpiryMinutes) * time.Minute

	return &Config{
		Port:         port,
		JWTSecret:    []byte(jwtSecret),
		TokenExpiry:  tokenExpiry,
		StorageType:  storageType,
		UserFilePath: userFilePath,
	}
}

// getEnv reads env or fallback
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

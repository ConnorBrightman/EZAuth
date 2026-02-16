package config

import (
	"log"
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

func LoadConfig() *Config {
	viper.SetConfigName("config") // config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv() // allow overriding via env vars

	// Default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "127.0.0.1")

	viper.SetDefault("JWT_SECRET", "super-secret-key")

	viper.SetDefault("ACCESS_TOKEN_EXPIRY", "5m")
	viper.SetDefault("REFRESH_TOKEN_EXPIRY", "168h")

	viper.SetDefault("STORAGE", "memory")
	viper.SetDefault("FILE_PATH", "./data/users.json")

	viper.SetDefault("LOGGING_ENABLED", true)

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, using defaults and environment variables")
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
		Port: viper.GetString("PORT"),
		Host: viper.GetString("HOST"),

		JWTSecret:          viper.GetString("JWT_SECRET"),
		AccessTokenExpiry:  accessDur,
		RefreshTokenExpiry: refreshDur,

		Storage:  viper.GetString("STORAGE"),
		FilePath: viper.GetString("FILE_PATH"),

		LoggingEnabled: viper.GetBool("LOGGING_ENABLED"),
	}
}

package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	Host        string
	JWTSecret   string
	TokenExpiry time.Duration
	Storage     string
	FilePath    string // used if Storage == file
}

func LoadConfig() *Config {
	viper.SetConfigName("config") // config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv() // allow overriding via env vars

	// Default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("Host", "127.0.0.1")
	viper.SetDefault("JWT_SECRET", "super-secret-key")
	viper.SetDefault("TOKEN_EXPIRY", 24*time.Hour)
	viper.SetDefault("STORAGE", "memory")
	viper.SetDefault("FILE_PATH", "./data/users.json")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, using defaults and environment variables")
	}

	return &Config{
		Port:        viper.GetString("PORT"),
		Host:        viper.GetString("HOST"),
		JWTSecret:   viper.GetString("JWT_SECRET"),
		TokenExpiry: viper.GetDuration("TOKEN_EXPIRY"),
		Storage:     viper.GetString("STORAGE"),
		FilePath:    viper.GetString("FILE_PATH"),
	}
}

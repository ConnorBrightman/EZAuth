package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ConnorBrightman/ezauth/internal/api"
	"github.com/ConnorBrightman/ezauth/internal/auth"
	"github.com/ConnorBrightman/ezauth/internal/config"
	"github.com/ConnorBrightman/ezauth/internal/fileserver"
	"github.com/ConnorBrightman/ezauth/internal/middleware"
)

func runStart() {
	// Load configuration
	cfg := config.LoadConfig()

	fmt.Printf("🚀 Starting ezauth on %s:%s\n", cfg.Host, cfg.Port)

	// Ensure data directory exists if using file storage
	if cfg.Storage == "file" {
		dataDir := filepath.Dir(cfg.FilePath)
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				log.Fatalf("Failed to create data directory: %v", err)
			}
		}
	}

	// Initialize user repository
	var repo auth.UserRepository
	var err error
	switch cfg.Storage {
	case "file":
		repo, err = auth.NewFileUserRepository(cfg.FilePath)
		if err != nil {
			log.Fatal(err)
		}
	case "sqlite":
		repo, err = auth.NewSQLiteUserRepository(cfg.DatabasePath)
		if err != nil {
			log.Fatal(err)
		}
	case "postgres":
		if cfg.DatabaseURL == "" {
			log.Fatal("storage: postgres requires database_url to be set in config.yaml")
		}
		repo, err = auth.NewPostgresUserRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
	case "mysql":
		if cfg.DatabaseURL == "" {
			log.Fatal("storage: mysql requires database_url to be set in .env or config.yaml")
		}
		repo, err = auth.NewMySQLUserRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
	case "memory":
		repo = auth.NewMemoryUserRepository()
	default:
		log.Fatal("unsupported storage backend: ", cfg.Storage)
	}

	// Create auth service
	service := auth.NewService(repo, []byte(cfg.JWTSecret), cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)

	// Create router with JWT secret
	router := api.NewRouter(service, []byte(cfg.JWTSecret))

	// Serve static files from ./public if it exists
	mainHandler := http.NewServeMux()
	if _, err := os.Stat("public"); err == nil {
		mainHandler.Handle("/", fileserver.ServePublic())
		mainHandler.HandleFunc("/register", fileserver.ServePage("/register", "register.html"))
		mainHandler.HandleFunc("/login", fileserver.ServePage("/login", "login.html"))
	}

	// Mount API router
	mainHandler.Handle("/auth/", router)

	handler := middleware.Logging(mainHandler)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	fmt.Printf(`
 _____ _____         _   _
|   __|__   |___ _ _| |_| |_
|   __|   __| .'| | |  _|   |
|_____|_____|__,|___|_| |_|_|
-- Authentication made EZ. --

`)
	fmt.Printf("storage: %s   port: %s   access token expiry: %s\n\n", cfg.Storage, cfg.Port, cfg.AccessTokenExpiry)
	fmt.Printf("API:  http://%s/auth\n", addr)
	if _, err := os.Stat("public"); err == nil {
		fmt.Printf("Demo: http://%s\n", addr)
	}
	fmt.Println()

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

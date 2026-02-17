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
	case "memory":
		repo = auth.NewMemoryUserRepository()
	default:
		log.Fatal("unsupported storage backend: ", cfg.Storage)
	}

	// Create auth service
	service := auth.NewService(repo, []byte(cfg.JWTSecret), cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)

	// Create router with JWT secret
	router := api.NewRouter(service, []byte(cfg.JWTSecret))

	// Wrap router with logging middleware
	handler := middleware.Logging(router)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Banner
	log.Println(`
 _____ _____         _   _   
|   __|__   |___ _ _| |_| |_ 
|   __|   __| .'| | |  _|   |
|_____|_____|__,|___|_| |_|_|              
Authentication made EZ          
`)

	log.Printf("Starting EZauth with storage=%s, port=%s, AccessTokenExpiry=%s, RefreshTokenExpiry=%s\n",
		cfg.Storage, cfg.Port, cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)
	log.Fatal(server.ListenAndServe())
}

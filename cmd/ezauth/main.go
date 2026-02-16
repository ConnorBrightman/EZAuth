package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ezauth/internal/api"
	"ezauth/internal/auth"
	"ezauth/internal/config"
	"ezauth/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	// Ensure data directory exists if using file storage
	if cfg.Storage == "file" {
		if _, err := os.Stat("./data"); os.IsNotExist(err) {
			os.Mkdir("./data", 0755)
		}
	}

	// Choose repository
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
		log.Fatal("unsupported storage backend")
	}

	// Create auth service
	service := auth.NewService(repo, []byte(cfg.JWTSecret), cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)

	// Create router with JWT secret
	router := api.NewRouter(service, []byte(cfg.JWTSecret))

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	// Wrap router with logging middleware
	handler := middleware.Logging(router)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Println(`                       
 _____ _____         _   _   
|   __|__   |___ _ _| |_| |_ 
|   __|   __| .'| | |  _|   |
|_____|_____|__,|___|_| |_|_|              
                                                                 
`)
	log.Printf("Starting EZauth with storage=%s, port=%s, AccesstokenExpiry=%s\n", cfg.Storage, cfg.Port, cfg.AccessTokenExpiry)
	log.Fatal(server.ListenAndServe())
}

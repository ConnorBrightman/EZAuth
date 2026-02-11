package main

import (
	"log"
	"net/http"
	"os"

	"ezauth/internal/api"
	"ezauth/internal/auth"
	"ezauth/internal/config"
	"ezauth/internal/middleware"
)

func main() {
	cfg := config.LoadConfig()

	// Ensure data directory exists
	if cfg.StorageType == "file" {
		if _, err := os.Stat("./data"); os.IsNotExist(err) {
			os.Mkdir("./data", 0755)
		}
	}

	// Choose repository
	var repo auth.UserRepository
	var err error
	if cfg.StorageType == "file" {
		repo, err = auth.NewFileUserRepository(cfg.UserFilePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		repo = auth.NewMemoryUserRepository()
	}

	service := auth.NewService(repo, cfg.JWTSecret, cfg.TokenExpiry)

	// Router
	router := api.NewRouter(service, cfg.JWTSecret)

	handler := middleware.Logging(router)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	log.Println("EZauth server running on http://localhost:" + cfg.Port)
	log.Fatal(server.ListenAndServe())
}

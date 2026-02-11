package main

import (
	"log"
	"net/http"
	"os"

	"ezauth/internal/api"
	"ezauth/internal/auth"
	"ezauth/internal/middleware"
)

func main() {
	// Ensure data directory exists
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.Mkdir("./data", 0755)
	}

	// Choose repository: Memory or File
	// repo := auth.NewMemoryUserRepository()
	repo, err := auth.NewFileUserRepository("./data/users.json")
	if err != nil {
		log.Fatal(err)
	}

	// Create auth service
	secretKey := []byte("super-secret-key") // In production, load from env
	service := auth.NewService(repo, secretKey)
	// Create router with service
	router := api.NewRouter(service, secretKey)

	// Wrap router with logging middleware
	handler := middleware.Logging(router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println(`
                                  
 _____ _____    _____     _   _   
|   __|__   |  |  _  |_ _| |_| |_ 
|   __|   __|  |     | | |  _|   |
|_____|_____|  |__|__|___|_| |_|_|
                                                                 
	`)
	log.Println("EZauth server running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

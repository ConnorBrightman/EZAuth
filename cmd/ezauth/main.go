package main

import (
	"log"
	"net/http"

	"ezauth/internal/api"
)

func main() {
	router := api.NewRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
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

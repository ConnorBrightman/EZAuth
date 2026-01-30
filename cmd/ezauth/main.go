package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting EZauth...")

	server := &http.Server{
		Addr: ":8080",
	}

	log.Fatal(server.ListenAndServe())
}

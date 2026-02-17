package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ConnorBrightman/ezauth/internal/config"
	"github.com/ConnorBrightman/ezauth/internal/templates"
)

func runInit() {
	// Initialize config and data
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}

	// Create public folder for static files
	publicDir := "public"
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			log.Fatalf("Failed to create public directory: %v", err)
		}
		log.Println("✅ Created public/ directory for static files")
	} else {
		log.Println("📁 public/ directory already exists")
	}

	// Scaffold starter HTML templates
	if err := templates.GenerateTemplates(); err != nil {
		log.Printf("❌ Failed to generate templates: %v", err)
	} else {
		log.Println("✅ Starter HTML templates generated in public/")
	}

	fmt.Println("✅ ezauth initialized successfully.")
	fmt.Println("Next step: run `ezauth start`")
}

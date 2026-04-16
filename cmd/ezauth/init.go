package main

import (
	"fmt"
	"os"

	"github.com/ConnorBrightman/ezauth/internal/config"
	"github.com/ConnorBrightman/ezauth/internal/templates"
)

func runInit() {
	if err := config.InitConfig(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	publicDir := "public"
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			fmt.Println("Error: failed to create public directory:", err)
			os.Exit(1)
		}
	}

	if err := templates.GenerateTemplates(); err != nil {
		fmt.Println("Error: failed to generate templates:", err)
		os.Exit(1)
	}

	fmt.Println("✅ config.yaml created")
	fmt.Println("✅ Demo pages written to public/")
	fmt.Println("")
	fmt.Println("Run `ezauth start` to start the server")
}

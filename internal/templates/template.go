package templates

import (
	"log"
	"os"
	"path/filepath"
)

// CreateStarterTemplates writes starter HTML files into ./public
func CreateStarterTemplates() {
	publicDir := "public"
	files := map[string]string{
		"index.html":    `<h1>Welcome to EZauth</h1>`,
		"register.html": `<h1>Register page</h1>`,
		"login.html":    `<h1>Login page</h1>`,
	}

	for name, content := range files {
		path := filepath.Join(publicDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				log.Printf("Failed to create %s: %v", name, err)
			}
		}
	}

	log.Println("✅ Starter templates written to ./public")
}

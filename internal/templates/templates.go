package templates

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

//go:embed HTML/*
var templateFiles embed.FS

func GenerateTemplates() error {
	publicDir := "public"

	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			return err
		}
	}

	// Walk embedded files
	return fs.WalkDir(templateFiles, "HTML", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("❌ Skipping %s due to error: %v", path, err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		data, err := templateFiles.ReadFile(path)
		if err != nil {
			log.Printf("❌ Failed to read embedded template %s: %v", path, err)
			return nil
		}

		dest := filepath.Join(publicDir, filepath.Base(path))
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			if err := os.WriteFile(dest, data, 0644); err != nil {
				log.Printf("❌ Failed to write %s: %v", dest, err)
			} else {
				log.Printf("✅ Created %s", dest)
			}
		}

		return nil
	})
}

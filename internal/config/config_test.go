package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig_FilePathPopulated(t *testing.T) {
	// Write a minimal config.yaml to a temp dir and cd into it
	dir := t.TempDir()
	content := `
storage: file
file_path: /some/absolute/path/users.json
port: "8080"
host: 127.0.0.1
access_token_expiry: 5m
refresh_token_expiry: 168h
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	viper.Reset()
	cfg := LoadConfig()

	if cfg.FilePath == "" {
		t.Fatal("FilePath is empty — viper failed to unmarshal file_path from config.yaml")
	}
	if cfg.Storage != "file" {
		t.Fatalf("expected storage=file, got %q", cfg.Storage)
	}
	t.Logf("FilePath = %q", cfg.FilePath)
}

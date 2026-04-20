package config

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate JWT secret: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

type Config struct {
	Port               string        `mapstructure:"port"`
	Host               string        `mapstructure:"host"`
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	Storage            string        `mapstructure:"storage"`
	FilePath           string        `mapstructure:"file_path"`
	DatabasePath       string        `mapstructure:"database_path"`
	DatabaseURL        string        `mapstructure:"database_url"`
	LoggingEnabled     bool          `mapstructure:"logging_enabled"`
}

// LoadConfig loads configuration from ./config.yaml or defaults
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  No .env file found, using system environment variables")
	}

	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("port", "8080")
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("storage", "memory")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠ No config.yaml found, using environment/defaults")
	}

	// Unmarshal all keys (including file_path, database_path) into the struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	// viper.Unmarshal can't decode duration strings — parse them manually
	cfg.AccessTokenExpiry, _ = time.ParseDuration(viper.GetString("access_token_expiry"))
	cfg.RefreshTokenExpiry, _ = time.ParseDuration(viper.GetString("refresh_token_expiry"))

	// .env overrides for secrets
	if v := viper.GetString("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := viper.GetString("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}

	return &cfg
}

// InitConfig bootstraps config.yaml and .env in the current directory
func InitConfig() error {
	configPath := "config.yaml"
	envPath := ".env"
	dataDir := "ezauth-data"
	reader := bufio.NewReader(os.Stdin)

	// 1. Prevent overwriting
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config.yaml already exists - delete it first to re-initialize")
	}

	// 2. Interactive Storage Selection
	fmt.Println("--- EZauth Setup ---")
	fmt.Println("Select Storage Method:")
	fmt.Println("1. Memory (Data lost on restart)")
	fmt.Println("2. JSON File (Local ezauth-data/users.json)")
	fmt.Println("3. SQLite (Local ezauth-data/ezauth.db)")
	fmt.Println("4. Postgres (External Database)")
	fmt.Println("5. MySQL (External Database)")
	fmt.Print("Choice (1-5): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	storageMap := map[string]string{
		"1": "memory",
		"2": "file",
		"3": "sqlite",
		"4": "postgres",
		"5": "mysql",
	}

	selectedStorage := storageMap[choice]
	if selectedStorage == "" {
		selectedStorage = "memory"
	}

	// 3. Prepare Database DSN (Connection String)
	var dsn string
	if selectedStorage == "postgres" || selectedStorage == "mysql" {
		fmt.Printf("\nEnter your %s URL (DSN):\n", strings.Title(selectedStorage))
		if selectedStorage == "postgres" {
			fmt.Println("Example: postgres://user:pass@localhost:5432/dbname?sslmode=disable")
		} else {
			fmt.Println("Example: root:password@tcp(127.0.0.1:3306)/dbname")
		}
		fmt.Print("DSN: ")
		dsn, _ = reader.ReadString('\n')
		dsn = strings.TrimSpace(dsn)
	}

	// 4. Create .env for Secrets
	// This is the ONLY place dsn and jwtSecret should be written
	jwtSecret := randomSecret()
	envContent := fmt.Sprintf("JWT_SECRET=%s\nDATABASE_URL=%s\n", jwtSecret, dsn)
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env: %v", err)
	}

	// 5. Create Data Directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	absDataDir := filepath.Join(cwd, dataDir)
	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// 6. Tell Viper which file we are working with
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 7. Set viper values for config.yaml
	viper.Set("PORT", "8080")
	viper.Set("HOST", "127.0.0.1")
	viper.Set("STORAGE", selectedStorage)
	viper.Set("ACCESS_TOKEN_EXPIRY", "5m")
	viper.Set("REFRESH_TOKEN_EXPIRY", "168h")
	viper.Set("LOGGING_ENABLED", true)

	// Absolute paths so ezauth can be run from any directory
	viper.Set("FILE_PATH", filepath.Join(absDataDir, "users.json"))
	viper.Set("DATABASE_PATH", filepath.Join(absDataDir, "ezauth.db"))

	// REMOVED: viper.Set("DATABASE_URL", dsn)
	// By not "Setting" it here, it won't show up in config.yaml

	// 8. Write config.yaml
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config.yaml: %v", err)
	}

	// 9. Create users.json if file mode was selected
	if selectedStorage == "file" {
		usersPath := filepath.Join(absDataDir, "users.json")
		_ = os.WriteFile(usersPath, []byte("{}"), 0644)
	}

	log.Println("\n✅ EZauth initialized successfully!")
	log.Println("🔑 Secrets saved to .env")
	log.Println("⚙️  Configuration saved to config.yaml")
	return nil
}

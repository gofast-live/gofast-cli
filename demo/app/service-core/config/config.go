package config

import (
	"os"
	"strings"
	"time"
)

func isRunningTest() bool {
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") {
			return true
		}
	}

	return false
}

func MustSetEnv(active bool, key string) string {
	value := os.Getenv(key)
	if active && value == "" {
		if isRunningTest() {
			return "test"
		}
		panic("Missing environment variable: " + key)
	}

	return os.Getenv(key)
}

type Config struct {
	// General
	LogLevel  string
	HTTPPort  string
	GRPCPort  string
	Domain    string
	CoreURL   string
	AdminURL  string
	ClientURL string
	TaskToken string

	// Constants
	MaxFileSize     int64
	HTTPTimeout     time.Duration
	ContextTimeout  time.Duration
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration

	// Database
	DatabaseProvider string
	// Postgres
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	// Turso
	TursoURL   string
	TursoToken string

	// OAuth
	GithubClientID        string
	GithubClientSecret    string
	GoogleClientID        string
	GoogleClientSecret    string
	MicrosoftClientID     string
	MicrosoftClientSecret string
}

func LoadConfig() *Config {
	const (
		HTTPTimeout     = 10 * time.Second
		ContextTimeout  = 10 * time.Second
		AccessTokenExp  = 15 * time.Minute
		RefreshTokenExp = 30 * 24 * time.Hour
		MaxFileSize     = 10 << 20
	)
	return &Config{
		LogLevel:              MustSetEnv(true, "LOG_LEVEL"),
		HTTPPort:              MustSetEnv(true, "HTTP_PORT"),
		GRPCPort:              MustSetEnv(true, "GRPC_PORT"),
		Domain:                MustSetEnv(true, "DOMAIN"),
		CoreURL:               MustSetEnv(true, "CORE_URL"),
		AdminURL:              MustSetEnv(true, "ADMIN_URL"),
		ClientURL:             MustSetEnv(true, "CLIENT_URL"),
		TaskToken:             MustSetEnv(true, "TASK_TOKEN"),
		HTTPTimeout:           HTTPTimeout,
		ContextTimeout:        ContextTimeout,
		AccessTokenExp:        AccessTokenExp,
		RefreshTokenExp:       RefreshTokenExp,
		MaxFileSize:           MaxFileSize,
		DatabaseProvider:      MustSetEnv(true, "DATABASE_PROVIDER"),
		PostgresHost:          MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "postgres", "POSTGRES_HOST"),
		PostgresPort:          MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "postgres", "POSTGRES_PORT"),
		PostgresDB:            MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "postgres", "POSTGRES_DB"),
		PostgresUser:          MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "postgres", "POSTGRES_USER"),
		PostgresPassword:      MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "postgres", "POSTGRES_PASSWORD"),
		TursoURL:              MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "turso", "TURSO_URL"),
		TursoToken:            MustSetEnv(os.Getenv("DATABASE_PROVIDER") == "turso", "TURSO_TOKEN"),
		GithubClientID:        os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret:    os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:        os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:    os.Getenv("GOOGLE_CLIENT_SECRET"),
		MicrosoftClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		MicrosoftClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
	}
}

func LoadTestConfig() *Config {
	const (
		HTTPTimeout                = 10 * time.Second
		ContextTimeout             = 10 * time.Second
		AccessTokenExp             = 5 * time.Minute
		RefreshTokenExp            = 30 * 24 * time.Hour
		MaxFileSize                = 10 << 20
	)
	return &Config{
		LogLevel:              "debug",
		HTTPPort:              "8080",
		GRPCPort:              "50051",
		Domain:                "localhost",
		CoreURL:               "http://localhost:8080",
		AdminURL:              "http://localhost:8080",
		ClientURL:             "http://localhost:3000",
		TaskToken:             "test",
		HTTPTimeout:           HTTPTimeout,
		ContextTimeout:        ContextTimeout,
		AccessTokenExp:        AccessTokenExp,
		RefreshTokenExp:       RefreshTokenExp,
		MaxFileSize:           MaxFileSize,
		DatabaseProvider:      "postgres",
		PostgresHost:          "localhost",
		PostgresPort:          "5432",
		PostgresDB:            "test",
		PostgresUser:          "test",
		PostgresPassword:      "test",
		TursoURL:              "http://localhost:8080",
		TursoToken:            "test",
		GithubClientID:        "github_client_id",
		GithubClientSecret:    "github_client_secret",
		GoogleClientID:        "google_client_id",
		GoogleClientSecret:    "google_client_secret",
		MicrosoftClientID:     "microsoft_client_id",
		MicrosoftClientSecret: "microsoft_client_secret",
	}
}

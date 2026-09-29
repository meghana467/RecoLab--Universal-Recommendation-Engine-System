package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GorseAPIURL       string
	GorseDashboardURL string
	GorseAPIKey       string
	RedisURL          string
	BackendPort       int
	BackendHost       string
	TenantID          string

	// PostgreSQL — used by the backend to persist events and cache enriched metadata
	PostgresDSN string
}

func Load() (*Config, error) {
	host := getEnv("POSTGRES_HOST", "localhost")
	portStr := getEnv("POSTGRES_PORT", "5435")
	user := getEnv("POSTGRES_USER", "gorse")
	password := getEnv("POSTGRES_PASSWORD", "gorse_pass")
	db := getEnv("POSTGRES_DB", "gorse")
	sslMode := getEnv("POSTGRES_SSLMODE", "disable")

	pgPort, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, pgPort, user, password, db, sslMode,
	)

	bPortStr := getEnv("BACKEND_PORT", "3001")
	bPort, err := strconv.Atoi(bPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid BACKEND_PORT: %w", err)
	}

	return &Config{
		GorseAPIURL:       getEnv("GORSE_API_URL", "http://localhost:8087"),
		GorseDashboardURL: getEnv("GORSE_DASHBOARD_URL", "http://localhost:8088"),
		GorseAPIKey:       getEnv("GORSE_API_KEY", "recolab_api_key_2026"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		BackendHost:       getEnv("BACKEND_HOST", "0.0.0.0"),
		TenantID:          getEnv("TENANT_ID", "recolab_demo"),
		BackendPort:       bPort,
		PostgresDSN:       dsn,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

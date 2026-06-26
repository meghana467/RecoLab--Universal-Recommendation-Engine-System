package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GorseAPIURL      string
	GorseDashboardURL string
	GorseAPIKey      string
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisURL         string
	BackendPort      int
	BackendHost      string
	TenantID         string
}

func Load() (*Config, error) {
	cfg := &Config{
		GorseAPIURL:      getEnv("GORSE_API_URL", "http://localhost:8087"),
		GorseDashboardURL: getEnv("GORSE_DASHBOARD_URL", "http://localhost:8088"),
		GorseAPIKey:      getEnv("GORSE_API_KEY", "recolab_api_key_2026"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresUser:     getEnv("POSTGRES_USER", "gorse"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "gorse_pass"),
		PostgresDB:       getEnv("POSTGRES_DB", "gorse"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		BackendHost:      getEnv("BACKEND_HOST", "0.0.0.0"),
		TenantID:         getEnv("TENANT_ID", "recolab_demo"),
	}

	portStr := getEnv("POSTGRES_PORT", "5435")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
	}
	cfg.PostgresPort = port

	bPortStr := getEnv("BACKEND_PORT", "3001")
	bPort, err := strconv.Atoi(bPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid BACKEND_PORT: %w", err)
	}
	cfg.BackendPort = bPort

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

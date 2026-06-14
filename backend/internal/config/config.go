package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	CORSOrigins []string
}

// Load reads configuration from the environment (and an optional .env file).
// It returns an error if any required value is missing or invalid so the
// process fails fast at startup instead of at first request.
func Load() (*Config, error) {
	// .env is optional: in production the platform injects real env vars.
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigins: splitAndTrim(getEnv("CORS_ORIGINS", "http://localhost:3000")),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" || cfg.JWTSecret == "change-me-to-a-long-random-secret" {
		return nil, fmt.Errorf("JWT_SECRET is required and must not be the default value")
	}

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "72"))
	if err != nil || ttlHours <= 0 {
		return nil, fmt.Errorf("JWT_TTL_HOURS must be a positive integer")
	}
	cfg.JWTTTL = time.Duration(ttlHours) * time.Hour

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds validated runtime configuration (env-only).
type Config struct {
	APIPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// CORSAllowedOrigins is explicit allowlist; empty disables CORS middleware (same-origin only).
	CORSAllowedOrigins []string

	JWTSecret            string
	JWTIssuer            string
	JWTAudience          string
	JWTAccessTTL         time.Duration
	PlatformClientID     string
	PlatformClientSecret string

	// MigrationExpectedVersion: if > 0, /readyz checks schema_migrations.version >= this and dirty=0.
	MigrationExpectedVersion int

	// Gin mode: "debug", "release", "test"
	GinMode string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	c := &Config{
		APIPort: strings.TrimSpace(os.Getenv("API_PORT")),
		DBHost:  strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:  strings.TrimSpace(os.Getenv("DB_PORT")),
		DBUser:  strings.TrimSpace(os.Getenv("DB_USER")),
		DBPass:  os.Getenv("DB_PASS"), // allow whitespace in password
		DBName:  strings.TrimSpace(os.Getenv("DB_NAME")),

		JWTSecret:            strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTIssuer:            strings.TrimSpace(envOrDefault("JWT_ISSUER", "go-lab")),
		JWTAudience:          strings.TrimSpace(envOrDefault("JWT_AUDIENCE", "go-lab-api")),
		PlatformClientID:     strings.TrimSpace(os.Getenv("PLATFORM_CLIENT_ID")),
		PlatformClientSecret: os.Getenv("PLATFORM_CLIENT_SECRET"),

		GinMode: strings.TrimSpace(envOrDefault("GIN_MODE", "release")),
	}

	if c.APIPort == "" {
		c.APIPort = "5000"
	}
	if c.DBPort == "" {
		c.DBPort = "3306"
	}
	if c.DBName == "" {
		c.DBName = "todosdb"
	}
	if c.DBHost == "" {
		return nil, errors.New("DB_HOST is required")
	}
	if c.DBUser == "" {
		return nil, errors.New("DB_USER is required")
	}

	ttlStr := strings.TrimSpace(envOrDefault("JWT_ACCESS_TTL_SECONDS", "900"))
	ttlSec, err := strconv.Atoi(ttlStr)
	if err != nil || ttlSec < 60 || ttlSec > 86400 {
		return nil, fmt.Errorf("JWT_ACCESS_TTL_SECONDS must be between 60 and 86400, got %q", ttlStr)
	}
	c.JWTAccessTTL = time.Duration(ttlSec) * time.Second

	if len(c.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET is required and must be at least 32 characters")
	}

	if c.PlatformClientID == "" || c.PlatformClientSecret == "" {
		return nil, errors.New("PLATFORM_CLIENT_ID and PLATFORM_CLIENT_SECRET are required for token issuance")
	}

	if v := strings.TrimSpace(os.Getenv("MIGRATION_EXPECTED_VERSION")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("MIGRATION_EXPECTED_VERSION must be a non-negative integer, got %q", v)
		}
		c.MigrationExpectedVersion = n
	}

	corsOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if corsOrigins != "" {
		for _, o := range strings.Split(corsOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				c.CORSAllowedOrigins = append(c.CORSAllowedOrigins, o)
			}
		}
	} else {
		c.CORSAllowedOrigins = []string{"http://localhost:4200"}
	}

	return c, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

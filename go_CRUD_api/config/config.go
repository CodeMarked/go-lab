package config

import (
	"errors"
	"fmt"
	"net/http"
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

	// Browser session cookies (HttpOnly; never expose to SPA JS).
	SessionCookieName     string
	SessionCookieSecure   bool
	SessionSameSiteMode   http.SameSite
	SessionIdleTTL        time.Duration
	SessionAbsoluteTTL    time.Duration

	// AuthBootstrapEnabled keeps POST /api/v1/auth/bootstrap available (temporary bridge).
	AuthBootstrapEnabled bool

	// JWTActiveKeyID is reserved for signing-key rotation telemetry (unused until multi-key JWTs land).
	JWTActiveKeyID string

	// JWTSecretPrevious: optional prior HS256 secret; ParseAccessToken accepts tokens signed with either (rotation window).
	JWTSecretPrevious string

	// CSRF double-submit: non-HttpOnly cookie + matching header on cookie-session mutating requests.
	CSRFCookieName  string
	CSRFHeaderName  string
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

		SessionCookieName:   strings.TrimSpace(envOrDefault("SESSION_COOKIE_NAME", "gl_session")),
		SessionCookieSecure: parseEnvBoolDefaultTrue("SESSION_COOKIE_SECURE"),
		JWTActiveKeyID:      strings.TrimSpace(os.Getenv("JWT_ACTIVE_KEY_ID")),
		JWTSecretPrevious:   strings.TrimSpace(os.Getenv("JWT_SECRET_PREVIOUS")),

		CSRFCookieName: strings.TrimSpace(envOrDefault("CSRF_COOKIE_NAME", "gl_csrf")),
		CSRFHeaderName: strings.TrimSpace(envOrDefault("CSRF_HEADER_NAME", "X-CSRF-Token")),
	}
	if c.SessionCookieName == "" {
		c.SessionCookieName = "gl_session"
	}
	if c.CSRFCookieName == "" {
		c.CSRFCookieName = "gl_csrf"
	}
	if c.CSRFHeaderName == "" {
		c.CSRFHeaderName = "X-CSRF-Token"
	}
	c.SessionSameSiteMode = parseSameSite(envOrDefault("SESSION_SAMESITE", "Lax"))

	idleSec, err := strconv.Atoi(strings.TrimSpace(envOrDefault("SESSION_IDLE_TTL_SECONDS", "1800")))
	if err != nil || idleSec < 60 || idleSec > 86400*7 {
		return nil, fmt.Errorf("SESSION_IDLE_TTL_SECONDS must be between 60 and 604800, got %q", envOrDefault("SESSION_IDLE_TTL_SECONDS", "1800"))
	}
	c.SessionIdleTTL = time.Duration(idleSec) * time.Second

	absSec, err := strconv.Atoi(strings.TrimSpace(envOrDefault("SESSION_ABSOLUTE_TTL_SECONDS", "86400")))
	if err != nil || absSec < 300 || absSec > 86400*30 {
		return nil, fmt.Errorf("SESSION_ABSOLUTE_TTL_SECONDS must be between 300 and 2592000, got %q", envOrDefault("SESSION_ABSOLUTE_TTL_SECONDS", "86400"))
	}
	c.SessionAbsoluteTTL = time.Duration(absSec) * time.Second
	if c.SessionAbsoluteTTL < c.SessionIdleTTL {
		return nil, errors.New("SESSION_ABSOLUTE_TTL_SECONDS must be >= SESSION_IDLE_TTL_SECONDS")
	}

	c.AuthBootstrapEnabled = parseEnvBoolDefaultTrue("AUTH_BOOTSTRAP_ENABLED")

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
	if c.JWTSecretPrevious != "" && len(c.JWTSecretPrevious) < 32 {
		return nil, errors.New("JWT_SECRET_PREVIOUS must be at least 32 characters when set")
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

// parseEnvBoolDefaultTrue returns true when unset; when set, only 1/true/yes (case-insensitive) are true.
func parseEnvBoolDefaultTrue(key string) bool {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		return true
	}
	v := strings.TrimSpace(strings.ToLower(raw))
	return v == "1" || v == "true" || v == "yes"
}

func parseSameSite(s string) http.SameSite {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

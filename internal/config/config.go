package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/leventa-backend/internal/domain/entity"
)

// Config holds all application configuration.
type Config struct {
	AppEnv  string
	AppHost string // empty = all interfaces (0.0.0.0)
	AppPort string

	DatabaseURL string

	JWTSecret            string
	JWTAccessExpiration  int // seconds
	JWTRefreshExpiration int // seconds

	UploadPath    string
	LogFile       string // optional: also append JSON logs to this file
	DocumentPath  string // private ownership documents; must not be a static route
	UploadMaxSize int64  // bytes

	RateLimitMax    int
	RateLimitWindow int // seconds

	// Per-IP limit for every /api/v1 request (guests are anonymous, so this is
	// the main protection against scraping and request floods).
	APIRateLimitMax    int
	APIRateLimitWindow int // seconds
	// Per-IP limit for image and document uploads.
	UploadRateLimitMax    int
	UploadRateLimitWindow int // seconds

	// CORSAllowedOrigins is a comma-separated allowlist. Empty means no
	// cross-origin browser access (mobile clients are not affected by CORS).
	CORSAllowedOrigins string

	// CommissionDefaultPercentage is applied to every finalized deal (admin may override while pending).
	CommissionDefaultPercentage float64

	// Brute-force protection / sessions.
	MaxLoginAttempts  int
	LockoutMinutes    int
	MaxActiveSessions int

	// TrustedProxies is a comma-separated list of reverse proxy IPs/CIDRs whose
	// X-Forwarded-For header may be trusted (needed for correct per-IP rate limiting behind a proxy).
	TrustedProxies string
}

// Load reads configuration from environment variables.
// In local environment, it also loads from .env file.
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "local"),
		AppHost: getEnv("APP_HOST", ""),
		AppPort: getEnv("APP_PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTAccessExpiration:  getEnvAsInt("JWT_ACCESS_EXPIRATION", 900),
		JWTRefreshExpiration: getEnvAsInt("JWT_REFRESH_EXPIRATION", 604800),

		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		DocumentPath:  getEnv("DOCUMENT_PATH", "./documents"),
		LogFile:       getEnv("LOG_FILE", ""),
		UploadMaxSize: int64(getEnvAsInt("UPLOAD_MAX_SIZE", 5242880)),

		RateLimitMax:    getEnvAsInt("RATE_LIMIT_MAX", 5),
		RateLimitWindow: getEnvAsInt("RATE_LIMIT_WINDOW", 60),

		APIRateLimitMax:       getEnvAsInt("API_RATE_LIMIT_MAX", 300),
		APIRateLimitWindow:    getEnvAsInt("API_RATE_LIMIT_WINDOW", 60),
		UploadRateLimitMax:    getEnvAsInt("UPLOAD_RATE_LIMIT_MAX", 30),
		UploadRateLimitWindow: getEnvAsInt("UPLOAD_RATE_LIMIT_WINDOW", 60),

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", ""),

		MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutMinutes:    getEnvAsInt("LOCKOUT_MINUTES", 15),
		MaxActiveSessions: getEnvAsInt("MAX_ACTIVE_SESSIONS", 5),
		TrustedProxies:    getEnv("TRUSTED_PROXIES", ""),

		CommissionDefaultPercentage: getEnvAsFloat("COMMISSION_DEFAULT_PERCENTAGE", entity.DefaultCommissionPercentage),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.CommissionDefaultPercentage < 0 || cfg.CommissionDefaultPercentage > 100 {
		return nil, fmt.Errorf("COMMISSION_DEFAULT_PERCENTAGE must be between 0 and 100")
	}
	if cfg.AppEnv == "production" && len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}

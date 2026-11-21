package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds externalized application settings loaded from environment variables.
type Config struct {
	Port                    string
	DatabaseURL             string
	ReadDatabaseURL         string
	RedisAddr               string
	TokenSecret             string
	AnonymizationSalt       string
	EmailPollInterval       time.Duration
	ProviderTimeout         time.Duration
	CacheTTL                time.Duration
	RateLimitRequestsPerMin int
	WebTemplatesDir         string
	WebStaticDir            string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "")
	readDBURL := getEnv("READ_DATABASE_URL", dbURL)
	redisAddr := getEnv("REDIS_ADDR", "")
	tokenSecret := getEnv("TOKEN_SECRET", "trotamundos-v2-hmac-signing-key-2026")
	anonSalt := getEnv("ANONYMIZATION_SALT", "trotamundos-v2-privacy-salt-9f82a1c4")
	pollSec := getEnvInt("EMAIL_POLL_INTERVAL_SECONDS", 60)
	rateLimit := getEnvInt("RATE_LIMIT_RPM", 240)
	templatesDir := getEnv("WEB_TEMPLATES_DIR", "web/templates")
	staticDir := getEnv("WEB_STATIC_DIR", "web/static")

	return Config{
		Port:                    port,
		DatabaseURL:             dbURL,
		ReadDatabaseURL:         readDBURL,
		RedisAddr:               redisAddr,
		TokenSecret:             tokenSecret,
		AnonymizationSalt:       anonSalt,
		EmailPollInterval:       time.Duration(pollSec) * time.Second,
		ProviderTimeout:         3 * time.Second,
		CacheTTL:                30 * time.Second,
		RateLimitRequestsPerMin: rateLimit,
		WebTemplatesDir:         templatesDir,
		WebStaticDir:            staticDir,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

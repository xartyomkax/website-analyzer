package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	Env             string
	RequestTimeout  time.Duration
	LinkTimeout     time.Duration
	MaxWorkers      int
	MaxResponseSize int64
	MaxURLLength    int
	MaxRedirects    int
}

func LoadConfig() *Config {
	// Default values are defined in docs/specs/REQUIREMENTS.md
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "production"),
		RequestTimeout:  getEnvDuration("REQUEST_TIMEOUT", 30*time.Second),
		LinkTimeout:     getEnvDuration("LINK_CHECK_TIMEOUT", 5*time.Second),
		MaxWorkers:      getEnvInt("MAX_WORKERS", 10),
		MaxResponseSize: getEnvInt64("MAX_RESPONSE_SIZE", 10*1024*1024), // 10MB
		MaxURLLength:    getEnvInt("MAX_URL_LENGTH", 2048),
		MaxRedirects:    getEnvInt("MAX_REDIRECTS", 10),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort        string
	DatabaseURL     string
	RedisAddr       string
	JWTSecret       string
	JWTExpiry       time.Duration
	RemotiveAPIURL  string
	RateLimitPerMin int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		DatabaseURL:     dbURL,
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:       jwtSecret,
		JWTExpiry:       72 * time.Hour,
		RemotiveAPIURL:  getEnv("REMOTIVE_API_URL", "https://remotive.com/api/remote-jobs"),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 60),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

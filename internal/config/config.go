package config

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	JWTAccessExpire  time.Duration
	JWTRefreshExpire time.Duration
	Port             string
}

func Load() (*Config, error) {
	// Load .env file (ignore error if not exists)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "restaurant_booking"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		Port:       getEnv("PORT", "8080"),
	}

	// Validate required fields
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	// Parse JWT expiration times
	accessExpire := getEnv("JWT_ACCESS_EXPIRE", "15m")
	refreshExpire := getEnv("JWT_REFRESH_EXPIRE", "7d")

	var err error
	cfg.JWTAccessExpire, err = time.ParseDuration(accessExpire)
	if err != nil {
		return nil, errors.New("invalid JWT_ACCESS_EXPIRE format")
	}

	cfg.JWTRefreshExpire, err = time.ParseDuration(refreshExpire)
	if err != nil {
		return nil, errors.New("invalid JWT_REFRESH_EXPIRE format")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	DB     DatabaseConfig
	Server ServerConfig
	Auth   AuthConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
}

type AuthConfig struct {
	Secret     string
	Expiration int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{}

	cfg.DB.Port = os.Getenv("DB_PORT")
	cfg.DB.Host = os.Getenv("DB_HOST")
	cfg.DB.User = os.Getenv("DB_USER")
	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	cfg.DB.DBName = os.Getenv("DB_NAME")
	cfg.DB.SSLMode = os.Getenv("DB_SSLMODE")

	cfg.Server.Port = os.Getenv("SERVER_PORT")

	cfg.Auth.Secret = os.Getenv("JWT_SECRET")
	expiration, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	cfg.Auth.Expiration = expiration

	return cfg, nil
}

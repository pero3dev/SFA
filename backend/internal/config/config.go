package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() Config {
	return Config{
		Env:             getEnv("APP_ENV", "local"),
		Port:            getEnv("APP_PORT", "8080"),
		DatabaseURL:     getEnv("APP_DATABASE_URL", "postgresql://sfa:sfa@localhost:5432/sfa?sslmode=disable"),
		JWTIssuer:       getEnv("APP_JWT_ISSUER", "sfa-local"),
		JWTAudience:     getEnv("APP_JWT_AUDIENCE", "sfa-web"),
		AccessTokenTTL:  time.Duration(getEnvInt("APP_JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL: time.Duration(getEnvInt("APP_JWT_REFRESH_TTL_HOURS", 720)) * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

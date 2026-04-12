package config

import (
	"log"
	"os"
)

type Config struct {
	BrowserlessURL   string
	BrowserlessToken string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
}

func Load() *Config {
	cfg := &Config{
		BrowserlessURL:   getEnv("BROWSERLESS_URL", "http://browserless:3000/screenshot"),
		BrowserlessToken: getEnv("BROWSERLESS_TOKEN", ""),
		DBHost:           getEnv("POSTGRES_HOST", "postgres"),
		DBPort:           getEnv("POSTGRES_PORT", "5432"),
		DBUser:           getEnv("POSTGRES_USER", ""),
		DBPassword:       getEnv("POSTGRES_PASSWORD", ""),
		DBName:           getEnv("POSTGRES_DB", ""),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		log.Fatal("POSTGRES_USER, POSTGRES_PASSWORD and POSTGRES_DB are required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

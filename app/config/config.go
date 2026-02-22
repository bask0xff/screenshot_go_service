package config

import "os"

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	BrowserlessURL string
}

func Load() *Config {
	return &Config{
		DBHost:         getEnv("POSTGRES_HOST", "postgres"),
		DBPort:         getEnv("POSTGRES_PORT", "5432"),
		DBUser:         getEnv("POSTGRES_USER", "admin"),
		DBPassword:     getEnv("POSTGRES_PASSWORD", ""),
		DBName:         getEnv("POSTGRES_DB", "mydata"),
		BrowserlessURL: "http://browserless:3000/screenshot?token=" + getEnv("BROWSERLESS_TOKEN", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

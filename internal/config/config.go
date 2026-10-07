package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr            string
	Store           string
	DatabaseURL     string
	DefaultLanguage string
	AITimeout       time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Addr:            getenv("ADDR", ":8080"),
		Store:           getenv("STORE", "memory"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		DefaultLanguage: getenv("DEFAULT_LANGUAGE", "tr"),
		AITimeout:       10 * time.Second,
	}
	if value := os.Getenv("AI_TIMEOUT"); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds <= 0 {
			return Config{}, fmt.Errorf("AI_TIMEOUT must be a positive number of seconds")
		}
		cfg.AITimeout = time.Duration(seconds) * time.Second
	}
	if cfg.Store != "memory" && cfg.Store != "postgres" {
		return Config{}, fmt.Errorf("STORE must be memory or postgres")
	}
	if cfg.Store == "postgres" && cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required when STORE=postgres")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

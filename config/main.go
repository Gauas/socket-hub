package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env          string
	ServiceName  string
	Port         string
	APIKey       string
	WriteTimeout time.Duration
}

func New() *Config {
	return &Config{
		Env:          get("ENV", "development"),
		ServiceName:  "socket-hub",
		Port:         get("PORT", "8085"),
		APIKey:       get("SOCKET_API_KEY", ""),
		WriteTimeout: time.Duration(getInt("WRITE_TIMEOUT_SECONDS", 10)) * time.Second,
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

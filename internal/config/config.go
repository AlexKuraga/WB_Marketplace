package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPHost        = "0.0.0.0"
	defaultHTTPPort        = 8080
	defaultDatabaseURL     = "postgres://postgres:postgres@localhost:5432/wb_marketplace?sslmode=disable"
	defaultShutdownTimeout = 10 * time.Second
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	HTTPHost        string
	HTTPPort        int
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

// HTTPAddr returns the listen address for the HTTP server.
func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	port, err := envInt("HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_PORT: %w", err)
	}

	shutdownSec, err := envInt("SHUTDOWN_TIMEOUT_SEC", int(defaultShutdownTimeout.Seconds()))
	if err != nil {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT_SEC: %w", err)
	}

	host := envString("HTTP_HOST", defaultHTTPHost)
	databaseURL := envString("DATABASE_URL", defaultDatabaseURL)
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}

	return Config{
		HTTPHost:        host,
		HTTPPort:        port,
		DatabaseURL:     databaseURL,
		ShutdownTimeout: time.Duration(shutdownSec) * time.Second,
	}, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", v)
	}
	return n, nil
}

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort       = 8080
	defaultMaxWorkers = 4
)

// Config holds application configuration sourced from environment variables.
type Config struct {
	// DatabaseURL is the Postgres DSN. Required. Env: DATABASE_URL.
	DatabaseURL string
	// Port is the HTTP listen port. Env: PORT. Default: 8080.
	Port int
	// Debug enables verbose logging. Env: DEBUG. Default: false.
	Debug bool
	// MaxWorkers is the worker-pool size. Env: MAX_WORKERS. Default: 4.
	MaxWorkers int
}

// Load reads configuration from environment variables.
// Returns an error if a required variable is missing or a numeric variable
// cannot be parsed.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        defaultPort,
		Debug:       os.Getenv("DEBUG") == "true",
		MaxWorkers:  defaultMaxWorkers,
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("config: DATABASE_URL is required")
	}

	if raw := os.Getenv("PORT"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("config: invalid PORT %q: %w", raw, err)
		}
		cfg.Port = v
	}

	if raw := os.Getenv("MAX_WORKERS"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("config: invalid MAX_WORKERS %q: %w", raw, err)
		}
		cfg.MaxWorkers = v
	}

	return cfg, nil
}

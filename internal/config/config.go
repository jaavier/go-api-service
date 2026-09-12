package config

import (
	"os"
	"strconv"
)

// Config holds application configuration.
// BUG: no validation, no defaults documented, silently uses zero-values.
type Config struct {
	DatabaseURL string
	Port        int
	Debug       bool
	MaxWorkers  int
}

// Load reads config from environment.
// BUG: errors from strconv are silently ignored.
// BUG: no required-field validation.
func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))       // error discarded
	workers, _ := strconv.Atoi(os.Getenv("MAX_WORKERS")) // error discarded

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        port,
		Debug:       os.Getenv("DEBUG") == "true",
		MaxWorkers:  workers,
	}
}

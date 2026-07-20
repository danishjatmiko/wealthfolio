// Package config loads Wealthfolio's runtime configuration from environment
// variables.
package config

import "os"

// Config holds all runtime configuration for the API server.
type Config struct {
	DatabaseURL string
	Port        string
	CORSOrigin  string
}

// Load reads configuration from environment variables, applying defaults
// where the spec allows it.
func Load() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        getEnvOr("PORT", "8080"),
		CORSOrigin:  getEnvOr("CORS_ORIGIN", "http://localhost:5173"),
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

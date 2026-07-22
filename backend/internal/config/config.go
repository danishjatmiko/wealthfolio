// Package config loads Wealthfolio's runtime configuration from environment
// variables.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	DatabaseURL string
	Port        string
	CORSOrigin  string

	// GoogleClientID/GoogleClientSecret identify the OAuth 2.0 client
	// registered in Google Cloud Console. GoogleRedirectURL must exactly
	// match one of that client's authorized redirect URIs.
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// AppBaseURL is the frontend's origin; the OAuth callback redirects
	// here once a session cookie is set.
	AppBaseURL string

	// CookieSecure controls the session cookie's Secure flag. Defaults to
	// true iff AppBaseURL is https, since a Secure cookie is silently
	// dropped by browsers over plain HTTP (e.g. local dev), but can be
	// overridden explicitly via COOKIE_SECURE.
	CookieSecure bool
}

// Load reads configuration from environment variables, applying defaults
// where the spec allows it.
func Load() Config {
	appBaseURL := getEnvOr("APP_BASE_URL", "http://localhost:5173")
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        getEnvOr("PORT", "8080"),
		CORSOrigin:  getEnvOr("CORS_ORIGIN", "http://localhost:5173"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getEnvOr("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),

		AppBaseURL:   appBaseURL,
		CookieSecure: getEnvBoolOr("COOKIE_SECURE", strings.HasPrefix(appBaseURL, "https://")),
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBoolOr(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

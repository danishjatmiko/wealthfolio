package httpapi

import (
	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/service"
)

// Handler holds the dependencies shared by every resource's HTTP handlers.
// Simple pass-through CRUD (categories, debts, passive income) reads/writes
// via repos directly; anything with derived business logic goes through
// svc.
type Handler struct {
	repos *db.Repos
	svc   *service.Services

	// appBaseURL is where the OAuth callback sends the browser after
	// login. cookieSecure controls the Secure flag on every cookie this
	// package sets (session + OAuth state).
	appBaseURL   string
	cookieSecure bool
}

// NewHandler builds a Handler backed by the given repositories/services.
func NewHandler(repos *db.Repos, svc *service.Services, cfg config.Config) *Handler {
	return &Handler{repos: repos, svc: svc, appBaseURL: cfg.AppBaseURL, cookieSecure: cfg.CookieSecure}
}

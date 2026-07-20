package httpapi

import (
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
}

// NewHandler builds a Handler backed by the given repositories/services.
func NewHandler(repos *db.Repos, svc *service.Services) *Handler {
	return &Handler{repos: repos, svc: svc}
}

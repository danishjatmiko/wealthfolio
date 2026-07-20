package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/service"
)

// NewRouter builds the full Wealthfolio HTTP router: /healthz (no auth, no
// prefix) plus every /api/v1 resource route, with CORS restricted to
// cfg.CORSOrigin and the fixed-user middleware applied to the API group.
func NewRouter(cfg config.Config, repos *db.Repos, svc *service.Services) http.Handler {
	h := NewHandler(repos, svc)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           int((5 * time.Minute).Seconds()),
	}))

	r.Get("/healthz", h.healthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(CurrentUserMiddleware)

		r.Get("/categories", h.listCategories)

		r.Get("/rates", h.listRates)
		r.Get("/rates/latest", h.getLatestRate)
		r.Post("/rates", h.createRate)

		r.Get("/snapshots", h.listSnapshots)
		r.Get("/snapshots/latest", h.getLatestSnapshot)
		r.Post("/snapshots", h.createSnapshot)
		r.Get("/snapshots/{date}", h.getSnapshotByDate)
		r.Get("/snapshots/{date}/holdings", h.listHoldingsForDate)
		r.Post("/snapshots/{date}/holdings", h.createHolding)

		r.Put("/holdings/{id}", h.updateHolding)
		r.Delete("/holdings/{id}", h.deleteHolding)

		r.Get("/debts", h.listDebts)
		r.Post("/debts", h.createDebt)
		r.Put("/debts/{id}", h.updateDebt)
		r.Delete("/debts/{id}", h.deleteDebt)

		r.Get("/passive-income", h.listPassiveIncome)
		r.Post("/passive-income", h.createPassiveIncome)
		r.Put("/passive-income/{id}", h.updatePassiveIncome)
		r.Delete("/passive-income/{id}", h.deletePassiveIncome)

		r.Get("/targets", h.listTargets)
		r.Post("/targets", h.createTarget)
		r.Put("/targets/{id}", h.updateTarget)
		r.Delete("/targets/{id}", h.deleteTarget)

		r.Get("/dashboard", h.getDashboard)
		r.Get("/progress", h.getProgress)
	})

	return r
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

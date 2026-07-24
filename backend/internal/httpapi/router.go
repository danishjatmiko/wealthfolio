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

// NewRouter builds the full Etherna HTTP router: /healthz (no auth, no
// prefix) plus every /api/v1 resource route, with CORS restricted to
// cfg.CORSOrigin (with credentials, since the session lives in a cookie)
// and AuthMiddleware applied to every route except the Google OAuth
// login/callback, password login, and logout, which by definition run
// without a session.
func NewRouter(cfg config.Config, repos *db.Repos, svc *service.Services) http.Handler {
	h := NewHandler(repos, svc, cfg)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(BodyLimit)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           int((5 * time.Minute).Seconds()),
	}))

	r.Get("/healthz", h.healthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth/google", func(r chi.Router) {
			r.Get("/login", h.googleLogin)
			r.Get("/callback", h.googleCallback)
		})
		r.Post("/auth/logout", h.logout)
		r.Post("/auth/login", h.login)

		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			r.Get("/auth/me", h.me)

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
			r.Delete("/snapshots/{id}", h.deleteSnapshot)

			r.Put("/holdings/{id}", h.updateHolding)
			r.Delete("/holdings/{id}", h.deleteHolding)

			r.Get("/debt-snapshots", h.listDebtSnapshots)
			r.Get("/debt-snapshots/latest", h.getLatestDebtSnapshot)
			r.Post("/debt-snapshots", h.createDebtSnapshot)
			r.Get("/debt-snapshots/{date}", h.getDebtSnapshotByDate)
			r.Post("/debt-snapshots/{date}/entries", h.createDebtEntry)
			r.Delete("/debt-snapshots/{id}", h.deleteDebtSnapshot)

			r.Put("/debt-entries/{id}", h.updateDebtEntry)
			r.Delete("/debt-entries/{id}", h.deleteDebtEntry)

			r.Get("/expense-periods", h.listExpensePeriods)
			r.Get("/expense-periods/latest", h.getLatestExpensePeriod)
			r.Post("/expense-periods", h.createExpensePeriod)
			r.Get("/expense-periods/{id}", h.getExpensePeriod)
			r.Delete("/expense-periods/{id}", h.deleteExpensePeriod)
			r.Post("/expense-periods/{periodId}/envelopes", h.createBudgetEnvelope)
			r.Post("/expense-periods/{periodId}/fixed-expenses", h.createFixedExpense)

			r.Put("/budget-envelopes/{id}", h.updateBudgetEnvelope)
			r.Delete("/budget-envelopes/{id}", h.deleteBudgetEnvelope)

			r.Put("/fixed-expenses/{id}", h.updateFixedExpense)
			r.Delete("/fixed-expenses/{id}", h.deleteFixedExpense)

			r.Get("/expense-source-mappings", h.listExpenseSourceMappings)
			r.Put("/expense-source-mappings/{source}", h.upsertExpenseSourceMapping)
			r.Post("/expense-ingestions", h.ingestExpense)

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
			r.Get("/debt-progress", h.getDebtProgress)
		})
	})

	return r
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

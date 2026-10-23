package httpapi

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	dto, err := h.svc.Dashboard.Get(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

// getDashboardExpense returns the dashboard's expense block for the period
// named by ?period_id=, so the Home page can switch periods without
// refetching the whole dashboard.
func (h *Handler) getDashboardExpense(w http.ResponseWriter, r *http.Request) {
	periodID, err := uuid.Parse(r.URL.Query().Get("period_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid period_id")
		return
	}
	userID := currentUserID(r.Context())
	dto, err := h.svc.Dashboard.Expense(r.Context(), userID, periodID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense period not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

// getDashboardAssets returns the equity hero and allocation as of the asset
// snapshot dated ?snapshot_date=, so the Home page can switch snapshots
// without refetching the whole dashboard.
func (h *Handler) getDashboardAssets(w http.ResponseWriter, r *http.Request) {
	date, err := domain.ParseDate(r.URL.Query().Get("snapshot_date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid snapshot_date")
		return
	}
	userID := currentUserID(r.Context())
	dto, err := h.svc.Dashboard.Assets(r.Context(), userID, date)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "snapshot not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

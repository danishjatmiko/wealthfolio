package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
)

func (h *Handler) listExpensePeriods(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.ExpensePeriods.ListSummaries(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getLatestExpensePeriod(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	detail, err := h.svc.ExpensePeriods.GetLatestDetail(r.Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no expense periods yet")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) getExpensePeriod(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid period id")
		return
	}

	userID := currentUserID(r.Context())
	detail, err := h.svc.ExpensePeriods.GetDetail(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense period not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type createExpensePeriodRequest struct {
	Year          int  `json:"year"`
	Month         int  `json:"month"`
	CopyEnvelopes bool `json:"copy_envelopes"`
}

// createExpensePeriod creates the period named after the given year/month
// (see service.ExpensePeriodsService.Create) — the client always supplies
// the target month explicitly, so any month can be picked, not just the
// next chronological one.
func (h *Handler) createExpensePeriod(w http.ResponseWriter, r *http.Request) {
	var req createExpensePeriodRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Month < 1 || req.Month > 12 {
		writeError(w, http.StatusBadRequest, "month must be between 1 and 12")
		return
	}

	userID := currentUserID(r.Context())
	detail, err := h.svc.ExpensePeriods.Create(r.Context(), userID, req.Year, time.Month(req.Month), req.CopyEnvelopes)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (h *Handler) deleteExpensePeriod(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid period id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.ExpensePeriods.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

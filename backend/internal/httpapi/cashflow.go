package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"wealthfolio/backend/internal/service"
)

func (h *Handler) getCashflow(w http.ResponseWriter, r *http.Request) {
	year := time.Now().In(time.FixedZone("WIB", 7*60*60)).Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}
		year = parsed
	}
	result, err := h.svc.Cashflow.Summary(r.Context(), currentUserID(r.Context()), year)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) saveActiveIncome(w http.ResponseWriter, r *http.Request) {
	var req service.ActiveIncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.svc.Cashflow.SaveActiveIncome(r.Context(), currentUserID(r.Context()), chi.URLParam(r, "month"), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteActiveIncome(w http.ResponseWriter, r *http.Request) {
	err := h.svc.Cashflow.DeleteActiveIncome(r.Context(), currentUserID(r.Context()), chi.URLParam(r, "month"))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package httpapi

import (
	"errors"
	"net/http"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

type rateRequest struct {
	EntryDate string   `json:"entry_date"`
	Antam     *float64 `json:"antam"`
	Kinghalim *float64 `json:"kinghalim"`
	Ubs       *float64 `json:"ubs"`
	UsdIdr    *float64 `json:"usd_idr"`
}

func (h *Handler) listRates(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	rates, err := h.repos.Rates.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rates)
}

func (h *Handler) getLatestRate(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	rate, err := h.repos.Rates.GetLatest(r.Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no rate entries yet")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rate)
}

func (h *Handler) createRate(w http.ResponseWriter, r *http.Request) {
	var req rateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.EntryDate == "" {
		writeError(w, http.StatusBadRequest, "entry_date is required")
		return
	}
	date, err := domain.ParseDate(req.EntryDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Antam == nil || req.Kinghalim == nil || req.Ubs == nil || req.UsdIdr == nil {
		writeError(w, http.StatusBadRequest, "antam, kinghalim, ubs, and usd_idr are required")
		return
	}

	userID := currentUserID(r.Context())
	rate, err := h.repos.Rates.Upsert(r.Context(), userID, date, *req.Antam, *req.Kinghalim, *req.Ubs, *req.UsdIdr)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rate)
}

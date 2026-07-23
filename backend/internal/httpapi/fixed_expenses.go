package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

type fixedExpenseRequest struct {
	Name       *string    `json:"name"`
	AmountIdr  *int64     `json:"amount_idr"`
	EnvelopeID *uuid.UUID `json:"envelope_id"`
}

func (req fixedExpenseRequest) toServiceRequest() service.FixedExpenseRequest {
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	var amount int64
	if req.AmountIdr != nil {
		amount = *req.AmountIdr
	}
	var envelopeID uuid.UUID
	if req.EnvelopeID != nil {
		envelopeID = *req.EnvelopeID
	}
	return service.FixedExpenseRequest{Name: name, AmountIdr: amount, EnvelopeID: envelopeID}
}

func (h *Handler) createFixedExpense(w http.ResponseWriter, r *http.Request) {
	periodID, err := uuid.Parse(chi.URLParam(r, "periodId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid period id")
		return
	}

	var req fixedExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID := currentUserID(r.Context())
	expense, err := h.svc.FixedExpenses.Create(r.Context(), userID, periodID, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, expense)
}

func (h *Handler) updateFixedExpense(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fixed expense id")
		return
	}

	var req fixedExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID := currentUserID(r.Context())
	expense, err := h.svc.FixedExpenses.Update(r.Context(), userID, id, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *Handler) deleteFixedExpense(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fixed expense id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.FixedExpenses.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

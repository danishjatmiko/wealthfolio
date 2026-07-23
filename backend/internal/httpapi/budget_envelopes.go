package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

type budgetEnvelopeRequest struct {
	CategoryID         *uuid.UUID `json:"category_id"`
	Name               *string    `json:"name"`
	CommittedAmountIdr *int64     `json:"committed_amount_idr"`
}

func (req budgetEnvelopeRequest) toServiceRequest() service.BudgetEnvelopeRequest {
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	var amount int64
	if req.CommittedAmountIdr != nil {
		amount = *req.CommittedAmountIdr
	}
	var categoryID uuid.UUID
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}
	return service.BudgetEnvelopeRequest{CategoryID: categoryID, Name: name, CommittedAmountIdr: amount}
}

func (h *Handler) createBudgetEnvelope(w http.ResponseWriter, r *http.Request) {
	periodID, err := uuid.Parse(chi.URLParam(r, "periodId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid period id")
		return
	}

	var req budgetEnvelopeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID := currentUserID(r.Context())
	env, err := h.svc.BudgetEnvelopes.Create(r.Context(), userID, periodID, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, env)
}

func (h *Handler) updateBudgetEnvelope(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid envelope id")
		return
	}

	var req budgetEnvelopeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID := currentUserID(r.Context())
	env, err := h.svc.BudgetEnvelopes.Update(r.Context(), userID, id, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (h *Handler) deleteBudgetEnvelope(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid envelope id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.BudgetEnvelopes.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

type debtEntryRequest struct {
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	ValueIdr  *int64  `json:"value_idr"`
	Direction *string `json:"direction"`
}

func (req debtEntryRequest) toServiceRequest() (service.DebtEntryRequest, error) {
	if req.Name == nil || *req.Name == "" {
		return service.DebtEntryRequest{}, errors.New("name is required")
	}
	if req.Direction == nil {
		return service.DebtEntryRequest{}, errors.New("direction is required")
	}
	debtType := ""
	if req.Type != nil {
		debtType = *req.Type
	}
	var value int64
	if req.ValueIdr != nil {
		value = *req.ValueIdr
	}
	return service.DebtEntryRequest{
		Name:      *req.Name,
		Type:      debtType,
		ValueIdr:  value,
		Direction: *req.Direction,
	}, nil
}

func (h *Handler) createDebtEntry(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req debtEntryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	svcReq, err := req.toServiceRequest()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := currentUserID(r.Context())
	entry, err := h.svc.DebtEntries.Create(r.Context(), userID, date, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (h *Handler) updateDebtEntry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt entry id")
		return
	}

	var req debtEntryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	svcReq, err := req.toServiceRequest()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := currentUserID(r.Context())
	entry, err := h.svc.DebtEntries.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *Handler) deleteDebtEntry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt entry id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.DebtEntries.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

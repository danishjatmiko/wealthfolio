package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type debtRequest struct {
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	ValueIdr  *int64  `json:"value_idr"`
	Direction *string `json:"direction"`
}

func (req debtRequest) validate() error {
	if req.Name == nil || *req.Name == "" {
		return errors.New("name is required")
	}
	if req.ValueIdr == nil {
		return errors.New("value_idr is required")
	}
	if req.Direction == nil || (*req.Direction != "i_owe" && *req.Direction != "owed_to_me") {
		return errors.New("direction must be 'i_owe' or 'owed_to_me'")
	}
	return nil
}

func (h *Handler) listDebts(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	debts, err := h.repos.Debts.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, debts)
}

func (h *Handler) createDebt(w http.ResponseWriter, r *http.Request) {
	var req debtRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	debtType := ""
	if req.Type != nil {
		debtType = *req.Type
	}
	userID := currentUserID(r.Context())
	debt, err := h.repos.Debts.Create(r.Context(), userID, *req.Name, debtType, *req.ValueIdr, *req.Direction)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, debt)
}

func (h *Handler) updateDebt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	var req debtRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	debtType := ""
	if req.Type != nil {
		debtType = *req.Type
	}
	debt, err := h.repos.Debts.Update(r.Context(), id, *req.Name, debtType, *req.ValueIdr, *req.Direction)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, debt)
}

func (h *Handler) deleteDebt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}
	if err := h.repos.Debts.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

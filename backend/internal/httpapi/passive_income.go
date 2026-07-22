package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
)

type passiveIncomeRequest struct {
	CategoryID *int16  `json:"category_id"`
	Name       *string `json:"name"`
	PerYearIdr *int64  `json:"per_year_idr"`
}

func (req passiveIncomeRequest) validate() error {
	if req.CategoryID == nil {
		return errors.New("category_id is required")
	}
	if req.Name == nil || *req.Name == "" {
		return errors.New("name is required")
	}
	if req.PerYearIdr == nil {
		return errors.New("per_year_idr is required")
	}
	return nil
}

func (h *Handler) validateCategory(r *http.Request, id int16) error {
	if _, err := h.repos.Categories.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return errInvalidCategory
		}
		return err
	}
	return nil
}

var errInvalidCategory = errors.New("invalid category_id")

func (h *Handler) listPassiveIncome(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.repos.PassiveIncome.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) createPassiveIncome(w http.ResponseWriter, r *http.Request) {
	var req passiveIncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validateCategory(r, *req.CategoryID); err != nil {
		if errors.Is(err, errInvalidCategory) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleServiceError(w, err)
		return
	}

	userID := currentUserID(r.Context())
	p, err := h.repos.PassiveIncome.Create(r.Context(), userID, *req.CategoryID, *req.Name, *req.PerYearIdr)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) updatePassiveIncome(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req passiveIncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validateCategory(r, *req.CategoryID); err != nil {
		if errors.Is(err, errInvalidCategory) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleServiceError(w, err)
		return
	}

	userID := currentUserID(r.Context())
	p, err := h.repos.PassiveIncome.Update(r.Context(), userID, id, *req.CategoryID, *req.Name, *req.PerYearIdr)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deletePassiveIncome(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID := currentUserID(r.Context())
	if err := h.repos.PassiveIncome.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

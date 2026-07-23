package httpapi

import (
	"net/http"
)

func (h *Handler) listExpenseCategories(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.ExpenseCategories.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type createExpenseCategoryRequest struct {
	Name *string `json:"name"`
}

func (h *Handler) createExpenseCategory(w http.ResponseWriter, r *http.Request) {
	var req createExpenseCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	var name string
	if req.Name != nil {
		name = *req.Name
	}

	userID := currentUserID(r.Context())
	category, err := h.svc.ExpenseCategories.Create(r.Context(), userID, name)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

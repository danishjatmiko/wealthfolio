package httpapi

import "net/http"

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repos.Categories.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

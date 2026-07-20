package httpapi

import "net/http"

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	dto, err := h.svc.Dashboard.Get(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

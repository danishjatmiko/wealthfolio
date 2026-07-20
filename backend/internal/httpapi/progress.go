package httpapi

import "net/http"

func (h *Handler) getProgress(w http.ResponseWriter, r *http.Request) {
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "monthly"
	}

	userID := currentUserID(r.Context())
	dto, err := h.svc.Progress.Get(r.Context(), userID, granularity)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

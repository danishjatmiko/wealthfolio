package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"wealthfolio/backend/internal/service"
)

func (h *Handler) listExpenseSourceMappings(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.ExpenseSourceMappings.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type upsertExpenseSourceMappingRequest struct {
	EnvelopeName *string `json:"envelope_name"`
}

func (h *Handler) upsertExpenseSourceMapping(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")

	var req upsertExpenseSourceMappingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	var envelopeName string
	if req.EnvelopeName != nil {
		envelopeName = *req.EnvelopeName
	}

	userID := currentUserID(r.Context())
	mapping, err := h.svc.ExpenseSourceMappings.Upsert(r.Context(), userID, source, envelopeName)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapping)
}

type ingestExpenseRequest struct {
	IdempotencyKey string    `json:"idempotency_key"`
	Source         string    `json:"source"`
	RawTitle       *string   `json:"raw_title"`
	RawText        *string   `json:"raw_text"`
	RawBigText     *string   `json:"raw_big_text"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// ingestExpense is the Android app's outbox sync target: forwards a
// captured notification (or a retry of one already forwarded), and always
// responds 200 with a "created" or "ignored" status — idempotent on
// idempotency_key, never a 409. See service.ExpenseIngestionService.
func (h *Handler) ingestExpense(w http.ResponseWriter, r *http.Request) {
	var req ingestExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.IdempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}
	if req.Source == "" {
		writeError(w, http.StatusBadRequest, "source is required")
		return
	}
	if req.OccurredAt.IsZero() {
		writeError(w, http.StatusBadRequest, "occurred_at is required")
		return
	}

	userID := currentUserID(r.Context())
	result, err := h.svc.ExpenseIngestion.Ingest(r.Context(), userID, service.ExpenseIngestionRequest{
		IdempotencyKey: req.IdempotencyKey,
		Source:         req.Source,
		RawTitle:       req.RawTitle,
		RawText:        req.RawText,
		RawBigText:     req.RawBigText,
		OccurredAt:     req.OccurredAt,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

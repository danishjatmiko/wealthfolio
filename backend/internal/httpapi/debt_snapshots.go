package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service"
)

func (h *Handler) listDebtSnapshots(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.DebtSnapshots.ListSummaries(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getLatestDebtSnapshot(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	detail, err := h.svc.DebtSnapshots.GetLatestDetail(r.Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no debt snapshots yet")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) getDebtSnapshotByDate(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID := currentUserID(r.Context())
	detail, err := h.svc.DebtSnapshots.GetByDateDetail(r.Context(), userID, date)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "debt snapshot not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type createDebtSnapshotRequest struct {
	SnapshotDate   string             `json:"snapshot_date"`
	CopyFromLatest bool               `json:"copy_from_latest"`
	Entries        []debtEntryRequest `json:"entries"`
}

func (h *Handler) createDebtSnapshot(w http.ResponseWriter, r *http.Request) {
	var req createDebtSnapshotRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.SnapshotDate == "" {
		writeError(w, http.StatusBadRequest, "snapshot_date is required")
		return
	}
	date, err := domain.ParseDate(req.SnapshotDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	entryReqs := make([]service.DebtEntryRequest, 0, len(req.Entries))
	for _, er := range req.Entries {
		svcReq, err := er.toServiceRequest()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		entryReqs = append(entryReqs, svcReq)
	}

	userID := currentUserID(r.Context())
	detail, err := h.svc.DebtSnapshots.Create(r.Context(), userID, date, req.CopyFromLatest, entryReqs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (h *Handler) deleteDebtSnapshot(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt snapshot id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.DebtSnapshots.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

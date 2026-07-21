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

func parseDateParam(r *http.Request) (domain.Date, error) {
	return domain.ParseDate(chi.URLParam(r, "date"))
}

func (h *Handler) listSnapshots(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.Snapshots.ListSummaries(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getLatestSnapshot(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	detail, err := h.svc.Snapshots.GetLatestDetail(r.Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no snapshots yet")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) getSnapshotByDate(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID := currentUserID(r.Context())
	detail, err := h.svc.Snapshots.GetByDateDetail(r.Context(), userID, date)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "snapshot not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type createSnapshotRequest struct {
	SnapshotDate   string           `json:"snapshot_date"`
	CopyFromLatest bool             `json:"copy_from_latest"`
	Holdings       []holdingRequest `json:"holdings"`
}

func (h *Handler) createSnapshot(w http.ResponseWriter, r *http.Request) {
	var req createSnapshotRequest
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

	holdingReqs := make([]service.HoldingRequest, 0, len(req.Holdings))
	for _, hr := range req.Holdings {
		svcReq, err := hr.toServiceRequest()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		holdingReqs = append(holdingReqs, svcReq)
	}

	userID := currentUserID(r.Context())
	detail, err := h.svc.Snapshots.Create(r.Context(), userID, date, req.CopyFromLatest, holdingReqs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (h *Handler) deleteSnapshot(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid snapshot id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.Snapshots.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listHoldingsForDate(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID := currentUserID(r.Context())
	holdings, err := h.svc.Snapshots.ListHoldingsForDate(r.Context(), userID, date)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "snapshot not found")
			return
		}
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, holdings)
}

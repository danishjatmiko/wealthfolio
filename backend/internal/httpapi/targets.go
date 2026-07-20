package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

type targetRequest struct {
	Name               *string  `json:"name"`
	Year               *int     `json:"year"`
	MetricType         *string  `json:"metric_type"`
	TargetValue        *float64 `json:"target_value"`
	Unit               *string  `json:"unit"`
	ManualCurrentValue *float64 `json:"manual_current_value"`
}

func (req targetRequest) toServiceRequest() (service.TargetRequest, error) {
	if req.Name == nil || *req.Name == "" {
		return service.TargetRequest{}, errors.New("name is required")
	}
	if req.Year == nil {
		return service.TargetRequest{}, errors.New("year is required")
	}
	if req.MetricType == nil {
		return service.TargetRequest{}, errors.New("metric_type is required")
	}
	if req.TargetValue == nil {
		return service.TargetRequest{}, errors.New("target_value is required")
	}
	unit := ""
	if req.Unit != nil {
		unit = *req.Unit
	}
	return service.TargetRequest{
		Name:               *req.Name,
		Year:               *req.Year,
		MetricType:         *req.MetricType,
		TargetValue:        *req.TargetValue,
		Unit:               unit,
		ManualCurrentValue: req.ManualCurrentValue,
	}, nil
}

func (h *Handler) listTargets(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.Targets.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) createTarget(w http.ResponseWriter, r *http.Request) {
	var req targetRequest
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
	t, err := h.svc.Targets.Create(r.Context(), userID, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) updateTarget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req targetRequest
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
	t, err := h.svc.Targets.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTarget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Targets.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

type holdingRequest struct {
	CategoryID *int16   `json:"category_id"`
	Name       *string  `json:"name"`
	Gram       *float64 `json:"gram"`
	Qty        *float64 `json:"qty"`
	Brand      *string  `json:"brand"`
	UsdValue   *float64 `json:"usd_value"`
	Currency   *string  `json:"currency"`
	ValueIdr   *float64 `json:"value_idr"`
	Detail     *string  `json:"detail"`
}

func (req holdingRequest) toServiceRequest() (service.HoldingRequest, error) {
	if req.CategoryID == nil {
		return service.HoldingRequest{}, errors.New("category_id is required")
	}
	if req.Name == nil || *req.Name == "" {
		return service.HoldingRequest{}, errors.New("name is required")
	}
	return service.HoldingRequest{
		CategoryID: *req.CategoryID,
		Name:       *req.Name,
		Gram:       req.Gram,
		Qty:        req.Qty,
		Brand:      req.Brand,
		UsdValue:   req.UsdValue,
		Currency:   req.Currency,
		ValueIdr:   req.ValueIdr,
		Detail:     req.Detail,
	}, nil
}

func (h *Handler) createHolding(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req holdingRequest
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
	holding, err := h.svc.Holdings.Create(r.Context(), userID, date, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, holding)
}

func (h *Handler) updateHolding(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid holding id")
		return
	}

	var req holdingRequest
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
	holding, err := h.svc.Holdings.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, holding)
}

func (h *Handler) deleteHolding(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid holding id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.Holdings.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

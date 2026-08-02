package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service"
)

type passiveIncomeRequest struct {
	CategoryID   *int16  `json:"category_id"`
	Name         *string `json:"name"`
	AmountIdr    *int64  `json:"amount_idr"`
	ReceivedDate *string `json:"received_date"`
	IncomeType   *string `json:"income_type"`
	Note         *string `json:"note"`
}

func (req passiveIncomeRequest) toServiceRequest() (service.PassiveIncomeRequest, error) {
	var out service.PassiveIncomeRequest
	if req.CategoryID != nil {
		out.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		out.Name = *req.Name
	}
	if req.AmountIdr != nil {
		out.AmountIdr = *req.AmountIdr
	}
	if req.IncomeType != nil {
		out.IncomeType = *req.IncomeType
	}
	if req.Note != nil {
		out.Note = *req.Note
	}
	if req.ReceivedDate != nil {
		d, err := domain.ParseDate(*req.ReceivedDate)
		if err != nil {
			return service.PassiveIncomeRequest{}, err
		}
		out.ReceivedDate = d
	}
	return out, nil
}

func (h *Handler) listPassiveIncome(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	list, err := h.svc.PassiveIncome.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getPassiveIncomeCalendar(w http.ResponseWriter, r *http.Request) {
	// An absent or unparseable ?year= means "this year" rather than an
	// error — the page loads without a year on first paint.
	year := time.Now().UTC().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			year = parsed
		}
	}

	userID := currentUserID(r.Context())
	calendar, err := h.svc.PassiveIncome.Calendar(r.Context(), userID, year)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calendar)
}

func (h *Handler) createPassiveIncome(w http.ResponseWriter, r *http.Request) {
	var req passiveIncomeRequest
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
	entry, err := h.svc.PassiveIncome.Create(r.Context(), userID, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
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
	svcReq, err := req.toServiceRequest()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := currentUserID(r.Context())
	entry, err := h.svc.PassiveIncome.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *Handler) deletePassiveIncome(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID := currentUserID(r.Context())
	if err := h.svc.PassiveIncome.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

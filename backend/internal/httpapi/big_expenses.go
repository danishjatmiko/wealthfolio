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

type bigExpenseRequest struct {
	Name        *string `json:"name"`
	AmountIdr   *int64  `json:"amount_idr"`
	ExpenseDate *string `json:"expense_date"`
	Category    *string `json:"category"`
}

func (req bigExpenseRequest) toServiceRequest() (service.BigExpenseRequest, error) {
	var out service.BigExpenseRequest
	if req.Name != nil {
		out.Name = *req.Name
	}
	if req.AmountIdr != nil {
		out.AmountIdr = *req.AmountIdr
	}
	if req.Category != nil {
		out.Category = *req.Category
	}
	if req.ExpenseDate != nil {
		d, err := domain.ParseDate(*req.ExpenseDate)
		if err != nil {
			return service.BigExpenseRequest{}, err
		}
		out.ExpenseDate = d
	}
	return out, nil
}

func (h *Handler) listBigExpenses(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	expenses, err := h.svc.BigExpenses.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, expenses)
}

func (h *Handler) getBigExpensesSummary(w http.ResponseWriter, r *http.Request) {
	// An absent or unparseable ?year= means "this year" rather than an
	// error — the page loads without a year on first paint.
	year := time.Now().UTC().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			year = parsed
		}
	}

	userID := currentUserID(r.Context())
	summary, err := h.svc.BigExpenses.Summary(r.Context(), userID, year)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) createBigExpense(w http.ResponseWriter, r *http.Request) {
	var req bigExpenseRequest
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
	expense, err := h.svc.BigExpenses.Create(r.Context(), userID, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, expense)
}

func (h *Handler) updateBigExpense(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid big expense id")
		return
	}

	var req bigExpenseRequest
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
	expense, err := h.svc.BigExpenses.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *Handler) deleteBigExpense(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid big expense id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.BigExpenses.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service"
)

type receivableLoanRequest struct {
	BorrowerName     *string `json:"borrower_name"`
	StartDate        *string `json:"start_date"`
	TermMonths       *int    `json:"term_months"`
	InterestIdr      *int64  `json:"interest_idr"`
	RemainingDebtIdr *int64  `json:"remaining_debt_idr"`
	Note             *string `json:"note"`
}

// toServiceRequest returns an error because start_date arrives as a string
// and can fail to parse — the same reason bondPurchaseRequest.
// toServiceRequest does.
func (req receivableLoanRequest) toServiceRequest() (service.ReceivableLoanRequest, error) {
	var out service.ReceivableLoanRequest
	if req.BorrowerName != nil {
		out.BorrowerName = *req.BorrowerName
	}
	if req.TermMonths != nil {
		out.TermMonths = *req.TermMonths
	}
	if req.InterestIdr != nil {
		out.InterestIdr = *req.InterestIdr
	}
	if req.RemainingDebtIdr != nil {
		out.RemainingDebtIdr = *req.RemainingDebtIdr
	}
	if req.Note != nil {
		out.Note = *req.Note
	}

	if req.StartDate != nil {
		d, err := domain.ParseDate(*req.StartDate)
		if err != nil {
			return service.ReceivableLoanRequest{}, err
		}
		out.StartDate = d
	}
	return out, nil
}

func (h *Handler) listReceivableLoans(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	loans, err := h.svc.ReceivableLoans.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loans)
}

func (h *Handler) createReceivableLoan(w http.ResponseWriter, r *http.Request) {
	var req receivableLoanRequest
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
	loan, err := h.svc.ReceivableLoans.Create(r.Context(), userID, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

func (h *Handler) updateReceivableLoan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid receivable loan id")
		return
	}

	var req receivableLoanRequest
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
	loan, err := h.svc.ReceivableLoans.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (h *Handler) deleteReceivableLoan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid receivable loan id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.ReceivableLoans.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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

type bondPurchaseRequest struct {
	BondName           *string  `json:"bond_name"`
	Platform           *string  `json:"platform"`
	InterestRate       *float64 `json:"interest_rate"`
	BuyDate            *string  `json:"buy_date"`
	MaturityDate       *string  `json:"maturity_date"`
	Quantity           *float64 `json:"quantity"`
	FaceValueUsd       *float64 `json:"face_value_usd"`
	PriceUsd           *float64 `json:"price_usd"`
	AccruedInterestUsd *float64 `json:"accrued_interest_usd"`
	UsdIdrAtPurchase   *float64 `json:"usd_idr_at_purchase"`
}

// toServiceRequest returns an error (unlike most request structs) because
// the two dates arrive as strings and can fail to parse — the same reason
// holdingRequest.toServiceRequest does.
//
// Omitted optional fields fall back to the defaults a broker statement
// implies: one lot of $1,000 face with no accrued interest.
func (req bondPurchaseRequest) toServiceRequest() (service.BondPurchaseRequest, error) {
	out := service.BondPurchaseRequest{
		Quantity:     1,
		FaceValueUsd: 1000,
	}
	if req.BondName != nil {
		out.BondName = *req.BondName
	}
	if req.Platform != nil {
		out.Platform = *req.Platform
	}
	if req.InterestRate != nil {
		out.InterestRate = *req.InterestRate
	}
	if req.Quantity != nil {
		out.Quantity = *req.Quantity
	}
	if req.FaceValueUsd != nil {
		out.FaceValueUsd = *req.FaceValueUsd
	}
	if req.PriceUsd != nil {
		out.PriceUsd = *req.PriceUsd
	}
	if req.AccruedInterestUsd != nil {
		out.AccruedInterestUsd = *req.AccruedInterestUsd
	}
	if req.UsdIdrAtPurchase != nil {
		out.UsdIdrAtPurchase = *req.UsdIdrAtPurchase
	}

	if req.BuyDate != nil {
		d, err := domain.ParseDate(*req.BuyDate)
		if err != nil {
			return service.BondPurchaseRequest{}, err
		}
		out.BuyDate = d
	}
	if req.MaturityDate != nil {
		d, err := domain.ParseDate(*req.MaturityDate)
		if err != nil {
			return service.BondPurchaseRequest{}, err
		}
		out.MaturityDate = d
	}
	return out, nil
}

func (h *Handler) listBondPurchases(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	purchases, err := h.svc.BondPurchases.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, purchases)
}

func (h *Handler) getBondPurchasesSummary(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	summary, err := h.svc.BondPurchases.Summary(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) getCouponCalendar(w http.ResponseWriter, r *http.Request) {
	// An absent or unparseable ?year= means "this year" rather than an
	// error — the page loads without a year on first paint.
	year := time.Now().UTC().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			year = parsed
		}
	}

	userID := currentUserID(r.Context())
	calendar, err := h.svc.BondPurchases.CouponCalendar(r.Context(), userID, year)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calendar)
}

func (h *Handler) createBondPurchase(w http.ResponseWriter, r *http.Request) {
	var req bondPurchaseRequest
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
	purchase, err := h.svc.BondPurchases.Create(r.Context(), userID, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, purchase)
}

func (h *Handler) updateBondPurchase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bond purchase id")
		return
	}

	var req bondPurchaseRequest
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
	purchase, err := h.svc.BondPurchases.Update(r.Context(), userID, id, svcReq)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, purchase)
}

func (h *Handler) deleteBondPurchase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bond purchase id")
		return
	}

	userID := currentUserID(r.Context())
	if err := h.svc.BondPurchases.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

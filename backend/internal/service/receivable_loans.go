package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// ReceivableLoansService owns the loan-terms ledger behind a receivable:
// plain CRUD over receivable_loans, plus the derived figures the Debts and
// Passive Income pages need. Earns a service layer for the same reason
// BondPurchasesService does — PassiveIncomeService reuses its payment
// schedule to merge receivable income into the shared calendar.
type ReceivableLoansService struct {
	repos *db.Repos
}

func NewReceivableLoansService(repos *db.Repos) *ReceivableLoansService {
	return &ReceivableLoansService{repos: repos}
}

// ReceivableLoanDTO is one loan — the stored columns plus everything
// derivable from them, computed at read time so an edit can't leave a
// derived figure stale.
type ReceivableLoanDTO struct {
	ID           uuid.UUID   `json:"id"`
	BorrowerName string      `json:"borrower_name"`
	StartDate    domain.Date `json:"start_date"`
	TermMonths   int         `json:"term_months"`
	InterestIdr  int64       `json:"interest_idr"`
	Note         string      `json:"note"`

	EndDate          domain.Date `json:"end_date"`
	MonthlyAmountIdr int64       `json:"monthly_amount_idr"`
	IsActive         bool        `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func receivableLoanDTO(l domain.ReceivableLoan, today time.Time) ReceivableLoanDTO {
	end := addMonthsAnchored(l.StartDate.Time, l.TermMonths, l.StartDate.Time.Day())
	return ReceivableLoanDTO{
		ID: l.ID, BorrowerName: l.BorrowerName,
		StartDate: l.StartDate, TermMonths: l.TermMonths,
		InterestIdr: l.InterestIdr,
		Note:        l.Note,

		EndDate:          domain.NewDate(end),
		MonthlyAmountIdr: receivableMonthlyAmountIdr(l),
		IsActive:         receivableIsActiveOn(l, today),

		CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt,
	}
}

// List returns every loan for the user, newest first.
func (s *ReceivableLoansService) List(ctx context.Context, userID uuid.UUID) ([]ReceivableLoanDTO, error) {
	loans, err := s.repos.ReceivableLoans.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	today := time.Now().UTC()
	out := make([]ReceivableLoanDTO, 0, len(loans))
	for _, l := range loans {
		out = append(out, receivableLoanDTO(l, today))
	}
	return out, nil
}

// ReceivableLoanRequest is the parsed POST/PUT body for a loan write.
type ReceivableLoanRequest struct {
	BorrowerName string
	StartDate    domain.Date
	TermMonths   int
	InterestIdr  int64
	Note         string
}

func (r ReceivableLoanRequest) validate() error {
	if r.BorrowerName == "" {
		return fmt.Errorf("%w: borrower_name is required", ErrInvalidInput)
	}
	if r.StartDate.Time.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}
	if r.TermMonths <= 0 || r.TermMonths > maxReceivableTermMonths {
		return fmt.Errorf("%w: term_months must be between 1 and %d", ErrInvalidInput, maxReceivableTermMonths)
	}
	if r.InterestIdr <= 0 {
		return fmt.Errorf("%w: interest_idr must be greater than 0", ErrInvalidInput)
	}
	return nil
}

func (r ReceivableLoanRequest) toWrite() db.ReceivableLoanWrite {
	return db.ReceivableLoanWrite{
		BorrowerName: r.BorrowerName, StartDate: r.StartDate,
		TermMonths: r.TermMonths, InterestIdr: r.InterestIdr, Note: r.Note,
	}
}

func (s *ReceivableLoansService) Create(ctx context.Context, userID uuid.UUID, req ReceivableLoanRequest) (ReceivableLoanDTO, error) {
	if err := req.validate(); err != nil {
		return ReceivableLoanDTO{}, err
	}
	l, err := s.repos.ReceivableLoans.Create(ctx, userID, req.toWrite())
	if err != nil {
		return ReceivableLoanDTO{}, err
	}
	return receivableLoanDTO(l, time.Now().UTC()), nil
}

func (s *ReceivableLoansService) Update(ctx context.Context, userID, id uuid.UUID, req ReceivableLoanRequest) (ReceivableLoanDTO, error) {
	if err := req.validate(); err != nil {
		return ReceivableLoanDTO{}, err
	}
	l, err := s.repos.ReceivableLoans.Update(ctx, userID, id, req.toWrite())
	if err != nil {
		return ReceivableLoanDTO{}, err
	}
	return receivableLoanDTO(l, time.Now().UTC()), nil
}

func (s *ReceivableLoansService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.ReceivableLoans.Delete(ctx, userID, id)
}

// receivableEntriesInYear projects every loan's monthly payments landing in
// the given year into IncomeEntryDTOs, one per (borrower, pay date) —
// mirrors BondPurchasesService.couponEntriesInYear, but IDR-native (no
// AmountUsd/rate — a personal loan was never denominated in dollars).
func (s *ReceivableLoansService) receivableEntriesInYear(ctx context.Context, userID uuid.UUID, year int) ([]IncomeEntryDTO, error) {
	loans, err := s.repos.ReceivableLoans.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := []IncomeEntryDTO{}
	for _, l := range loans {
		monthly := receivableMonthlyAmountIdr(l)
		for _, d := range receivablePaymentDatesInYear(l, year) {
			out = append(out, IncomeEntryDTO{
				Kind:      incomeKindReceivable,
				Name:      l.BorrowerName,
				Source:    "Cicilan",
				PayDate:   d,
				AmountIdr: monthly,
			})
		}
	}
	return out, nil
}

// MonthlyIncomePerYearIdr is the user's forward-looking annual receivable
// income — every currently-active loan's monthly amount, annualized.
// Shared by the dashboard and the targets service so the two always agree
// on what "passive income" includes, the same role
// BondPurchasesService.CouponPerYearIdr plays for bond coupons.
func (s *ReceivableLoansService) MonthlyIncomePerYearIdr(ctx context.Context, userID uuid.UUID) (int64, error) {
	loans, err := s.repos.ReceivableLoans.List(ctx, userID)
	if err != nil {
		return 0, err
	}
	today := time.Now().UTC()
	var perYear int64
	for _, l := range loans {
		if !receivableIsActiveOn(l, today) {
			continue
		}
		perYear += receivableMonthlyAmountIdr(l) * 12
	}
	return perYear, nil
}

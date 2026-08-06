package service

import (
	"context"
	"fmt"
	"strings"
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

// DebtEntryDTO is one debt entry as the API returns it: the stored row
// exactly as typed, plus — for a receivable whose loan terms report one —
// the remaining debt alongside it. The two never overwrite each other.
// ValueIdr is always the initial debt, set once and left alone;
// RemainingDebtIdr is the live figure that pages display and total, and is
// 0 when no loan reports one.
type DebtEntryDTO struct {
	domain.DebtEntry
	RemainingDebtIdr int64 `json:"remaining_debt_idr"`
}

// ShownIdr is the figure to display and total for this entry: the remaining
// debt once a loan reports one, otherwise the initial debt as typed.
func (d DebtEntryDTO) ShownIdr() int64 {
	if d.RemainingDebtIdr > 0 {
		return d.RemainingDebtIdr
	}
	return d.ValueIdr
}

// AttachRemainingDebt pairs each owed_to_me entry with its matching loan's
// remaining debt (matched on borrower name, case-insensitively), leaving
// the stored ValueIdr untouched so the initial debt always round-trips
// intact through an edit. i_owe entries and receivables with no matching
// loan simply carry a zero RemainingDebtIdr. Callers are responsible for
// only applying this to the latest snapshot — remaining debt is a live,
// current figure, not something that belongs on locked history.
func (s *ReceivableLoansService) AttachRemainingDebt(ctx context.Context, userID uuid.UUID, entries []domain.DebtEntry) ([]DebtEntryDTO, error) {
	loans, err := s.repos.ReceivableLoans.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	remainingByName := make(map[string]int64, len(loans))
	for _, l := range loans {
		if l.RemainingDebtIdr > 0 {
			remainingByName[strings.ToLower(l.BorrowerName)] = l.RemainingDebtIdr
		}
	}

	out := make([]DebtEntryDTO, 0, len(entries))
	for _, e := range entries {
		dto := DebtEntryDTO{DebtEntry: e}
		if e.Direction == "owed_to_me" {
			dto.RemainingDebtIdr = remainingByName[strings.ToLower(e.Name)]
		}
		out = append(out, dto)
	}
	return out, nil
}

// PlainDebtEntryDTOs wraps entries with no remaining debt attached — used
// for locked historical snapshots, which show exactly what was recorded.
func PlainDebtEntryDTOs(entries []domain.DebtEntry) []DebtEntryDTO {
	out := make([]DebtEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, DebtEntryDTO{DebtEntry: e})
	}
	return out
}

// ReceivableLoanDTO is one loan — the stored columns plus everything
// derivable from them, computed at read time so an edit can't leave a
// derived figure stale.
type ReceivableLoanDTO struct {
	ID               uuid.UUID   `json:"id"`
	BorrowerName     string      `json:"borrower_name"`
	StartDate        domain.Date `json:"start_date"`
	TermMonths       int         `json:"term_months"`
	InterestIdr      int64       `json:"interest_idr"`
	RemainingDebtIdr int64       `json:"remaining_debt_idr"`
	Note             string      `json:"note"`

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
		InterestIdr:      l.InterestIdr,
		RemainingDebtIdr: l.RemainingDebtIdr,
		Note:             l.Note,

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
	BorrowerName     string
	StartDate        domain.Date
	TermMonths       int
	InterestIdr      int64
	RemainingDebtIdr int64
	Note             string
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
	if r.RemainingDebtIdr < 0 {
		return fmt.Errorf("%w: remaining_debt_idr cannot be negative", ErrInvalidInput)
	}
	return nil
}

func (r ReceivableLoanRequest) toWrite() db.ReceivableLoanWrite {
	return db.ReceivableLoanWrite{
		BorrowerName: r.BorrowerName, StartDate: r.StartDate,
		TermMonths: r.TermMonths, InterestIdr: r.InterestIdr,
		RemainingDebtIdr: r.RemainingDebtIdr, Note: r.Note,
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

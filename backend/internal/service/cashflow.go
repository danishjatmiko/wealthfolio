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

type CashflowService struct {
	repos   *db.Repos
	passive *PassiveIncomeService
}

func NewCashflowService(repos *db.Repos, passive *PassiveIncomeService) *CashflowService {
	return &CashflowService{repos: repos, passive: passive}
}

type CashflowMonthDTO struct {
	Month                     int                     `json:"month"`
	Label                     string                  `json:"label"`
	PeriodStartDate           domain.Date             `json:"period_start_date"`
	PeriodEndDate             domain.Date             `json:"period_end_date"`
	ActiveIncome              *db.MonthlyActiveIncome `json:"active_income"`
	PassiveIncomeIdr          int64                   `json:"passive_income_idr"`
	ProjectedPassiveIncomeIdr int64                   `json:"projected_passive_income_idr"`
	ExpensesIdr               int64                   `json:"expenses_idr"`
	HasExpensePeriod          bool                    `json:"has_expense_period"`
	BigExpensesIdr            int64                   `json:"big_expenses_idr"`
	TotalIncomeIdr            int64                   `json:"total_income_idr"`
	TotalExpensesIdr          int64                   `json:"total_expenses_idr"`
	NetIdr                    int64                   `json:"net_idr"`
}

type CashflowDTO struct {
	ReferenceYear int                `json:"reference_year"`
	Months        []CashflowMonthDTO `json:"months"`
}

func validateCashflowYear(year int) error {
	if year < 2026 || year > 9998 {
		return fmt.Errorf("%w: year must be between 2026 and 9998", ErrInvalidInput)
	}
	return nil
}

// ParseCashflowMonth accepts only an exact YYYY-MM month from August 2026.
func ParseCashflowMonth(raw string) (domain.Date, error) {
	parsed, err := time.Parse("2006-01", raw)
	if err != nil || parsed.Year() < 2026 || parsed.Year() > 9998 || (parsed.Year() == 2026 && parsed.Month() < time.August) {
		return domain.Date{}, fmt.Errorf("%w: month must be YYYY-MM, from 2026-08 onward (through 9998-12)", ErrInvalidInput)
	}
	return domain.NewDate(parsed), nil
}

type ActiveIncomeRequest struct {
	AmountIdr *int64 `json:"amount_idr"`
	Note      string `json:"note"`
}

func (r ActiveIncomeRequest) validate() error {
	if r.AmountIdr == nil || *r.AmountIdr < 0 || *r.AmountIdr > 9007199254740991 {
		return fmt.Errorf("%w: amount_idr must be a non-negative whole Rupiah amount (at most 9007199254740991)", ErrInvalidInput)
	}
	if len(r.Note) > 2000 {
		return fmt.Errorf("%w: note must be at most 2000 bytes", ErrInvalidInput)
	}
	return nil
}

func (s *CashflowService) SaveActiveIncome(ctx context.Context, userID uuid.UUID, rawMonth string, req ActiveIncomeRequest) (db.MonthlyActiveIncome, error) {
	month, err := ParseCashflowMonth(rawMonth)
	if err != nil {
		return db.MonthlyActiveIncome{}, err
	}
	if err := req.validate(); err != nil {
		return db.MonthlyActiveIncome{}, err
	}
	return s.repos.Cashflow.SaveActiveIncome(ctx, userID, month, *req.AmountIdr, strings.TrimSpace(req.Note))
}

func (s *CashflowService) DeleteActiveIncome(ctx context.Context, userID uuid.UUID, rawMonth string) error {
	month, err := ParseCashflowMonth(rawMonth)
	if err != nil {
		return err
	}
	return s.repos.Cashflow.DeleteActiveIncome(ctx, userID, month)
}

func (s *CashflowService) Summary(ctx context.Context, userID uuid.UUID, year int) (CashflowDTO, error) {
	if err := validateCashflowYear(year); err != nil {
		return CashflowDTO{}, err
	}
	passive, err := s.passive.Calendar(ctx, userID, year)
	if err != nil {
		return CashflowDTO{}, err
	}
	periods, err := s.repos.ExpensePeriods.ListWithAgg(ctx, userID)
	if err != nil {
		return CashflowDTO{}, err
	}
	bigExpenses, err := s.repos.BigExpenses.List(ctx, userID)
	if err != nil {
		return CashflowDTO{}, err
	}
	active, err := s.repos.Cashflow.ListActiveIncome(ctx, userID, year)
	if err != nil {
		return CashflowDTO{}, err
	}
	return aggregateCashflow(year, active, passive.Months, periods, bigExpenses), nil
}

func aggregateCashflow(year int, active []db.MonthlyActiveIncome, passive []IncomeMonthDTO, periods []db.ExpensePeriodAgg, bigExpenses []domain.BigExpense) CashflowDTO {
	firstMonth := 1
	if year == 2026 {
		firstMonth = 8
	}
	out := CashflowDTO{ReferenceYear: year, Months: make([]CashflowMonthDTO, 13-firstMonth)}
	for i := range out.Months {
		month := firstMonth + i
		start, end := boundsForPeriodMonth(year, time.Month(month))
		out.Months[i] = CashflowMonthDTO{
			Month: month, Label: time.Month(month).String(),
			PeriodStartDate: start, PeriodEndDate: end,
		}
	}
	byMonth := func(month int) *CashflowMonthDTO {
		if month < firstMonth || month > 12 {
			return nil
		}
		return &out.Months[month-firstMonth]
	}
	for _, entry := range active {
		if entry.Month.Year() != year {
			continue
		}
		if m := byMonth(int(entry.Month.Month())); m != nil {
			entry := entry
			m.ActiveIncome = &entry
			m.TotalIncomeIdr += entry.AmountIdr
		}
	}
	for _, p := range passive {
		if m := byMonth(p.Month); m != nil {
			m.PassiveIncomeIdr += p.AmountIdr
			m.ProjectedPassiveIncomeIdr += p.CouponIdr + p.ReceivableIdr
			m.TotalIncomeIdr += p.AmountIdr
		}
	}
	// Expenses follow their existing period label (end month), not created_at.
	// This also counts backfilled historical expenses in their intended month.
	for _, p := range periods {
		if p.Period.EndDate.Year() != year {
			continue
		}
		if m := byMonth(int(p.Period.EndDate.Month())); m != nil {
			if !m.HasExpensePeriod {
				m.PeriodStartDate = p.Period.StartDate
				m.PeriodEndDate = p.Period.EndDate
			} else {
				if p.Period.StartDate.Before(m.PeriodStartDate.Time) {
					m.PeriodStartDate = p.Period.StartDate
				}
				if p.Period.EndDate.After(m.PeriodEndDate.Time) {
					m.PeriodEndDate = p.Period.EndDate
				}
			}
			m.HasExpensePeriod = true
			m.ExpensesIdr += p.ActualTotalIdr
		}
	}
	for _, e := range bigExpenses {
		if e.ExpenseDate.Year() != year {
			continue
		}
		if m := byMonth(int(e.ExpenseDate.Month())); m != nil {
			m.BigExpensesIdr += e.AmountIdr
		}
	}
	for i := range out.Months {
		m := &out.Months[i]
		m.TotalExpensesIdr = m.ExpensesIdr + m.BigExpensesIdr
		m.NetIdr = m.TotalIncomeIdr - m.TotalExpensesIdr
	}
	return out
}

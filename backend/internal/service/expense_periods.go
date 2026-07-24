package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// ExpensePeriodsService implements the 25th-of-month pay-cycle bounds
// computation and period creation. Periods never lock — every period
// stays fully editable indefinitely, unlike Snapshots/DebtSnapshots.
type ExpensePeriodsService struct {
	repos *db.Repos
}

func NewExpensePeriodsService(repos *db.Repos) *ExpensePeriodsService {
	return &ExpensePeriodsService{repos: repos}
}

// ExpensePeriodSummary is the shape returned by GET /expense-periods.
type ExpensePeriodSummary struct {
	ID                uuid.UUID   `json:"id"`
	StartDate         domain.Date `json:"start_date"`
	EndDate           domain.Date `json:"end_date"`
	Label             string      `json:"label"`
	ActualTotalIdr    int64       `json:"actual_total_idr"`
	CommittedTotalIdr int64       `json:"committed_total_idr"`
}

// BudgetEnvelopeDetail is one envelope plus its fixed expenses (the full
// domain.FixedExpense rows, same shape POST/PUT /fixed-expenses return —
// same nest-the-domain-struct pattern as SnapshotDetail.Holdings) and
// computed actual total, within period detail.
type BudgetEnvelopeDetail struct {
	ID                 uuid.UUID             `json:"id"`
	Name               string                `json:"name"`
	CommittedAmountIdr int64                 `json:"committed_amount_idr"`
	ActualTotalIdr     int64                 `json:"actual_total_idr"`
	FixedExpenses      []domain.FixedExpense `json:"fixed_expenses"`
}

// ExpensePeriodDetail is the shape returned by GET /expense-periods/latest,
// GET /expense-periods/{id}, and POST /expense-periods.
type ExpensePeriodDetail struct {
	ID                uuid.UUID              `json:"id"`
	StartDate         domain.Date            `json:"start_date"`
	EndDate           domain.Date            `json:"end_date"`
	Label             string                 `json:"label"`
	Envelopes         []BudgetEnvelopeDetail `json:"envelopes"`
	ActualTotalIdr    int64                  `json:"actual_total_idr"`
	CommittedTotalIdr int64                  `json:"committed_total_idr"`
}

// periodLabel formats a period's display name from its end date's month +
// year (e.g. end_date=2026-08-24 -> "August 2026") — the period is named
// after the month it ends in, per the 25th-of-month convention.
func periodLabel(end domain.Date) string {
	return end.Time.Format("January 2006")
}

// boundsForPeriodMonth derives the 25th-to-24th pay-cycle window for a
// period named after the given year/month (e.g. year=2026, month=August
// -> 25 Jul 2026 to 24 Aug 2026) — the period is named after, and ends in,
// the given month, per periodLabel's convention.
func boundsForPeriodMonth(year int, month time.Month) (start, end domain.Date) {
	startYear, startMonth := year, month-1
	if startMonth < time.January {
		startMonth = time.December
		startYear--
	}
	s := time.Date(startYear, startMonth, 25, 0, 0, 0, 0, time.UTC)
	e := time.Date(year, month, 24, 0, 0, 0, 0, time.UTC)
	return domain.NewDate(s), domain.NewDate(e)
}

func (s *ExpensePeriodsService) detailFromPeriod(ctx context.Context, period domain.ExpensePeriod) (ExpensePeriodDetail, error) {
	envelopes, err := s.repos.BudgetEnvelopes.ListByPeriod(ctx, period.ID)
	if err != nil {
		return ExpensePeriodDetail{}, err
	}
	expenses, err := s.repos.FixedExpenses.ListByPeriod(ctx, period.ID)
	if err != nil {
		return ExpensePeriodDetail{}, err
	}

	byEnvelope := map[uuid.UUID][]domain.FixedExpense{}
	var actualTotal int64
	for _, e := range expenses {
		actualTotal += e.AmountIdr
		byEnvelope[e.EnvelopeID] = append(byEnvelope[e.EnvelopeID], e)
	}

	envelopeDetails := make([]BudgetEnvelopeDetail, 0, len(envelopes))
	var committedTotal int64
	for _, env := range envelopes {
		children := byEnvelope[env.ID]
		if children == nil {
			children = []domain.FixedExpense{}
		}
		var envActual int64
		for _, c := range children {
			envActual += c.AmountIdr
		}
		committedTotal += env.CommittedAmountIdr
		envelopeDetails = append(envelopeDetails, BudgetEnvelopeDetail{
			ID:                 env.ID,
			Name:               env.Name,
			CommittedAmountIdr: env.CommittedAmountIdr,
			ActualTotalIdr:     envActual,
			FixedExpenses:      children,
		})
	}

	return ExpensePeriodDetail{
		ID:                period.ID,
		StartDate:         period.StartDate,
		EndDate:           period.EndDate,
		Label:             periodLabel(period.EndDate),
		Envelopes:         envelopeDetails,
		ActualTotalIdr:    actualTotal,
		CommittedTotalIdr: committedTotal,
	}, nil
}

// ListSummaries returns every period for the user, newest first.
func (s *ExpensePeriodsService) ListSummaries(ctx context.Context, userID uuid.UUID) ([]ExpensePeriodSummary, error) {
	aggs, err := s.repos.ExpensePeriods.ListWithAgg(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ExpensePeriodSummary, 0, len(aggs))
	for _, a := range aggs {
		out = append(out, ExpensePeriodSummary{
			ID:                a.Period.ID,
			StartDate:         a.Period.StartDate,
			EndDate:           a.Period.EndDate,
			Label:             periodLabel(a.Period.EndDate),
			ActualTotalIdr:    a.ActualTotalIdr,
			CommittedTotalIdr: a.CommittedTotalIdr,
		})
	}
	return out, nil
}

// GetLatestDetail returns the user's most recent period with full detail.
// Returns db.ErrNotFound if the user has no periods yet.
func (s *ExpensePeriodsService) GetLatestDetail(ctx context.Context, userID uuid.UUID) (ExpensePeriodDetail, error) {
	period, err := s.repos.ExpensePeriods.GetLatest(ctx, userID)
	if err != nil {
		return ExpensePeriodDetail{}, err
	}
	return s.detailFromPeriod(ctx, period)
}

// GetDetail returns a specific period with full detail. Returns
// db.ErrNotFound if it doesn't exist or isn't owned by userID.
func (s *ExpensePeriodsService) GetDetail(ctx context.Context, userID, periodID uuid.UUID) (ExpensePeriodDetail, error) {
	period, err := s.repos.ExpensePeriods.GetByID(ctx, userID, periodID)
	if err != nil {
		return ExpensePeriodDetail{}, err
	}
	return s.detailFromPeriod(ctx, period)
}

// Create makes a new period named after the given year/month (see
// boundsForPeriodMonth) and returns its detail. Unlike Snapshots/
// DebtSnapshots' today-or-later rule, any month may be picked — the
// frontend currently offers Jan 2026 through Dec 2028 — since periods
// never lock and backfilling past months is a normal use case here.
// Returns ErrPeriodMonthExists if a period already exists for that month.
// If copyEnvelopes, every envelope from the user's current latest period
// (name + committed target only, not its fixed expenses) is cloned into
// the new one, regardless of whether the new period is chronologically
// before or after that latest one.
func (s *ExpensePeriodsService) Create(ctx context.Context, userID uuid.UUID, year int, month time.Month, copyEnvelopes bool) (ExpensePeriodDetail, error) {
	latest, err := s.repos.ExpensePeriods.GetLatest(ctx, userID)
	hasLatest := true
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			hasLatest = false
		} else {
			return ExpensePeriodDetail{}, err
		}
	}

	start, end := boundsForPeriodMonth(year, month)

	if _, err := s.repos.ExpensePeriods.GetByStartDate(ctx, userID, start); err == nil {
		return ExpensePeriodDetail{}, ErrPeriodMonthExists
	} else if !errors.Is(err, db.ErrNotFound) {
		return ExpensePeriodDetail{}, err
	}

	newPeriod, err := s.repos.ExpensePeriods.Create(ctx, userID, start, end)
	if err != nil {
		return ExpensePeriodDetail{}, err
	}

	if copyEnvelopes && hasLatest {
		if err := s.repos.BudgetEnvelopes.CopyFromPeriod(ctx, latest.ID, newPeriod.ID); err != nil {
			return ExpensePeriodDetail{}, err
		}
	}

	return s.detailFromPeriod(ctx, newPeriod)
}

// Delete removes a period and everything inside it. Any period can be
// deleted, not just the latest — periods never lock, so there's no
// mutability rule to enforce here. Returns db.ErrNotFound if it doesn't
// exist or isn't owned by userID.
func (s *ExpensePeriodsService) Delete(ctx context.Context, userID, periodID uuid.UUID) error {
	return s.repos.ExpensePeriods.Delete(ctx, userID, periodID)
}

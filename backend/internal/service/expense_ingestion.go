package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/service/notificationparse"
)

// ExpenseIngestionRequest is a single notification capture forwarded by
// the Android app — raw fields only, no parsing done on-device.
type ExpenseIngestionRequest struct {
	IdempotencyKey string
	Source         string
	RawTitle       *string
	RawText        *string
	RawBigText     *string
	OccurredAt     time.Time
}

// ExpenseIngestionResult is the shape returned by POST /expense-ingestions.
// Status is always "created" or "ignored" — every outcome besides a
// genuine transient failure comes back as a 200 with one of these, since
// both are the *expected*, idempotent-on-replay result of a valid request.
type ExpenseIngestionResult struct {
	Status         string     `json:"status"`
	FixedExpenseID *uuid.UUID `json:"fixed_expense_id,omitempty"`
	EnvelopeID     *uuid.UUID `json:"envelope_id,omitempty"`
	AmountIdr      *int64     `json:"amount_idr,omitempty"`
	MerchantName   *string    `json:"merchant_name,omitempty"`
}

// ExpenseIngestionService turns captured notifications into Fixed
// Expenses, idempotently.
type ExpenseIngestionService struct {
	repos *db.Repos
}

func NewExpenseIngestionService(repos *db.Repos) *ExpenseIngestionService {
	return &ExpenseIngestionService{repos: repos}
}

// periodMonthForDate returns the year/month of the pay-cycle period that
// covers t, following the same 25th-of-month naming rule as
// boundsForPeriodMonth: on or after the 25th, t belongs to next month's
// period; before the 25th, it belongs to the current month's.
func periodMonthForDate(t time.Time) (int, time.Month) {
	year, month, day := t.Date()
	if day >= 25 {
		month++
		if month > time.December {
			month = time.January
			year++
		}
	}
	return year, month
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Ingest resolves a captured notification into a Fixed Expense:
//  1. parse the raw text — no match means "ignored", a terminal outcome
//     the caller should stop retrying, not an error.
//  2. resolve the period covering OccurredAt — missing period, source
//     mapping, or envelope are all retryable failures (ErrNoActivePeriod /
//     ErrNoSourceMapping / ErrEnvelopeNotFound) rather than "ignored",
//     since they resolve themselves once the user fixes their setup.
//  3. idempotently create the fixed_expense + audit event.
func (s *ExpenseIngestionService) Ingest(ctx context.Context, userID uuid.UUID, req ExpenseIngestionRequest) (ExpenseIngestionResult, error) {
	raw := db.RawFields{
		UserID:         userID,
		IdempotencyKey: req.IdempotencyKey,
		Source:         req.Source,
		RawTitle:       req.RawTitle,
		RawText:        req.RawText,
		RawBigText:     req.RawBigText,
		OccurredAt:     req.OccurredAt,
	}

	parsed, ok := notificationparse.Parse(req.Source, deref(req.RawTitle), deref(req.RawText), deref(req.RawBigText))
	if !ok {
		if _, err := s.repos.NotificationExpenseEvents.CreateIgnored(ctx, raw); err != nil {
			return ExpenseIngestionResult{}, err
		}
		return ExpenseIngestionResult{Status: "ignored"}, nil
	}

	year, month := periodMonthForDate(req.OccurredAt)
	start, _ := boundsForPeriodMonth(year, month)
	period, err := s.repos.ExpensePeriods.GetByStartDate(ctx, userID, start)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ExpenseIngestionResult{}, ErrNoActivePeriod
		}
		return ExpenseIngestionResult{}, err
	}

	mapping, err := s.repos.ExpenseSourceMappings.GetBySource(ctx, userID, req.Source)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ExpenseIngestionResult{}, ErrNoSourceMapping
		}
		return ExpenseIngestionResult{}, err
	}

	envelopes, err := s.repos.BudgetEnvelopes.ListByPeriod(ctx, period.ID)
	if err != nil {
		return ExpenseIngestionResult{}, err
	}
	var envelopeID uuid.UUID
	found := false
	for _, env := range envelopes {
		if env.Name == mapping.EnvelopeName {
			envelopeID = env.ID
			found = true
			break
		}
	}
	if !found {
		return ExpenseIngestionResult{}, ErrEnvelopeNotFound
	}

	event, err := s.repos.NotificationExpenseEvents.CreateExpense(ctx, db.CreateExpenseWrite{
		RawFields:  raw,
		PeriodID:   period.ID,
		EnvelopeID: envelopeID,
		AmountIdr:  parsed.AmountIdr,
		Merchant:   parsed.Merchant,
	})
	if err != nil {
		return ExpenseIngestionResult{}, err
	}

	return ExpenseIngestionResult{
		Status:         "created",
		FixedExpenseID: event.FixedExpenseID,
		EnvelopeID:     event.EnvelopeID,
		AmountIdr:      event.AmountIdr,
		MerchantName:   event.MerchantName,
	}, nil
}

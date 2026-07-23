package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// FixedExpensesService implements plain CRUD for fixed expenses. Periods
// never lock, so there's no latest-snapshot check anywhere in this
// service.
type FixedExpensesService struct {
	repos *db.Repos
}

func NewFixedExpensesService(repos *db.Repos) *FixedExpensesService {
	return &FixedExpensesService{repos: repos}
}

// FixedExpenseRequest is the parsed POST/PUT body for a fixed expense
// write. Every fixed expense belongs to a budget envelope — there's no
// more "standalone" expense.
type FixedExpenseRequest struct {
	Name       string
	AmountIdr  int64
	EnvelopeID uuid.UUID
}

func (r FixedExpenseRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if r.EnvelopeID == uuid.Nil {
		return fmt.Errorf("%w: envelope_id is required", ErrInvalidInput)
	}
	return nil
}

// checkEnvelopeBelongsToPeriod returns ErrInvalidInput if the envelope
// isn't owned by userID or belongs to a different period than periodID.
func (s *FixedExpensesService) checkEnvelopeBelongsToPeriod(ctx context.Context, userID, periodID, envelopeID uuid.UUID) error {
	env, err := s.repos.BudgetEnvelopes.GetByID(ctx, userID, envelopeID)
	if err != nil {
		return err
	}
	if env.PeriodID != periodID {
		return fmt.Errorf("%w: envelope belongs to a different period", ErrInvalidInput)
	}
	return nil
}

// Create adds a new fixed expense to the given period. Returns
// db.ErrNotFound if the period doesn't exist or isn't owned by userID.
func (s *FixedExpensesService) Create(ctx context.Context, userID, periodID uuid.UUID, req FixedExpenseRequest) (domain.FixedExpense, error) {
	if err := req.validate(); err != nil {
		return domain.FixedExpense{}, err
	}
	if _, err := s.repos.ExpensePeriods.GetByID(ctx, userID, periodID); err != nil {
		return domain.FixedExpense{}, err
	}
	if err := s.checkEnvelopeBelongsToPeriod(ctx, userID, periodID, req.EnvelopeID); err != nil {
		return domain.FixedExpense{}, err
	}
	return s.repos.FixedExpenses.Create(ctx, db.FixedExpenseWrite{
		PeriodID:   periodID,
		EnvelopeID: req.EnvelopeID,
		Name:       req.Name,
		AmountIdr:  req.AmountIdr,
	})
}

// Update overwrites an existing fixed expense. Returns db.ErrNotFound if
// it doesn't exist or isn't owned by userID.
func (s *FixedExpensesService) Update(ctx context.Context, userID, id uuid.UUID, req FixedExpenseRequest) (domain.FixedExpense, error) {
	if err := req.validate(); err != nil {
		return domain.FixedExpense{}, err
	}
	existing, err := s.repos.FixedExpenses.GetByID(ctx, userID, id)
	if err != nil {
		return domain.FixedExpense{}, err
	}
	if err := s.checkEnvelopeBelongsToPeriod(ctx, userID, existing.PeriodID, req.EnvelopeID); err != nil {
		return domain.FixedExpense{}, err
	}
	return s.repos.FixedExpenses.Update(ctx, userID, id, req.EnvelopeID, req.Name, req.AmountIdr)
}

// Delete removes a fixed expense. Returns db.ErrNotFound if it doesn't
// exist or isn't owned by userID.
func (s *FixedExpensesService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.FixedExpenses.Delete(ctx, userID, id)
}

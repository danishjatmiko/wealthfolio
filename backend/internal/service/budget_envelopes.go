package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// BudgetEnvelopesService implements plain CRUD for budget envelopes.
// Periods never lock, so — unlike Holdings/DebtEntries — there's no
// latest-snapshot check anywhere in this service.
type BudgetEnvelopesService struct {
	repos *db.Repos
}

func NewBudgetEnvelopesService(repos *db.Repos) *BudgetEnvelopesService {
	return &BudgetEnvelopesService{repos: repos}
}

// BudgetEnvelopeRequest is the parsed POST/PUT body for an envelope write.
type BudgetEnvelopeRequest struct {
	Name               string
	CommittedAmountIdr int64
}

func (r BudgetEnvelopeRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return nil
}

// Create adds a new envelope to the given period. Returns db.ErrNotFound
// if the period doesn't exist or isn't owned by userID.
func (s *BudgetEnvelopesService) Create(ctx context.Context, userID, periodID uuid.UUID, req BudgetEnvelopeRequest) (domain.BudgetEnvelope, error) {
	if err := req.validate(); err != nil {
		return domain.BudgetEnvelope{}, err
	}
	if _, err := s.repos.ExpensePeriods.GetByID(ctx, userID, periodID); err != nil {
		return domain.BudgetEnvelope{}, err
	}
	return s.repos.BudgetEnvelopes.Create(ctx, db.BudgetEnvelopeWrite{
		PeriodID:           periodID,
		Name:               req.Name,
		CommittedAmountIdr: req.CommittedAmountIdr,
	})
}

// Update overwrites an existing envelope. Returns db.ErrNotFound if it
// doesn't exist or isn't owned by userID.
func (s *BudgetEnvelopesService) Update(ctx context.Context, userID, id uuid.UUID, req BudgetEnvelopeRequest) (domain.BudgetEnvelope, error) {
	if err := req.validate(); err != nil {
		return domain.BudgetEnvelope{}, err
	}
	return s.repos.BudgetEnvelopes.Update(ctx, userID, id, req.Name, req.CommittedAmountIdr)
}

// Delete removes an envelope and every fixed expense inside it (cascade).
// Returns db.ErrNotFound if it doesn't exist or isn't owned by userID.
func (s *BudgetEnvelopesService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.BudgetEnvelopes.Delete(ctx, userID, id)
}

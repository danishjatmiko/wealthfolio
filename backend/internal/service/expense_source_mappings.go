package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// ExpenseSourceMappingsService implements plain CRUD for the per-source
// (GoPay/DANA/BCA) default envelope mapping the Android app's Settings
// screen configures.
type ExpenseSourceMappingsService struct {
	repos *db.Repos
}

func NewExpenseSourceMappingsService(repos *db.Repos) *ExpenseSourceMappingsService {
	return &ExpenseSourceMappingsService{repos: repos}
}

// List returns every source mapping the user has configured.
func (s *ExpenseSourceMappingsService) List(ctx context.Context, userID uuid.UUID) ([]domain.ExpenseSourceMapping, error) {
	return s.repos.ExpenseSourceMappings.ListByUser(ctx, userID)
}

// Upsert sets which envelope a source's captured expenses auto-file into.
// A source is valid to map as long as it exists in the notification_apps
// catalog — even if currently disabled, since a disabled entry can still
// legitimately have an old mapping.
func (s *ExpenseSourceMappingsService) Upsert(ctx context.Context, userID uuid.UUID, source, envelopeName string) (domain.ExpenseSourceMapping, error) {
	known, err := s.repos.NotificationApps.ExistsSource(ctx, source)
	if err != nil {
		return domain.ExpenseSourceMapping{}, err
	}
	if !known {
		return domain.ExpenseSourceMapping{}, fmt.Errorf("%w: unknown source %q", ErrInvalidInput, source)
	}
	if envelopeName == "" {
		return domain.ExpenseSourceMapping{}, fmt.Errorf("%w: envelope_name is required", ErrInvalidInput)
	}
	return s.repos.ExpenseSourceMappings.Upsert(ctx, userID, source, envelopeName)
}

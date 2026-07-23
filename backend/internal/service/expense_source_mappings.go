package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service/notificationparse"
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

func isKnownSource(source string) bool {
	switch source {
	case notificationparse.SourceGoPay, notificationparse.SourceDANA, notificationparse.SourceBCA:
		return true
	default:
		return false
	}
}

// List returns every source mapping the user has configured.
func (s *ExpenseSourceMappingsService) List(ctx context.Context, userID uuid.UUID) ([]domain.ExpenseSourceMapping, error) {
	return s.repos.ExpenseSourceMappings.ListByUser(ctx, userID)
}

// Upsert sets which envelope a source's captured expenses auto-file into.
func (s *ExpenseSourceMappingsService) Upsert(ctx context.Context, userID uuid.UUID, source, envelopeName string) (domain.ExpenseSourceMapping, error) {
	if !isKnownSource(source) {
		return domain.ExpenseSourceMapping{}, fmt.Errorf("%w: unknown source %q", ErrInvalidInput, source)
	}
	if envelopeName == "" {
		return domain.ExpenseSourceMapping{}, fmt.Errorf("%w: envelope_name is required", ErrInvalidInput)
	}
	return s.repos.ExpenseSourceMappings.Upsert(ctx, userID, source, envelopeName)
}

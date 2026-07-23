package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// ExpenseCategoriesService implements plain CRUD for expense categories —
// a free-form, user-created grouping for budget envelopes.
type ExpenseCategoriesService struct {
	repos *db.Repos
}

func NewExpenseCategoriesService(repos *db.Repos) *ExpenseCategoriesService {
	return &ExpenseCategoriesService{repos: repos}
}

// List returns every category for the user, oldest-created first.
func (s *ExpenseCategoriesService) List(ctx context.Context, userID uuid.UUID) ([]domain.ExpenseCategory, error) {
	return s.repos.ExpenseCategories.ListByUser(ctx, userID)
}

// Create adds a new category for the user. Returns db.ErrDuplicateName if
// the user already has a category with this exact name.
func (s *ExpenseCategoriesService) Create(ctx context.Context, userID uuid.UUID, name string) (domain.ExpenseCategory, error) {
	if name == "" {
		return domain.ExpenseCategory{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return s.repos.ExpenseCategories.Create(ctx, userID, name)
}

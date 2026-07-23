package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// ExpenseCategoriesRepo manages expense_categories rows.
type ExpenseCategoriesRepo struct {
	pool *pgxpool.Pool
}

func NewExpenseCategoriesRepo(pool *pgxpool.Pool) *ExpenseCategoriesRepo {
	return &ExpenseCategoriesRepo{pool: pool}
}

const expenseCategorySelectCols = `id, user_id, name, created_at`

func scanExpenseCategory(row interface{ Scan(dest ...any) error }) (domain.ExpenseCategory, error) {
	var c domain.ExpenseCategory
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt)
	if err != nil {
		return domain.ExpenseCategory{}, err
	}
	return c, nil
}

// ListByUser returns every category for the user, oldest-created first
// (the order new categories are assigned a chart color by).
func (r *ExpenseCategoriesRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ExpenseCategory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+expenseCategorySelectCols+`
		FROM expense_categories WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ExpenseCategory{}
	for rows.Next() {
		c, err := scanExpenseCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetByID returns a single category owned by userID. ErrNotFound if
// missing or owned by someone else.
func (r *ExpenseCategoriesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.ExpenseCategory, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+expenseCategorySelectCols+`
		FROM expense_categories WHERE id = $1 AND user_id = $2`, id, userID)
	c, err := scanExpenseCategory(row)
	if err != nil {
		return domain.ExpenseCategory{}, wrapNotFound(err)
	}
	return c, nil
}

// Create inserts a new category for the user. ErrDuplicateName if the user
// already has a category with this exact name (unique per user).
func (r *ExpenseCategoriesRepo) Create(ctx context.Context, userID uuid.UUID, name string) (domain.ExpenseCategory, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO expense_categories (user_id, name)
		VALUES ($1, $2)
		RETURNING `+expenseCategorySelectCols,
		userID, name)
	c, err := scanExpenseCategory(row)
	if err != nil {
		return domain.ExpenseCategory{}, wrapUniqueViolation(err, ErrDuplicateName)
	}
	return c, nil
}

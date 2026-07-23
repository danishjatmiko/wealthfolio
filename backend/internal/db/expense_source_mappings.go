package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// ExpenseSourceMappingsRepo manages expense_source_mappings rows.
type ExpenseSourceMappingsRepo struct {
	pool *pgxpool.Pool
}

func NewExpenseSourceMappingsRepo(pool *pgxpool.Pool) *ExpenseSourceMappingsRepo {
	return &ExpenseSourceMappingsRepo{pool: pool}
}

const expenseSourceMappingSelectCols = `id, user_id, source, envelope_name, updated_at`

func scanExpenseSourceMapping(row interface{ Scan(dest ...any) error }) (domain.ExpenseSourceMapping, error) {
	var m domain.ExpenseSourceMapping
	err := row.Scan(&m.ID, &m.UserID, &m.Source, &m.EnvelopeName, &m.UpdatedAt)
	if err != nil {
		return domain.ExpenseSourceMapping{}, err
	}
	return m, nil
}

// ListByUser returns every source mapping the user has configured.
func (r *ExpenseSourceMappingsRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ExpenseSourceMapping, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+expenseSourceMappingSelectCols+`
		FROM expense_source_mappings WHERE user_id = $1 ORDER BY source`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ExpenseSourceMapping{}
	for rows.Next() {
		m, err := scanExpenseSourceMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetBySource returns the user's mapping for one source. ErrNotFound if
// they haven't configured it yet.
func (r *ExpenseSourceMappingsRepo) GetBySource(ctx context.Context, userID uuid.UUID, source string) (domain.ExpenseSourceMapping, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+expenseSourceMappingSelectCols+`
		FROM expense_source_mappings WHERE user_id = $1 AND source = $2`, userID, source)
	m, err := scanExpenseSourceMapping(row)
	if err != nil {
		return domain.ExpenseSourceMapping{}, wrapNotFound(err)
	}
	return m, nil
}

// Upsert sets the envelope name a source maps to, creating the mapping if
// it doesn't exist yet or overwriting it if it does (one mapping per
// user+source).
func (r *ExpenseSourceMappingsRepo) Upsert(ctx context.Context, userID uuid.UUID, source, envelopeName string) (domain.ExpenseSourceMapping, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO expense_source_mappings (user_id, source, envelope_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, source) DO UPDATE SET envelope_name = $3, updated_at = now()
		RETURNING `+expenseSourceMappingSelectCols,
		userID, source, envelopeName)
	return scanExpenseSourceMapping(row)
}

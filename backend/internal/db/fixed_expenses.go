package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// FixedExpensesRepo manages fixed_expenses rows.
type FixedExpensesRepo struct {
	pool *pgxpool.Pool
}

func NewFixedExpensesRepo(pool *pgxpool.Pool) *FixedExpensesRepo {
	return &FixedExpensesRepo{pool: pool}
}

const fixedExpenseSelectCols = `id, period_id, envelope_id, name, amount_idr, created_at, updated_at`

func scanFixedExpense(row interface{ Scan(dest ...any) error }) (domain.FixedExpense, error) {
	var e domain.FixedExpense
	err := row.Scan(&e.ID, &e.PeriodID, &e.EnvelopeID, &e.Name, &e.AmountIdr, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.FixedExpense{}, err
	}
	return e, nil
}

// ListByPeriod returns every fixed expense in a period (standalone and
// envelope-linked alike), oldest-created first.
func (r *FixedExpensesRepo) ListByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.FixedExpense, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+fixedExpenseSelectCols+`
		FROM fixed_expenses WHERE period_id = $1 ORDER BY created_at`, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.FixedExpense{}
	for rows.Next() {
		e, err := scanFixedExpense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns a single fixed expense owned by userID (via its
// period). ErrNotFound if missing or owned by someone else.
func (r *FixedExpensesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.FixedExpense, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT e.id, e.period_id, e.envelope_id, e.name, e.amount_idr, e.created_at, e.updated_at
		FROM fixed_expenses e
		JOIN expense_periods p ON p.id = e.period_id
		WHERE e.id = $1 AND p.user_id = $2`, id, userID)
	e, err := scanFixedExpense(row)
	if err != nil {
		return domain.FixedExpense{}, wrapNotFound(err)
	}
	return e, nil
}

// FixedExpenseWrite is the set of columns needed to insert/update a fixed
// expense.
type FixedExpenseWrite struct {
	PeriodID   uuid.UUID
	EnvelopeID uuid.UUID
	Name       string
	AmountIdr  int64
}

// Create inserts a new fixed expense and returns the full row.
func (r *FixedExpensesRepo) Create(ctx context.Context, w FixedExpenseWrite) (domain.FixedExpense, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO fixed_expenses (period_id, envelope_id, name, amount_idr)
		VALUES ($1, $2, $3, $4)
		RETURNING `+fixedExpenseSelectCols,
		w.PeriodID, w.EnvelopeID, w.Name, w.AmountIdr)
	return scanFixedExpense(row)
}

// Update overwrites an existing fixed expense's mutable fields.
// ErrNotFound if the id doesn't exist or isn't owned by userID (via its
// period).
func (r *FixedExpensesRepo) Update(ctx context.Context, userID, id uuid.UUID, envelopeID uuid.UUID, name string, amountIdr int64) (domain.FixedExpense, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE fixed_expenses
		SET envelope_id = $1, name = $2, amount_idr = $3, updated_at = now()
		WHERE id = $4 AND period_id IN (SELECT id FROM expense_periods WHERE user_id = $5)
		RETURNING `+fixedExpenseSelectCols,
		envelopeID, name, amountIdr, id, userID)
	e, err := scanFixedExpense(row)
	if err != nil {
		return domain.FixedExpense{}, wrapNotFound(err)
	}
	return e, nil
}

// Delete removes a fixed expense by id. ErrNotFound if it didn't exist or
// isn't owned by userID (via its period).
func (r *FixedExpensesRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM fixed_expenses
		WHERE id = $1 AND period_id IN (SELECT id FROM expense_periods WHERE user_id = $2)`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

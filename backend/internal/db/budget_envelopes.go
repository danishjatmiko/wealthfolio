package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// BudgetEnvelopesRepo manages budget_envelopes rows.
type BudgetEnvelopesRepo struct {
	pool *pgxpool.Pool
}

func NewBudgetEnvelopesRepo(pool *pgxpool.Pool) *BudgetEnvelopesRepo {
	return &BudgetEnvelopesRepo{pool: pool}
}

// budgetEnvelopeJoinCols selects a budget_envelopes row ("e") joined with
// its category ("ec") for the category_name convenience field — same
// pattern as Holding.CategoryKey/CategoryLabel.
const budgetEnvelopeJoinCols = `e.id, e.period_id, e.category_id, ec.name, e.name, e.committed_amount_idr, e.created_at, e.updated_at`

func scanBudgetEnvelope(row interface{ Scan(dest ...any) error }) (domain.BudgetEnvelope, error) {
	var e domain.BudgetEnvelope
	err := row.Scan(&e.ID, &e.PeriodID, &e.CategoryID, &e.CategoryName, &e.Name, &e.CommittedAmountIdr, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.BudgetEnvelope{}, err
	}
	return e, nil
}

// ListByPeriod returns every envelope in a period, oldest-created first.
func (r *BudgetEnvelopesRepo) ListByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.BudgetEnvelope, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+budgetEnvelopeJoinCols+`
		FROM budget_envelopes e
		JOIN expense_categories ec ON ec.id = e.category_id
		WHERE e.period_id = $1 ORDER BY e.created_at`, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.BudgetEnvelope{}
	for rows.Next() {
		e, err := scanBudgetEnvelope(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns a single envelope owned by userID (via its period).
// ErrNotFound if missing or owned by someone else.
func (r *BudgetEnvelopesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.BudgetEnvelope, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+budgetEnvelopeJoinCols+`
		FROM budget_envelopes e
		JOIN expense_categories ec ON ec.id = e.category_id
		JOIN expense_periods p ON p.id = e.period_id
		WHERE e.id = $1 AND p.user_id = $2`, id, userID)
	e, err := scanBudgetEnvelope(row)
	if err != nil {
		return domain.BudgetEnvelope{}, wrapNotFound(err)
	}
	return e, nil
}

// BudgetEnvelopeWrite is the set of columns needed to insert/update an
// envelope.
type BudgetEnvelopeWrite struct {
	PeriodID           uuid.UUID
	CategoryID         uuid.UUID
	Name               string
	CommittedAmountIdr int64
}

// Create inserts a new envelope and returns the full row (with its
// category joined in).
func (r *BudgetEnvelopesRepo) Create(ctx context.Context, w BudgetEnvelopeWrite) (domain.BudgetEnvelope, error) {
	row := r.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO budget_envelopes (period_id, category_id, name, committed_amount_idr)
			VALUES ($1, $2, $3, $4)
			RETURNING id, period_id, category_id, name, committed_amount_idr, created_at, updated_at
		)
		SELECT i.id, i.period_id, i.category_id, ec.name, i.name, i.committed_amount_idr, i.created_at, i.updated_at
		FROM inserted i JOIN expense_categories ec ON ec.id = i.category_id`,
		w.PeriodID, w.CategoryID, w.Name, w.CommittedAmountIdr)
	return scanBudgetEnvelope(row)
}

// Update overwrites an existing envelope's mutable fields. ErrNotFound if
// the id doesn't exist or isn't owned by userID (via its period).
func (r *BudgetEnvelopesRepo) Update(ctx context.Context, userID, id uuid.UUID, categoryID uuid.UUID, name string, committedAmountIdr int64) (domain.BudgetEnvelope, error) {
	row := r.pool.QueryRow(ctx, `
		WITH updated AS (
			UPDATE budget_envelopes
			SET category_id = $1, name = $2, committed_amount_idr = $3, updated_at = now()
			WHERE id = $4 AND period_id IN (SELECT id FROM expense_periods WHERE user_id = $5)
			RETURNING id, period_id, category_id, name, committed_amount_idr, created_at, updated_at
		)
		SELECT u.id, u.period_id, u.category_id, ec.name, u.name, u.committed_amount_idr, u.created_at, u.updated_at
		FROM updated u JOIN expense_categories ec ON ec.id = u.category_id`,
		categoryID, name, committedAmountIdr, id, userID)
	e, err := scanBudgetEnvelope(row)
	if err != nil {
		return domain.BudgetEnvelope{}, wrapNotFound(err)
	}
	return e, nil
}

// Delete removes an envelope by id (cascades to its fixed_expenses).
// ErrNotFound if it didn't exist or isn't owned by userID (via its period).
func (r *BudgetEnvelopesRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM budget_envelopes
		WHERE id = $1 AND period_id IN (SELECT id FROM expense_periods WHERE user_id = $2)`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CopyFromPeriod duplicates every envelope of fromPeriodID into
// toPeriodID (category + name + committed target, fresh ids/timestamps,
// zero fixed expenses) — used when creating a new period with envelopes
// copied forward.
func (r *BudgetEnvelopesRepo) CopyFromPeriod(ctx context.Context, fromPeriodID, toPeriodID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO budget_envelopes (period_id, category_id, name, committed_amount_idr)
		SELECT $2, category_id, name, committed_amount_idr
		FROM budget_envelopes
		WHERE period_id = $1`, fromPeriodID, toPeriodID)
	return err
}

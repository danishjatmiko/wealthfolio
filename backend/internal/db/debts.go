package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// MaxUpdatedAt returns the most recent updated_at across the user's debts,
// or nil if they have none yet.
func (r *DebtsRepo) MaxUpdatedAt(ctx context.Context, userID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM debts WHERE user_id = $1`, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

// DebtsRepo manages debts rows.
type DebtsRepo struct {
	pool *pgxpool.Pool
}

func NewDebtsRepo(pool *pgxpool.Pool) *DebtsRepo {
	return &DebtsRepo{pool: pool}
}

func scanDebt(row interface{ Scan(dest ...any) error }) (domain.Debt, error) {
	var d domain.Debt
	err := row.Scan(&d.ID, &d.UserID, &d.Name, &d.Type, &d.ValueIdr, &d.Direction)
	if err != nil {
		return domain.Debt{}, err
	}
	return d, nil
}

// List returns every debt for the user.
func (r *DebtsRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, type, value_idr, direction
		FROM debts WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Debt{}
	for rows.Next() {
		d, err := scanDebt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetByID returns a single debt. ErrNotFound if missing.
func (r *DebtsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Debt, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, type, value_idr, direction
		FROM debts WHERE id = $1`, id)
	d, err := scanDebt(row)
	if err != nil {
		return domain.Debt{}, wrapNotFound(err)
	}
	return d, nil
}

// Create inserts a new debt.
func (r *DebtsRepo) Create(ctx context.Context, userID uuid.UUID, name, debtType string, valueIdr int64, direction string) (domain.Debt, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO debts (user_id, name, type, value_idr, direction)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, type, value_idr, direction`,
		userID, name, debtType, valueIdr, direction)
	return scanDebt(row)
}

// Update overwrites a debt's fields. ErrNotFound if the id doesn't exist.
func (r *DebtsRepo) Update(ctx context.Context, id uuid.UUID, name, debtType string, valueIdr int64, direction string) (domain.Debt, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE debts SET name = $1, type = $2, value_idr = $3, direction = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, user_id, name, type, value_idr, direction`,
		name, debtType, valueIdr, direction, id)
	d, err := scanDebt(row)
	if err != nil {
		return domain.Debt{}, wrapNotFound(err)
	}
	return d, nil
}

// Delete removes a debt by id. ErrNotFound if it didn't exist.
func (r *DebtsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM debts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SumByDirection returns the total debt value where direction = 'i_owe'
// and where direction = 'owed_to_me', for the given user.
func (r *DebtsRepo) SumByDirection(ctx context.Context, userID uuid.UUID) (iOwe int64, owedToMe int64, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(value_idr) FILTER (WHERE direction = 'i_owe'), 0),
			COALESCE(SUM(value_idr) FILTER (WHERE direction = 'owed_to_me'), 0)
		FROM debts WHERE user_id = $1`, userID).Scan(&iOwe, &owedToMe)
	return iOwe, owedToMe, err
}

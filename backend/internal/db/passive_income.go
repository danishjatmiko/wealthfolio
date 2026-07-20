package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// MaxUpdatedAt returns the most recent updated_at across the user's passive
// income sources, or nil if they have none yet.
func (r *PassiveIncomeRepo) MaxUpdatedAt(ctx context.Context, userID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM passive_income_sources WHERE user_id = $1`, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

// PassiveIncomeRepo manages passive_income_sources rows, joined with their
// category for the category_key/category_label convenience fields.
type PassiveIncomeRepo struct {
	pool *pgxpool.Pool
}

func NewPassiveIncomeRepo(pool *pgxpool.Pool) *PassiveIncomeRepo {
	return &PassiveIncomeRepo{pool: pool}
}

func scanPassiveIncome(row interface{ Scan(dest ...any) error }) (domain.PassiveIncomeSource, error) {
	var p domain.PassiveIncomeSource
	err := row.Scan(&p.ID, &p.UserID, &p.CategoryID, &p.CategoryKey, &p.CategoryLabel, &p.Name, &p.PerYearIdr)
	if err != nil {
		return domain.PassiveIncomeSource{}, err
	}
	return p, nil
}

const passiveIncomeSelect = `
	SELECT p.id, p.user_id, p.category_id, c.key, c.label, p.name, p.per_year_idr
	FROM passive_income_sources p
	JOIN categories c ON c.id = p.category_id`

// List returns every passive income source for the user.
func (r *PassiveIncomeRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.PassiveIncomeSource, error) {
	rows, err := r.pool.Query(ctx, passiveIncomeSelect+` WHERE p.user_id = $1 ORDER BY p.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.PassiveIncomeSource{}
	for rows.Next() {
		p, err := scanPassiveIncome(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetByID returns a single passive income source. ErrNotFound if missing.
func (r *PassiveIncomeRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.PassiveIncomeSource, error) {
	row := r.pool.QueryRow(ctx, passiveIncomeSelect+` WHERE p.id = $1`, id)
	p, err := scanPassiveIncome(row)
	if err != nil {
		return domain.PassiveIncomeSource{}, wrapNotFound(err)
	}
	return p, nil
}

// Create inserts a new passive income source.
func (r *PassiveIncomeRepo) Create(ctx context.Context, userID uuid.UUID, categoryID int16, name string, perYearIdr int64) (domain.PassiveIncomeSource, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO passive_income_sources (user_id, category_id, name, per_year_idr)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, userID, categoryID, name, perYearIdr).Scan(&id)
	if err != nil {
		return domain.PassiveIncomeSource{}, err
	}
	return r.GetByID(ctx, id)
}

// Update overwrites a passive income source's fields. ErrNotFound if the id
// doesn't exist.
func (r *PassiveIncomeRepo) Update(ctx context.Context, id uuid.UUID, categoryID int16, name string, perYearIdr int64) (domain.PassiveIncomeSource, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE passive_income_sources
		SET category_id = $1, name = $2, per_year_idr = $3, updated_at = now()
		WHERE id = $4`, categoryID, name, perYearIdr, id)
	if err != nil {
		return domain.PassiveIncomeSource{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.PassiveIncomeSource{}, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// Delete removes a passive income source by id. ErrNotFound if it didn't
// exist.
func (r *PassiveIncomeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM passive_income_sources WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Sum returns the total per_year_idr across all of the user's passive
// income sources.
func (r *PassiveIncomeRepo) Sum(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(per_year_idr), 0) FROM passive_income_sources WHERE user_id = $1`, userID).Scan(&total)
	return total, err
}

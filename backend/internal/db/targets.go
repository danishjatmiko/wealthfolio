package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// TargetsRepo manages targets rows. The computed current_value/percent/
// lower_is_better fields on domain.Target are NOT populated by this repo;
// see internal/service/targets.go.
type TargetsRepo struct {
	pool *pgxpool.Pool
}

func NewTargetsRepo(pool *pgxpool.Pool) *TargetsRepo {
	return &TargetsRepo{pool: pool}
}

func scanTarget(row interface{ Scan(dest ...any) error }) (domain.Target, error) {
	var t domain.Target
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Year, &t.MetricType, &t.TargetValue, &t.Unit, &t.ManualCurrentValue)
	if err != nil {
		return domain.Target{}, err
	}
	return t, nil
}

const targetSelect = `
	SELECT id, user_id, name, year, metric_type, target_value, unit, manual_current_value
	FROM targets`

// List returns every target for the user.
func (r *TargetsRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Target, error) {
	rows, err := r.pool.Query(ctx, targetSelect+` WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Target{}
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetByID returns a single target. ErrNotFound if missing.
func (r *TargetsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Target, error) {
	row := r.pool.QueryRow(ctx, targetSelect+` WHERE id = $1`, id)
	t, err := scanTarget(row)
	if err != nil {
		return domain.Target{}, wrapNotFound(err)
	}
	return t, nil
}

// Create inserts a new target.
func (r *TargetsRepo) Create(ctx context.Context, userID uuid.UUID, name string, year int, metricType string, targetValue float64, unit string, manualCurrentValue *float64) (domain.Target, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO targets (user_id, name, year, metric_type, target_value, unit, manual_current_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, year, metric_type, target_value, unit, manual_current_value`,
		userID, name, year, metricType, targetValue, unit, manualCurrentValue)
	return scanTarget(row)
}

// Update overwrites a target's fields. ErrNotFound if the id doesn't exist.
func (r *TargetsRepo) Update(ctx context.Context, id uuid.UUID, name string, year int, metricType string, targetValue float64, unit string, manualCurrentValue *float64) (domain.Target, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE targets
		SET name = $1, year = $2, metric_type = $3, target_value = $4, unit = $5, manual_current_value = $6, updated_at = now()
		WHERE id = $7
		RETURNING id, user_id, name, year, metric_type, target_value, unit, manual_current_value`,
		name, year, metricType, targetValue, unit, manualCurrentValue, id)
	t, err := scanTarget(row)
	if err != nil {
		return domain.Target{}, wrapNotFound(err)
	}
	return t, nil
}

// Delete removes a target by id. ErrNotFound if it didn't exist.
func (r *TargetsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM targets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FirstTargetValueByMetricType returns the target_value of the first
// (oldest-created) target row with the given metric_type, and whether one
// exists at all.
func (r *TargetsRepo) FirstTargetValueByMetricType(ctx context.Context, userID uuid.UUID, metricType string) (float64, bool, error) {
	var v float64
	err := r.pool.QueryRow(ctx, `
		SELECT target_value FROM targets
		WHERE user_id = $1 AND metric_type = $2
		ORDER BY created_at
		LIMIT 1`, userID, metricType).Scan(&v)
	if err != nil {
		if errors.Is(wrapNotFound(err), ErrNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return v, true, nil
}

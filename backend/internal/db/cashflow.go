package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"wealthfolio/backend/internal/domain"
)

type CashflowRepo struct{ pool *pgxpool.Pool }

func NewCashflowRepo(pool *pgxpool.Pool) *CashflowRepo { return &CashflowRepo{pool: pool} }

type MonthlyActiveIncome struct {
	Month     domain.Date `json:"month"`
	AmountIdr int64       `json:"amount_idr"`
	Note      string      `json:"note"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (r *CashflowRepo) ListActiveIncome(ctx context.Context, userID uuid.UUID, year int) ([]MonthlyActiveIncome, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	rows, err := r.pool.Query(ctx, `SELECT month, amount_idr, note, updated_at
        FROM monthly_active_income WHERE user_id = $1 AND month >= $2 AND month < $3 ORDER BY month`,
		userID, start, start.AddDate(1, 0, 0))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MonthlyActiveIncome{}
	for rows.Next() {
		var entry MonthlyActiveIncome
		if err := rows.Scan(&entry.Month, &entry.AmountIdr, &entry.Note, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// One total per user/month; repeated saves replace the total, never duplicate it.
func (r *CashflowRepo) SaveActiveIncome(ctx context.Context, userID uuid.UUID, month domain.Date, amount int64, note string) (MonthlyActiveIncome, error) {
	var entry MonthlyActiveIncome
	err := r.pool.QueryRow(ctx, `INSERT INTO monthly_active_income (user_id, month, amount_idr, note)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, month) DO UPDATE
        SET amount_idr = EXCLUDED.amount_idr, note = EXCLUDED.note, updated_at = now()
        RETURNING month, amount_idr, note, updated_at`, userID, month.Time, amount, note).
		Scan(&entry.Month, &entry.AmountIdr, &entry.Note, &entry.UpdatedAt)
	return entry, err
}

func (r *CashflowRepo) DeleteActiveIncome(ctx context.Context, userID uuid.UUID, month domain.Date) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM monthly_active_income WHERE user_id = $1 AND month = $2`, userID, month.Time)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

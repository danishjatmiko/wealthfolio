package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// RatesRepo manages rate_entries rows.
type RatesRepo struct {
	pool *pgxpool.Pool
}

func NewRatesRepo(pool *pgxpool.Pool) *RatesRepo {
	return &RatesRepo{pool: pool}
}

func scanRateEntry(row interface {
	Scan(dest ...any) error
}) (domain.RateEntry, error) {
	var (
		re        domain.RateEntry
		entryDate time.Time
	)
	err := row.Scan(&re.ID, &re.UserID, &entryDate, &re.Antam, &re.Kinghalim, &re.Ubs, &re.UsdIdr, &re.CreatedAt)
	if err != nil {
		return domain.RateEntry{}, err
	}
	re.EntryDate = domain.NewDate(entryDate)
	return re, nil
}

// List returns all rate entries for the user, newest entry_date first.
func (r *RatesRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.RateEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, entry_date, antam, kinghalim, ubs, usd_idr, created_at
		FROM rate_entries
		WHERE user_id = $1
		ORDER BY entry_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.RateEntry{}
	for rows.Next() {
		re, err := scanRateEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, re)
	}
	return out, rows.Err()
}

// GetLatest returns the most recent rate entry for the user (by
// entry_date), or ErrNotFound if the user has none yet.
func (r *RatesRepo) GetLatest(ctx context.Context, userID uuid.UUID) (domain.RateEntry, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, entry_date, antam, kinghalim, ubs, usd_idr, created_at
		FROM rate_entries
		WHERE user_id = $1
		ORDER BY entry_date DESC
		LIMIT 1`, userID)
	re, err := scanRateEntry(row)
	if err != nil {
		return domain.RateEntry{}, wrapNotFound(err)
	}
	return re, nil
}

// Upsert inserts a new rate entry or, if one already exists for
// (user_id, entry_date), updates it in place. Returns the resulting row.
func (r *RatesRepo) Upsert(ctx context.Context, userID uuid.UUID, entryDate domain.Date, antam, kinghalim, ubs, usdIdr float64) (domain.RateEntry, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO rate_entries (user_id, entry_date, antam, kinghalim, ubs, usd_idr)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, entry_date) DO UPDATE
			SET antam = excluded.antam,
				kinghalim = excluded.kinghalim,
				ubs = excluded.ubs,
				usd_idr = excluded.usd_idr
		RETURNING id, user_id, entry_date, antam, kinghalim, ubs, usd_idr, created_at`,
		userID, entryDate.Time, antam, kinghalim, ubs, usdIdr)
	re, err := scanRateEntry(row)
	if err != nil {
		return domain.RateEntry{}, err
	}
	return re, nil
}

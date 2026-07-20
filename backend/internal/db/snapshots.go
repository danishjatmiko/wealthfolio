package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// SnapshotsRepo manages snapshots rows.
type SnapshotsRepo struct {
	pool *pgxpool.Pool
}

func NewSnapshotsRepo(pool *pgxpool.Pool) *SnapshotsRepo {
	return &SnapshotsRepo{pool: pool}
}

// SnapshotAgg is a snapshot joined with aggregate figures over its holdings.
type SnapshotAgg struct {
	Snapshot      domain.Snapshot
	HoldingsCount int64
	NetEquityIdr  int64
}

func scanSnapshot(row interface{ Scan(dest ...any) error }) (domain.Snapshot, error) {
	var (
		s  domain.Snapshot
		sd time.Time
	)
	if err := row.Scan(&s.ID, &s.UserID, &sd, &s.CreatedAt); err != nil {
		return domain.Snapshot{}, err
	}
	s.SnapshotDate = domain.NewDate(sd)
	return s, nil
}

// ListWithAgg returns every snapshot for the user, newest snapshot_date
// first, along with each snapshot's holdings count and net equity
// (sum of asset holdings minus sum of liability holdings).
func (r *SnapshotsRepo) ListWithAgg(ctx context.Context, userID uuid.UUID) ([]SnapshotAgg, error) {
	return r.listWithAgg(ctx, userID, "DESC")
}

// ListWithAggAsc is ListWithAgg ordered oldest-first (used by the progress
// series computation).
func (r *SnapshotsRepo) ListWithAggAsc(ctx context.Context, userID uuid.UUID) ([]SnapshotAgg, error) {
	return r.listWithAgg(ctx, userID, "ASC")
}

func (r *SnapshotsRepo) listWithAgg(ctx context.Context, userID uuid.UUID, order string) ([]SnapshotAgg, error) {
	dir := "DESC"
	if order == "ASC" {
		dir = "ASC"
	}
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.user_id, s.snapshot_date, s.created_at,
			COUNT(h.id) AS holdings_count,
			COALESCE(SUM(h.value_idr), 0) AS net_equity_idr
		FROM snapshots s
		LEFT JOIN holdings h ON h.snapshot_id = s.id
		WHERE s.user_id = $1
		GROUP BY s.id
		ORDER BY s.snapshot_date `+dir, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SnapshotAgg{}
	for rows.Next() {
		var (
			agg SnapshotAgg
			sd  time.Time
		)
		if err := rows.Scan(&agg.Snapshot.ID, &agg.Snapshot.UserID, &sd, &agg.Snapshot.CreatedAt, &agg.HoldingsCount, &agg.NetEquityIdr); err != nil {
			return nil, err
		}
		agg.Snapshot.SnapshotDate = domain.NewDate(sd)
		out = append(out, agg)
	}
	return out, rows.Err()
}

// GetByID returns a single snapshot by id. ErrNotFound if missing.
func (r *SnapshotsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Snapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM snapshots WHERE id = $1`, id)
	s, err := scanSnapshot(row)
	if err != nil {
		return domain.Snapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// GetByDate returns the user's snapshot for the given date. ErrNotFound if
// there isn't one.
func (r *SnapshotsRepo) GetByDate(ctx context.Context, userID uuid.UUID, date domain.Date) (domain.Snapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM snapshots WHERE user_id = $1 AND snapshot_date = $2`, userID, date.Time)
	s, err := scanSnapshot(row)
	if err != nil {
		return domain.Snapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// GetLatest returns the user's snapshot with the maximum snapshot_date.
// ErrNotFound if the user has no snapshots yet.
func (r *SnapshotsRepo) GetLatest(ctx context.Context, userID uuid.UUID) (domain.Snapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM snapshots
		WHERE user_id = $1
		ORDER BY snapshot_date DESC
		LIMIT 1`, userID)
	s, err := scanSnapshot(row)
	if err != nil {
		return domain.Snapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// Create inserts a new snapshot for the user on the given date.
func (r *SnapshotsRepo) Create(ctx context.Context, userID uuid.UUID, date domain.Date) (domain.Snapshot, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO snapshots (user_id, snapshot_date)
		VALUES ($1, $2)
		RETURNING id, user_id, snapshot_date, created_at`, userID, date.Time)
	return scanSnapshot(row)
}

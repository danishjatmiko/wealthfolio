package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// DebtSnapshotsRepo manages debt_snapshots rows.
type DebtSnapshotsRepo struct {
	pool *pgxpool.Pool
}

func NewDebtSnapshotsRepo(pool *pgxpool.Pool) *DebtSnapshotsRepo {
	return &DebtSnapshotsRepo{pool: pool}
}

// DebtSnapshotAgg is a debt snapshot joined with aggregate figures over its
// entries.
type DebtSnapshotAgg struct {
	Snapshot     domain.DebtSnapshot
	EntriesCount int64
	IOweIdr      int64
	OwedToMeIdr  int64
}

func scanDebtSnapshot(row interface{ Scan(dest ...any) error }) (domain.DebtSnapshot, error) {
	var (
		s  domain.DebtSnapshot
		sd time.Time
	)
	if err := row.Scan(&s.ID, &s.UserID, &sd, &s.CreatedAt); err != nil {
		return domain.DebtSnapshot{}, err
	}
	s.SnapshotDate = domain.NewDate(sd)
	return s, nil
}

// ListWithAgg returns every debt snapshot for the user, newest
// snapshot_date first, along with each snapshot's entry count and totals
// by direction.
func (r *DebtSnapshotsRepo) ListWithAgg(ctx context.Context, userID uuid.UUID) ([]DebtSnapshotAgg, error) {
	return r.listWithAgg(ctx, userID, "DESC")
}

// ListWithAggAsc is ListWithAgg ordered oldest-first (used by the debt
// progress series computation).
func (r *DebtSnapshotsRepo) ListWithAggAsc(ctx context.Context, userID uuid.UUID) ([]DebtSnapshotAgg, error) {
	return r.listWithAgg(ctx, userID, "ASC")
}

func (r *DebtSnapshotsRepo) listWithAgg(ctx context.Context, userID uuid.UUID, order string) ([]DebtSnapshotAgg, error) {
	dir := "DESC"
	if order == "ASC" {
		dir = "ASC"
	}
	rows, err := r.pool.Query(ctx, `
		SELECT ds.id, ds.user_id, ds.snapshot_date, ds.created_at,
			COUNT(de.id) AS entries_count,
			COALESCE(SUM(de.value_idr) FILTER (WHERE de.direction = 'i_owe'), 0) AS i_owe_idr,
			COALESCE(SUM(de.value_idr) FILTER (WHERE de.direction = 'owed_to_me'), 0) AS owed_to_me_idr
		FROM debt_snapshots ds
		LEFT JOIN debt_entries de ON de.debt_snapshot_id = ds.id
		WHERE ds.user_id = $1 AND ds.deleted_at IS NULL
		GROUP BY ds.id
		ORDER BY ds.snapshot_date `+dir, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []DebtSnapshotAgg{}
	for rows.Next() {
		var (
			agg DebtSnapshotAgg
			sd  time.Time
		)
		if err := rows.Scan(&agg.Snapshot.ID, &agg.Snapshot.UserID, &sd, &agg.Snapshot.CreatedAt, &agg.EntriesCount, &agg.IOweIdr, &agg.OwedToMeIdr); err != nil {
			return nil, err
		}
		agg.Snapshot.SnapshotDate = domain.NewDate(sd)
		out = append(out, agg)
	}
	return out, rows.Err()
}

// GetByID returns a single non-deleted debt snapshot by id. ErrNotFound if
// missing or soft-deleted.
func (r *DebtSnapshotsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.DebtSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM debt_snapshots WHERE id = $1 AND deleted_at IS NULL`, id)
	s, err := scanDebtSnapshot(row)
	if err != nil {
		return domain.DebtSnapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// GetByDate returns the user's non-deleted debt snapshot for the given
// date. ErrNotFound if there isn't one.
func (r *DebtSnapshotsRepo) GetByDate(ctx context.Context, userID uuid.UUID, date domain.Date) (domain.DebtSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM debt_snapshots WHERE user_id = $1 AND snapshot_date = $2 AND deleted_at IS NULL`, userID, date.Time)
	s, err := scanDebtSnapshot(row)
	if err != nil {
		return domain.DebtSnapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// GetLatest returns the user's non-deleted debt snapshot with the maximum
// snapshot_date. ErrNotFound if the user has no debt snapshots yet.
func (r *DebtSnapshotsRepo) GetLatest(ctx context.Context, userID uuid.UUID) (domain.DebtSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, snapshot_date, created_at
		FROM debt_snapshots
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY snapshot_date DESC
		LIMIT 1`, userID)
	s, err := scanDebtSnapshot(row)
	if err != nil {
		return domain.DebtSnapshot{}, wrapNotFound(err)
	}
	return s, nil
}

// Create inserts a new debt snapshot for the user on the given date.
func (r *DebtSnapshotsRepo) Create(ctx context.Context, userID uuid.UUID, date domain.Date) (domain.DebtSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO debt_snapshots (user_id, snapshot_date)
		VALUES ($1, $2)
		RETURNING id, user_id, snapshot_date, created_at`, userID, date.Time)
	return scanDebtSnapshot(row)
}

// Delete soft-deletes a debt snapshot owned by userID by stamping
// deleted_at. ErrNotFound if it doesn't exist, isn't owned by userID, or is
// already deleted.
func (r *DebtSnapshotsRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE debt_snapshots SET deleted_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// DebtEntriesRepo manages debt_entries rows.
type DebtEntriesRepo struct {
	pool *pgxpool.Pool
}

func NewDebtEntriesRepo(pool *pgxpool.Pool) *DebtEntriesRepo {
	return &DebtEntriesRepo{pool: pool}
}

const debtEntrySelectCols = `id, debt_snapshot_id, name, type, value_idr, direction, created_at, updated_at`

func scanDebtEntry(row interface{ Scan(dest ...any) error }) (domain.DebtEntry, error) {
	var e domain.DebtEntry
	err := row.Scan(&e.ID, &e.DebtSnapshotID, &e.Name, &e.Type, &e.ValueIdr, &e.Direction, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.DebtEntry{}, err
	}
	return e, nil
}

// ListByDebtSnapshot returns every entry in a debt snapshot, oldest-created
// first.
func (r *DebtEntriesRepo) ListByDebtSnapshot(ctx context.Context, debtSnapshotID uuid.UUID) ([]domain.DebtEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+debtEntrySelectCols+`
		FROM debt_entries WHERE debt_snapshot_id = $1 ORDER BY created_at`, debtSnapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.DebtEntry{}
	for rows.Next() {
		e, err := scanDebtEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns a single debt entry (with its debt_snapshot_id) owned by
// userID (via its debt snapshot). ErrNotFound if missing or owned by
// someone else.
func (r *DebtEntriesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.DebtEntry, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT e.id, e.debt_snapshot_id, e.name, e.type, e.value_idr, e.direction, e.created_at, e.updated_at
		FROM debt_entries e
		JOIN debt_snapshots ds ON ds.id = e.debt_snapshot_id
		WHERE e.id = $1 AND ds.user_id = $2`, id, userID)
	e, err := scanDebtEntry(row)
	if err != nil {
		return domain.DebtEntry{}, wrapNotFound(err)
	}
	return e, nil
}

// DebtEntryWrite is the set of columns needed to insert/update a debt entry.
type DebtEntryWrite struct {
	DebtSnapshotID uuid.UUID
	Name           string
	Type           string
	ValueIdr       int64
	Direction      string
}

// Create inserts a new debt entry and returns the full row.
func (r *DebtEntriesRepo) Create(ctx context.Context, w DebtEntryWrite) (domain.DebtEntry, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO debt_entries (debt_snapshot_id, name, type, value_idr, direction)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+debtEntrySelectCols,
		w.DebtSnapshotID, w.Name, w.Type, w.ValueIdr, w.Direction)
	return scanDebtEntry(row)
}

// Update overwrites an existing debt entry's mutable fields and returns the
// refreshed row. ErrNotFound if the id doesn't exist or isn't owned by
// userID (via its debt snapshot).
func (r *DebtEntriesRepo) Update(ctx context.Context, userID, id uuid.UUID, w DebtEntryWrite) (domain.DebtEntry, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE debt_entries
		SET name = $1, type = $2, value_idr = $3, direction = $4, updated_at = now()
		WHERE id = $5 AND debt_snapshot_id IN (SELECT id FROM debt_snapshots WHERE user_id = $6)
		RETURNING `+debtEntrySelectCols,
		w.Name, w.Type, w.ValueIdr, w.Direction, id, userID)
	e, err := scanDebtEntry(row)
	if err != nil {
		return domain.DebtEntry{}, wrapNotFound(err)
	}
	return e, nil
}

// Delete removes a debt entry by id. ErrNotFound if it didn't exist or
// isn't owned by userID (via its debt snapshot).
func (r *DebtEntriesRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM debt_entries
		WHERE id = $1 AND debt_snapshot_id IN (SELECT id FROM debt_snapshots WHERE user_id = $2)`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CopyFromSnapshot duplicates every entry of fromDebtSnapshotID into
// toDebtSnapshotID, generating fresh ids/timestamps.
func (r *DebtEntriesRepo) CopyFromSnapshot(ctx context.Context, fromDebtSnapshotID, toDebtSnapshotID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO debt_entries (debt_snapshot_id, name, type, value_idr, direction)
		SELECT $2, name, type, value_idr, direction
		FROM debt_entries
		WHERE debt_snapshot_id = $1`, fromDebtSnapshotID, toDebtSnapshotID)
	return err
}

// MaxUpdatedAt returns the most recent updated_at across the given debt
// snapshot's entries, or nil if it has none.
func (r *DebtEntriesRepo) MaxUpdatedAt(ctx context.Context, debtSnapshotID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM debt_entries WHERE debt_snapshot_id = $1`, debtSnapshotID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

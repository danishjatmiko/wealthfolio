package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// HoldingsRepo manages holdings rows, always joined with their category
// for the category_key/category_label convenience fields.
type HoldingsRepo struct {
	pool *pgxpool.Pool
}

func NewHoldingsRepo(pool *pgxpool.Pool) *HoldingsRepo {
	return &HoldingsRepo{pool: pool}
}

const holdingSelectCols = `
	h.id, h.snapshot_id, h.category_id, c.key, c.label, h.name, h.detail,
	h.value_idr, h.is_liability, h.gram, h.qty, h.brand, h.usd_value, h.currency,
	h.created_at, h.updated_at`

func scanHolding(row interface{ Scan(dest ...any) error }) (domain.Holding, error) {
	var h domain.Holding
	err := row.Scan(
		&h.ID, &h.SnapshotID, &h.CategoryID, &h.CategoryKey, &h.CategoryLabel, &h.Name, &h.Detail,
		&h.ValueIdr, &h.IsLiability, &h.Gram, &h.Qty, &h.Brand, &h.UsdValue, &h.Currency,
		&h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return domain.Holding{}, err
	}
	return h, nil
}

// ListBySnapshot returns every holding in a snapshot, oldest-created first.
func (r *HoldingsRepo) ListBySnapshot(ctx context.Context, snapshotID uuid.UUID) ([]domain.Holding, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+holdingSelectCols+`
		FROM holdings h
		JOIN categories c ON c.id = h.category_id
		WHERE h.snapshot_id = $1
		ORDER BY h.created_at`, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Holding{}
	for rows.Next() {
		h, err := scanHolding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// GetByID returns a single holding (with its snapshot_id) owned by userID
// (via its snapshot). ErrNotFound if missing or owned by someone else.
func (r *HoldingsRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.Holding, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+holdingSelectCols+`
		FROM holdings h
		JOIN categories c ON c.id = h.category_id
		JOIN snapshots s ON s.id = h.snapshot_id
		WHERE h.id = $1 AND s.user_id = $2`, id, userID)
	h, err := scanHolding(row)
	if err != nil {
		return domain.Holding{}, wrapNotFound(err)
	}
	return h, nil
}

// CreateInput is the set of columns needed to insert a holding. CategoryKey
// and CategoryLabel are supplied by the caller (already resolved from the
// categories table) so this method doesn't need a second round trip.
type HoldingWrite struct {
	SnapshotID    uuid.UUID
	CategoryID    int16
	CategoryKey   string
	CategoryLabel string
	Name          string
	Detail        string
	ValueIdr      int64
	IsLiability   bool
	Gram          *float64
	Qty           *float64
	Brand         *string
	UsdValue      *float64
	Currency      *string
}

// Create inserts a new holding and returns the full row (including
// generated id/timestamps).
func (r *HoldingsRepo) Create(ctx context.Context, w HoldingWrite) (domain.Holding, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO holdings (snapshot_id, category_id, name, detail, value_idr, is_liability, gram, qty, brand, usd_value, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`,
		w.SnapshotID, w.CategoryID, w.Name, w.Detail, w.ValueIdr, w.IsLiability, w.Gram, w.Qty, w.Brand, w.UsdValue, w.Currency)

	h := domain.Holding{
		SnapshotID:    w.SnapshotID,
		CategoryID:    w.CategoryID,
		CategoryKey:   w.CategoryKey,
		CategoryLabel: w.CategoryLabel,
		Name:          w.Name,
		Detail:        w.Detail,
		ValueIdr:      w.ValueIdr,
		IsLiability:   w.IsLiability,
		Gram:          w.Gram,
		Qty:           w.Qty,
		Brand:         w.Brand,
		UsdValue:      w.UsdValue,
		Currency:      w.Currency,
	}
	if err := row.Scan(&h.ID, &h.CreatedAt, &h.UpdatedAt); err != nil {
		return domain.Holding{}, err
	}
	return h, nil
}

// Update overwrites an existing holding's mutable fields and returns the
// refreshed row. ErrNotFound if the id doesn't exist or isn't owned by
// userID (via its snapshot).
func (r *HoldingsRepo) Update(ctx context.Context, userID, id uuid.UUID, w HoldingWrite) (domain.Holding, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE holdings
		SET category_id = $1, name = $2, detail = $3, value_idr = $4, is_liability = $5,
			gram = $6, qty = $7, brand = $8, usd_value = $9, currency = $10, updated_at = now()
		WHERE id = $11 AND snapshot_id IN (SELECT id FROM snapshots WHERE user_id = $12)
		RETURNING id, snapshot_id, created_at, updated_at`,
		w.CategoryID, w.Name, w.Detail, w.ValueIdr, w.IsLiability, w.Gram, w.Qty, w.Brand, w.UsdValue, w.Currency, id, userID)

	h := domain.Holding{
		CategoryID:    w.CategoryID,
		CategoryKey:   w.CategoryKey,
		CategoryLabel: w.CategoryLabel,
		Name:          w.Name,
		Detail:        w.Detail,
		ValueIdr:      w.ValueIdr,
		IsLiability:   w.IsLiability,
		Gram:          w.Gram,
		Qty:           w.Qty,
		Brand:         w.Brand,
		UsdValue:      w.UsdValue,
		Currency:      w.Currency,
	}
	if err := row.Scan(&h.ID, &h.SnapshotID, &h.CreatedAt, &h.UpdatedAt); err != nil {
		return domain.Holding{}, wrapNotFound(err)
	}
	return h, nil
}

// Delete removes a holding by id. ErrNotFound if it didn't exist or isn't
// owned by userID (via its snapshot).
func (r *HoldingsRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM holdings
		WHERE id = $1 AND snapshot_id IN (SELECT id FROM snapshots WHERE user_id = $2)`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

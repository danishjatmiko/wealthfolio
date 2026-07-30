package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// BondPurchasesRepo manages bond_purchases rows — the permanent, user-scoped
// ledger of individual USD bond buys. Unlike holdings, these never copy
// forward into a new snapshot and never lock, so ownership is a plain
// user_id column rather than a join through snapshots.
type BondPurchasesRepo struct {
	pool *pgxpool.Pool
}

func NewBondPurchasesRepo(pool *pgxpool.Pool) *BondPurchasesRepo {
	return &BondPurchasesRepo{pool: pool}
}

const bondPurchaseSelectCols = `id, user_id, bond_name, platform, interest_rate, buy_date, maturity_date,
	quantity, face_value_usd, price_usd, accrued_interest_usd, usd_idr_at_purchase, created_at, updated_at`

func scanBondPurchase(row interface{ Scan(dest ...any) error }) (domain.BondPurchase, error) {
	var (
		p             domain.BondPurchase
		buy, maturity time.Time
	)
	err := row.Scan(&p.ID, &p.UserID, &p.BondName, &p.Platform, &p.InterestRate, &buy, &maturity,
		&p.Quantity, &p.FaceValueUsd, &p.PriceUsd, &p.AccruedInterestUsd, &p.UsdIdrAtPurchase,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.BondPurchase{}, err
	}
	p.BuyDate = domain.NewDate(buy)
	p.MaturityDate = domain.NewDate(maturity)
	return p, nil
}

func collectBondPurchases(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}) ([]domain.BondPurchase, error) {
	defer rows.Close()

	out := []domain.BondPurchase{}
	for rows.Next() {
		p, err := scanBondPurchase(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// List returns every purchase for the user, grouped by bond name then
// oldest buy first — the order both the ledger UI and summarizeByName want.
func (r *BondPurchasesRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.BondPurchase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+bondPurchaseSelectCols+`
		FROM bond_purchases WHERE user_id = $1
		ORDER BY bond_name, buy_date, created_at`, userID)
	if err != nil {
		return nil, err
	}
	return collectBondPurchases(rows)
}

// ListActiveAsOf returns the purchases the user actually held on asOf:
// already bought, not yet matured. Drives the Assets auto-fill, where a
// matured bond has returned its principal and is no longer a position.
func (r *BondPurchasesRepo) ListActiveAsOf(ctx context.Context, userID uuid.UUID, asOf domain.Date) ([]domain.BondPurchase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+bondPurchaseSelectCols+`
		FROM bond_purchases
		WHERE user_id = $1 AND buy_date <= $2 AND maturity_date > $2
		ORDER BY bond_name, buy_date, created_at`, userID, asOf.Time)
	if err != nil {
		return nil, err
	}
	return collectBondPurchases(rows)
}

// GetByID returns a single purchase owned by userID. ErrNotFound if it's
// missing or belongs to someone else.
func (r *BondPurchasesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.BondPurchase, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+bondPurchaseSelectCols+`
		FROM bond_purchases WHERE id = $1 AND user_id = $2`, id, userID)
	p, err := scanBondPurchase(row)
	if err != nil {
		return domain.BondPurchase{}, wrapNotFound(err)
	}
	return p, nil
}

// BondPurchaseWrite is the set of columns needed to insert/update a
// purchase.
type BondPurchaseWrite struct {
	BondName           string
	Platform           string
	InterestRate       float64
	BuyDate            domain.Date
	MaturityDate       domain.Date
	Quantity           float64
	FaceValueUsd       float64
	PriceUsd           float64
	AccruedInterestUsd float64
	UsdIdrAtPurchase   float64
}

// Create inserts a new purchase and returns the full row.
func (r *BondPurchasesRepo) Create(ctx context.Context, userID uuid.UUID, w BondPurchaseWrite) (domain.BondPurchase, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO bond_purchases (user_id, bond_name, platform, interest_rate, buy_date, maturity_date,
			quantity, face_value_usd, price_usd, accrued_interest_usd, usd_idr_at_purchase)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING `+bondPurchaseSelectCols,
		userID, w.BondName, w.Platform, w.InterestRate, w.BuyDate.Time, w.MaturityDate.Time,
		w.Quantity, w.FaceValueUsd, w.PriceUsd, w.AccruedInterestUsd, w.UsdIdrAtPurchase)
	return scanBondPurchase(row)
}

// Update overwrites an existing purchase. ErrNotFound if the id doesn't
// exist or isn't owned by userID.
func (r *BondPurchasesRepo) Update(ctx context.Context, userID, id uuid.UUID, w BondPurchaseWrite) (domain.BondPurchase, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE bond_purchases
		SET bond_name = $1, platform = $2, interest_rate = $3, buy_date = $4, maturity_date = $5,
			quantity = $6, face_value_usd = $7, price_usd = $8, accrued_interest_usd = $9,
			usd_idr_at_purchase = $10, updated_at = now()
		WHERE id = $11 AND user_id = $12
		RETURNING `+bondPurchaseSelectCols,
		w.BondName, w.Platform, w.InterestRate, w.BuyDate.Time, w.MaturityDate.Time,
		w.Quantity, w.FaceValueUsd, w.PriceUsd, w.AccruedInterestUsd, w.UsdIdrAtPurchase,
		id, userID)
	p, err := scanBondPurchase(row)
	if err != nil {
		return domain.BondPurchase{}, wrapNotFound(err)
	}
	return p, nil
}

// Delete removes a purchase. ErrNotFound if it didn't exist or isn't owned
// by userID.
func (r *BondPurchasesRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bond_purchases WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// MaxUpdatedAt returns when the user last touched their ledger, or nil if
// they have no purchases — feeds the dashboard's passive-income "updated"
// stamp alongside the same figure from passive_income_sources.
func (r *BondPurchasesRepo) MaxUpdatedAt(ctx context.Context, userID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM bond_purchases WHERE user_id = $1`, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

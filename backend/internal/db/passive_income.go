package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// PassiveIncomeRepo manages passive_income_entries rows — the permanent,
// user-scoped ledger of income actually received (dividends, realized
// capital gains, fund redemptions), joined with their category for the
// category_key/category_label convenience fields.
type PassiveIncomeRepo struct {
	pool *pgxpool.Pool
}

func NewPassiveIncomeRepo(pool *pgxpool.Pool) *PassiveIncomeRepo {
	return &PassiveIncomeRepo{pool: pool}
}

// MaxUpdatedAt returns the most recent updated_at across the user's passive
// income entries, or nil if they have none yet.
func (r *PassiveIncomeRepo) MaxUpdatedAt(ctx context.Context, userID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM passive_income_entries WHERE user_id = $1`, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

func scanPassiveIncome(row interface{ Scan(dest ...any) error }) (domain.PassiveIncomeEntry, error) {
	var (
		p    domain.PassiveIncomeEntry
		date time.Time
	)
	err := row.Scan(&p.ID, &p.UserID, &p.CategoryID, &p.CategoryKey, &p.CategoryLabel,
		&p.Name, &p.AmountIdr, &date, &p.IncomeType, &p.Note, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.PassiveIncomeEntry{}, err
	}
	p.ReceivedDate = domain.NewDate(date)
	return p, nil
}

const passiveIncomeSelect = `
	SELECT p.id, p.user_id, p.category_id, c.key, c.label, p.name, p.amount_idr,
	       p.received_date, p.income_type, p.note, p.created_at, p.updated_at
	FROM passive_income_entries p
	JOIN categories c ON c.id = p.category_id`

// List returns every passive income entry for the user, newest first.
func (r *PassiveIncomeRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.PassiveIncomeEntry, error) {
	rows, err := r.pool.Query(ctx, passiveIncomeSelect+`
		WHERE p.user_id = $1
		ORDER BY p.received_date DESC, p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.PassiveIncomeEntry{}
	for rows.Next() {
		p, err := scanPassiveIncome(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListByYear returns the user's passive income entries received in the
// given calendar year, oldest first — the order the monthly calendar wants.
func (r *PassiveIncomeRepo) ListByYear(ctx context.Context, userID uuid.UUID, year int) ([]domain.PassiveIncomeEntry, error) {
	rows, err := r.pool.Query(ctx, passiveIncomeSelect+`
		WHERE p.user_id = $1 AND EXTRACT(YEAR FROM p.received_date) = $2
		ORDER BY p.received_date, p.created_at`, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.PassiveIncomeEntry{}
	for rows.Next() {
		p, err := scanPassiveIncome(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetByID returns a single passive income entry owned by userID.
// ErrNotFound if missing or owned by someone else.
func (r *PassiveIncomeRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.PassiveIncomeEntry, error) {
	row := r.pool.QueryRow(ctx, passiveIncomeSelect+` WHERE p.id = $1 AND p.user_id = $2`, id, userID)
	p, err := scanPassiveIncome(row)
	if err != nil {
		return domain.PassiveIncomeEntry{}, wrapNotFound(err)
	}
	return p, nil
}

// PassiveIncomeWrite is the set of columns needed to insert/update a
// passive income entry.
type PassiveIncomeWrite struct {
	CategoryID   int16
	Name         string
	AmountIdr    int64
	ReceivedDate domain.Date
	IncomeType   string
	Note         string
}

// Create inserts a new passive income entry and returns the full row.
func (r *PassiveIncomeRepo) Create(ctx context.Context, userID uuid.UUID, w PassiveIncomeWrite) (domain.PassiveIncomeEntry, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO passive_income_entries (user_id, category_id, name, amount_idr, received_date, income_type, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		userID, w.CategoryID, w.Name, w.AmountIdr, w.ReceivedDate.Time, w.IncomeType, w.Note).Scan(&id)
	if err != nil {
		return domain.PassiveIncomeEntry{}, err
	}
	return r.GetByID(ctx, userID, id)
}

// Update overwrites a passive income entry's fields. ErrNotFound if the id
// doesn't exist or isn't owned by userID.
func (r *PassiveIncomeRepo) Update(ctx context.Context, userID, id uuid.UUID, w PassiveIncomeWrite) (domain.PassiveIncomeEntry, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE passive_income_entries
		SET category_id = $1, name = $2, amount_idr = $3, received_date = $4,
		    income_type = $5, note = $6, updated_at = now()
		WHERE id = $7 AND user_id = $8`,
		w.CategoryID, w.Name, w.AmountIdr, w.ReceivedDate.Time, w.IncomeType, w.Note, id, userID)
	if err != nil {
		return domain.PassiveIncomeEntry{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.PassiveIncomeEntry{}, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

// Delete removes a passive income entry by id. ErrNotFound if it didn't
// exist or isn't owned by userID.
func (r *PassiveIncomeRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM passive_income_entries WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SumForYear returns the total amount_idr the user received in the given
// calendar year. Year-scoped rather than all-time because entries are dated
// receipts now: summing every year ever would answer a different question
// than "what is this portfolio paying me", which is what the dashboard and
// the passive_income target both ask.
func (r *PassiveIncomeRepo) SumForYear(ctx context.Context, userID uuid.UUID, year int) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_idr), 0) FROM passive_income_entries
		WHERE user_id = $1 AND EXTRACT(YEAR FROM received_date) = $2`, userID, year).Scan(&total)
	return total, err
}

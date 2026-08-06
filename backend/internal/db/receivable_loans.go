package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// ReceivableLoansRepo manages receivable_loans rows — the permanent,
// user-scoped record of a receivable's terms. Like bond_purchases, these
// never copy forward into a snapshot and never lock, so ownership is a
// plain user_id column rather than a join through snapshots.
type ReceivableLoansRepo struct {
	pool *pgxpool.Pool
}

func NewReceivableLoansRepo(pool *pgxpool.Pool) *ReceivableLoansRepo {
	return &ReceivableLoansRepo{pool: pool}
}

const receivableLoanSelectCols = `id, user_id, borrower_name, start_date,
	term_months, interest_idr, note, created_at, updated_at`

func scanReceivableLoan(row interface{ Scan(dest ...any) error }) (domain.ReceivableLoan, error) {
	var (
		l         domain.ReceivableLoan
		startDate time.Time
	)
	err := row.Scan(&l.ID, &l.UserID, &l.BorrowerName, &startDate,
		&l.TermMonths, &l.InterestIdr, &l.Note, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return domain.ReceivableLoan{}, err
	}
	l.StartDate = domain.NewDate(startDate)
	return l, nil
}

// List returns every loan for the user, newest first.
func (r *ReceivableLoansRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.ReceivableLoan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+receivableLoanSelectCols+`
		FROM receivable_loans WHERE user_id = $1
		ORDER BY start_date DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ReceivableLoan{}
	for rows.Next() {
		l, err := scanReceivableLoan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetByID returns a single loan owned by userID. ErrNotFound if it's
// missing or belongs to someone else.
func (r *ReceivableLoansRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.ReceivableLoan, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+receivableLoanSelectCols+`
		FROM receivable_loans WHERE id = $1 AND user_id = $2`, id, userID)
	l, err := scanReceivableLoan(row)
	if err != nil {
		return domain.ReceivableLoan{}, wrapNotFound(err)
	}
	return l, nil
}

// ReceivableLoanWrite is the set of columns needed to insert/update a loan.
type ReceivableLoanWrite struct {
	BorrowerName string
	StartDate    domain.Date
	TermMonths   int
	InterestIdr  int64
	Note         string
}

// Create inserts a new loan and returns the full row.
func (r *ReceivableLoansRepo) Create(ctx context.Context, userID uuid.UUID, w ReceivableLoanWrite) (domain.ReceivableLoan, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO receivable_loans (user_id, borrower_name, start_date,
			term_months, interest_idr, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+receivableLoanSelectCols,
		userID, w.BorrowerName, w.StartDate.Time,
		w.TermMonths, w.InterestIdr, w.Note)
	return scanReceivableLoan(row)
}

// Update overwrites an existing loan. ErrNotFound if the id doesn't exist
// or isn't owned by userID.
func (r *ReceivableLoansRepo) Update(ctx context.Context, userID, id uuid.UUID, w ReceivableLoanWrite) (domain.ReceivableLoan, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE receivable_loans
		SET borrower_name = $1, start_date = $2,
			term_months = $3, interest_idr = $4, note = $5, updated_at = now()
		WHERE id = $6 AND user_id = $7
		RETURNING `+receivableLoanSelectCols,
		w.BorrowerName, w.StartDate.Time,
		w.TermMonths, w.InterestIdr, w.Note, id, userID)
	l, err := scanReceivableLoan(row)
	if err != nil {
		return domain.ReceivableLoan{}, wrapNotFound(err)
	}
	return l, nil
}

// Delete removes a loan. ErrNotFound if it didn't exist or isn't owned by
// userID.
func (r *ReceivableLoansRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM receivable_loans WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// MaxUpdatedAt returns when the user last touched their loan terms, or nil
// if they have none — feeds the dashboard's passive-income "updated" stamp
// alongside the same figure from bond_purchases and passive_income_entries.
func (r *ReceivableLoansRepo) MaxUpdatedAt(ctx context.Context, userID uuid.UUID) (*domain.Date, error) {
	var d domain.Date
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at) FROM receivable_loans WHERE user_id = $1`, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if d.Time.IsZero() {
		return nil, nil
	}
	return &d, nil
}

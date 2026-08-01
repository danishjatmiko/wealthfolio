package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// BigExpensesRepo manages big_expenses rows — the permanent, user-scoped
// ledger of large one-off purchases. Like bond_purchases, these never copy
// forward into a snapshot and never lock, so ownership is a plain user_id
// column rather than a join through snapshots.
type BigExpensesRepo struct {
	pool *pgxpool.Pool
}

func NewBigExpensesRepo(pool *pgxpool.Pool) *BigExpensesRepo {
	return &BigExpensesRepo{pool: pool}
}

const bigExpenseSelectCols = `id, user_id, name, amount_idr, expense_date, category, created_at, updated_at`

func scanBigExpense(row interface{ Scan(dest ...any) error }) (domain.BigExpense, error) {
	var (
		e    domain.BigExpense
		date time.Time
	)
	err := row.Scan(&e.ID, &e.UserID, &e.Name, &e.AmountIdr, &date, &e.Category, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.BigExpense{}, err
	}
	e.ExpenseDate = domain.NewDate(date)
	return e, nil
}

// List returns every big expense for the user, newest first.
func (r *BigExpensesRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.BigExpense, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+bigExpenseSelectCols+`
		FROM big_expenses WHERE user_id = $1
		ORDER BY expense_date DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.BigExpense{}
	for rows.Next() {
		e, err := scanBigExpense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns a single big expense owned by userID. ErrNotFound if it's
// missing or belongs to someone else.
func (r *BigExpensesRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.BigExpense, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+bigExpenseSelectCols+`
		FROM big_expenses WHERE id = $1 AND user_id = $2`, id, userID)
	e, err := scanBigExpense(row)
	if err != nil {
		return domain.BigExpense{}, wrapNotFound(err)
	}
	return e, nil
}

// BigExpenseWrite is the set of columns needed to insert/update a big
// expense.
type BigExpenseWrite struct {
	Name        string
	AmountIdr   int64
	ExpenseDate domain.Date
	Category    string
}

// Create inserts a new big expense and returns the full row.
func (r *BigExpensesRepo) Create(ctx context.Context, userID uuid.UUID, w BigExpenseWrite) (domain.BigExpense, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO big_expenses (user_id, name, amount_idr, expense_date, category)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+bigExpenseSelectCols,
		userID, w.Name, w.AmountIdr, w.ExpenseDate.Time, w.Category)
	return scanBigExpense(row)
}

// Update overwrites an existing big expense. ErrNotFound if the id doesn't
// exist or isn't owned by userID.
func (r *BigExpensesRepo) Update(ctx context.Context, userID, id uuid.UUID, w BigExpenseWrite) (domain.BigExpense, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE big_expenses
		SET name = $1, amount_idr = $2, expense_date = $3, category = $4, updated_at = now()
		WHERE id = $5 AND user_id = $6
		RETURNING `+bigExpenseSelectCols,
		w.Name, w.AmountIdr, w.ExpenseDate.Time, w.Category, id, userID)
	e, err := scanBigExpense(row)
	if err != nil {
		return domain.BigExpense{}, wrapNotFound(err)
	}
	return e, nil
}

// Delete removes a big expense. ErrNotFound if it didn't exist or isn't
// owned by userID.
func (r *BigExpensesRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM big_expenses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

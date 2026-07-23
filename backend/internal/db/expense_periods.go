package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// ExpensePeriodsRepo manages expense_periods rows.
type ExpensePeriodsRepo struct {
	pool *pgxpool.Pool
}

func NewExpensePeriodsRepo(pool *pgxpool.Pool) *ExpensePeriodsRepo {
	return &ExpensePeriodsRepo{pool: pool}
}

const expensePeriodSelectCols = `id, user_id, start_date, end_date, created_at`

func scanExpensePeriod(row interface{ Scan(dest ...any) error }) (domain.ExpensePeriod, error) {
	var (
		p          domain.ExpensePeriod
		start, end time.Time
	)
	if err := row.Scan(&p.ID, &p.UserID, &start, &end, &p.CreatedAt); err != nil {
		return domain.ExpensePeriod{}, err
	}
	p.StartDate = domain.NewDate(start)
	p.EndDate = domain.NewDate(end)
	return p, nil
}

// ExpensePeriodAgg is a period joined with aggregate figures over its
// budget envelopes and fixed expenses.
type ExpensePeriodAgg struct {
	Period            domain.ExpensePeriod
	ActualTotalIdr    int64
	CommittedTotalIdr int64
}

// ListWithAgg returns every period for the user, newest start_date first,
// with each period's actual total (sum of fixed_expenses.amount_idr) and
// committed total (sum of budget_envelopes.committed_amount_idr).
func (r *ExpensePeriodsRepo) ListWithAgg(ctx context.Context, userID uuid.UUID) ([]ExpensePeriodAgg, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.user_id, p.start_date, p.end_date, p.created_at,
			COALESCE(fe.total, 0) AS actual_total_idr,
			COALESCE(be.total, 0) AS committed_total_idr
		FROM expense_periods p
		LEFT JOIN (
			SELECT period_id, SUM(amount_idr) AS total FROM fixed_expenses GROUP BY period_id
		) fe ON fe.period_id = p.id
		LEFT JOIN (
			SELECT period_id, SUM(committed_amount_idr) AS total FROM budget_envelopes GROUP BY period_id
		) be ON be.period_id = p.id
		WHERE p.user_id = $1
		ORDER BY p.start_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ExpensePeriodAgg{}
	for rows.Next() {
		var (
			agg        ExpensePeriodAgg
			start, end time.Time
		)
		if err := rows.Scan(&agg.Period.ID, &agg.Period.UserID, &start, &end, &agg.Period.CreatedAt, &agg.ActualTotalIdr, &agg.CommittedTotalIdr); err != nil {
			return nil, err
		}
		agg.Period.StartDate = domain.NewDate(start)
		agg.Period.EndDate = domain.NewDate(end)
		out = append(out, agg)
	}
	return out, rows.Err()
}

// GetByID returns a single period owned by userID. ErrNotFound otherwise.
func (r *ExpensePeriodsRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (domain.ExpensePeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+expensePeriodSelectCols+` FROM expense_periods WHERE id = $1 AND user_id = $2`, id, userID)
	p, err := scanExpensePeriod(row)
	if err != nil {
		return domain.ExpensePeriod{}, wrapNotFound(err)
	}
	return p, nil
}

// GetByStartDate returns the user's period starting on the given date.
// ErrNotFound if there isn't one — used to detect a duplicate month before
// creating a new period.
func (r *ExpensePeriodsRepo) GetByStartDate(ctx context.Context, userID uuid.UUID, start domain.Date) (domain.ExpensePeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+expensePeriodSelectCols+` FROM expense_periods WHERE user_id = $1 AND start_date = $2`, userID, start.Time)
	p, err := scanExpensePeriod(row)
	if err != nil {
		return domain.ExpensePeriod{}, wrapNotFound(err)
	}
	return p, nil
}

// GetLatest returns the user's period with the maximum start_date.
// ErrNotFound if the user has no periods yet.
func (r *ExpensePeriodsRepo) GetLatest(ctx context.Context, userID uuid.UUID) (domain.ExpensePeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+expensePeriodSelectCols+`
		FROM expense_periods
		WHERE user_id = $1
		ORDER BY start_date DESC
		LIMIT 1`, userID)
	p, err := scanExpensePeriod(row)
	if err != nil {
		return domain.ExpensePeriod{}, wrapNotFound(err)
	}
	return p, nil
}

// Create inserts a new period for the user with the given bounds.
func (r *ExpensePeriodsRepo) Create(ctx context.Context, userID uuid.UUID, start, end domain.Date) (domain.ExpensePeriod, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO expense_periods (user_id, start_date, end_date)
		VALUES ($1, $2, $3)
		RETURNING `+expensePeriodSelectCols,
		userID, start.Time, end.Time)
	return scanExpensePeriod(row)
}

// Delete removes a period (and, via cascade, every budget envelope and
// fixed expense inside it). Periods never lock, so this is a hard delete,
// unlike Snapshots/DebtSnapshots which soft-delete to preserve history.
// ErrNotFound if it didn't exist or isn't owned by userID.
func (r *ExpensePeriodsRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM expense_periods WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

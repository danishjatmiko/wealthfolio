package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned by repository Get* methods when no row matches.
// Callers (service/httpapi layers) should compare with errors.Is.
var ErrNotFound = errors.New("not found")

// ErrDuplicateName is returned by Create methods backed by a UNIQUE(user_id,
// name) constraint (e.g. expense categories) when the name is already taken.
var ErrDuplicateName = errors.New("name already in use")

// wrapNotFound normalizes pgx.ErrNoRows into the package-level ErrNotFound
// so upper layers never need to import pgx directly.
func wrapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// uniqueViolationCode is Postgres's SQLSTATE for a unique constraint
// violation (23505).
const uniqueViolationCode = "23505"

// wrapUniqueViolation normalizes a Postgres unique-constraint violation
// into dupErr, so upper layers never need to import pgconn directly.
func wrapUniqueViolation(err error, dupErr error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return dupErr
	}
	return err
}

// Repos aggregates every resource repository behind a single struct so
// callers only need to wire one value through the service layer.
type Repos struct {
	Users             *UsersRepo
	Sessions          *SessionsRepo
	Categories        *CategoriesRepo
	Rates             *RatesRepo
	Snapshots         *SnapshotsRepo
	Holdings          *HoldingsRepo
	DebtSnapshots     *DebtSnapshotsRepo
	DebtEntries       *DebtEntriesRepo
	ExpensePeriods    *ExpensePeriodsRepo
	BudgetEnvelopes   *BudgetEnvelopesRepo
	FixedExpenses     *FixedExpensesRepo
	ExpenseCategories *ExpenseCategoriesRepo
	PassiveIncome     *PassiveIncomeRepo
	Targets           *TargetsRepo
	Pool              *pgxpool.Pool
}

// NewRepos builds a Repos bundle backed by the given connection pool.
func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{
		Users:             NewUsersRepo(pool),
		Sessions:          NewSessionsRepo(pool),
		Categories:        NewCategoriesRepo(pool),
		Rates:             NewRatesRepo(pool),
		Snapshots:         NewSnapshotsRepo(pool),
		Holdings:          NewHoldingsRepo(pool),
		DebtSnapshots:     NewDebtSnapshotsRepo(pool),
		DebtEntries:       NewDebtEntriesRepo(pool),
		ExpensePeriods:    NewExpensePeriodsRepo(pool),
		BudgetEnvelopes:   NewBudgetEnvelopesRepo(pool),
		FixedExpenses:     NewFixedExpensesRepo(pool),
		ExpenseCategories: NewExpenseCategoriesRepo(pool),
		PassiveIncome:     NewPassiveIncomeRepo(pool),
		Targets:           NewTargetsRepo(pool),
		Pool:              pool,
	}
}

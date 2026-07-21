package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned by repository Get* methods when no row matches.
// Callers (service/httpapi layers) should compare with errors.Is.
var ErrNotFound = errors.New("not found")

// wrapNotFound normalizes pgx.ErrNoRows into the package-level ErrNotFound
// so upper layers never need to import pgx directly.
func wrapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Repos aggregates every resource repository behind a single struct so
// callers only need to wire one value through the service layer.
type Repos struct {
	Categories    *CategoriesRepo
	Rates         *RatesRepo
	Snapshots     *SnapshotsRepo
	Holdings      *HoldingsRepo
	DebtSnapshots *DebtSnapshotsRepo
	DebtEntries   *DebtEntriesRepo
	PassiveIncome *PassiveIncomeRepo
	Targets       *TargetsRepo
	Pool          *pgxpool.Pool
}

// NewRepos builds a Repos bundle backed by the given connection pool.
func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{
		Categories:    NewCategoriesRepo(pool),
		Rates:         NewRatesRepo(pool),
		Snapshots:     NewSnapshotsRepo(pool),
		Holdings:      NewHoldingsRepo(pool),
		DebtSnapshots: NewDebtSnapshotsRepo(pool),
		DebtEntries:   NewDebtEntriesRepo(pool),
		PassiveIncome: NewPassiveIncomeRepo(pool),
		Targets:       NewTargetsRepo(pool),
		Pool:          pool,
	}
}

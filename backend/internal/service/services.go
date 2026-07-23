package service

import (
	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
)

// Services aggregates every business-logic service behind a single struct
// so httpapi only needs to wire one value through its handlers.
type Services struct {
	Auth              *AuthService
	Holdings          *HoldingsService
	Snapshots         *SnapshotsService
	DebtEntries       *DebtEntriesService
	DebtSnapshots     *DebtSnapshotsService
	ExpensePeriods    *ExpensePeriodsService
	BudgetEnvelopes   *BudgetEnvelopesService
	FixedExpenses     *FixedExpensesService
	ExpenseCategories *ExpenseCategoriesService
	Dashboard         *DashboardService
	Progress          *ProgressService
	Targets           *TargetsService
}

// NewServices builds a Services bundle backed by the given repositories.
func NewServices(repos *db.Repos, cfg config.Config) *Services {
	holdings := NewHoldingsService(repos)
	debtEntries := NewDebtEntriesService(repos)
	return &Services{
		Auth:              NewAuthService(repos, cfg),
		Holdings:          holdings,
		Snapshots:         NewSnapshotsService(repos, holdings),
		DebtEntries:       debtEntries,
		DebtSnapshots:     NewDebtSnapshotsService(repos, debtEntries),
		ExpensePeriods:    NewExpensePeriodsService(repos),
		BudgetEnvelopes:   NewBudgetEnvelopesService(repos),
		FixedExpenses:     NewFixedExpensesService(repos),
		ExpenseCategories: NewExpenseCategoriesService(repos),
		Dashboard:         NewDashboardService(repos),
		Progress:          NewProgressService(repos),
		Targets:           NewTargetsService(repos),
	}
}

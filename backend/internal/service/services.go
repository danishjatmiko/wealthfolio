package service

import (
	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
)

// Services aggregates every business-logic service behind a single struct
// so httpapi only needs to wire one value through its handlers.
type Services struct {
	Auth            *AuthService
	Holdings        *HoldingsService
	Snapshots       *SnapshotsService
	DebtEntries     *DebtEntriesService
	DebtSnapshots   *DebtSnapshotsService
	ExpensePeriods  *ExpensePeriodsService
	BudgetEnvelopes *BudgetEnvelopesService
	FixedExpenses   *FixedExpensesService
	BondPurchases   *BondPurchasesService
	BigExpenses     *BigExpensesService
	PassiveIncome   *PassiveIncomeService
	Dashboard       *DashboardService
	Progress        *ProgressService
	Targets         *TargetsService

	ExpenseSourceMappings *ExpenseSourceMappingsService
	ExpenseIngestion      *ExpenseIngestionService
	NotificationCatalog   *NotificationCatalogService
}

// NewServices builds a Services bundle backed by the given repositories.
func NewServices(repos *db.Repos, cfg config.Config) *Services {
	holdings := NewHoldingsService(repos)
	snapshots := NewSnapshotsService(repos, holdings)
	debtEntries := NewDebtEntriesService(repos)
	notificationCatalog := NewNotificationCatalogService(repos)
	bonds := NewBondPurchasesService(repos, snapshots)
	return &Services{
		Auth:            NewAuthService(repos, cfg),
		Holdings:        holdings,
		Snapshots:       snapshots,
		DebtEntries:     debtEntries,
		DebtSnapshots:   NewDebtSnapshotsService(repos, debtEntries),
		ExpensePeriods:  NewExpensePeriodsService(repos),
		BudgetEnvelopes: NewBudgetEnvelopesService(repos),
		FixedExpenses:   NewFixedExpensesService(repos),
		BondPurchases:   bonds,
		BigExpenses:     NewBigExpensesService(repos),
		PassiveIncome:   NewPassiveIncomeService(repos, bonds),
		Dashboard:       NewDashboardService(repos, bonds),
		Progress:        NewProgressService(repos),
		Targets:         NewTargetsService(repos, bonds),

		ExpenseSourceMappings: NewExpenseSourceMappingsService(repos),
		ExpenseIngestion:      NewExpenseIngestionService(repos, notificationCatalog),
		NotificationCatalog:   notificationCatalog,
	}
}

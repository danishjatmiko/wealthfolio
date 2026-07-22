package service

import (
	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
)

// Services aggregates every business-logic service behind a single struct
// so httpapi only needs to wire one value through its handlers.
type Services struct {
	Auth          *AuthService
	Holdings      *HoldingsService
	Snapshots     *SnapshotsService
	DebtEntries   *DebtEntriesService
	DebtSnapshots *DebtSnapshotsService
	Dashboard     *DashboardService
	Progress      *ProgressService
	Targets       *TargetsService
}

// NewServices builds a Services bundle backed by the given repositories.
func NewServices(repos *db.Repos, cfg config.Config) *Services {
	holdings := NewHoldingsService(repos)
	debtEntries := NewDebtEntriesService(repos)
	return &Services{
		Auth:          NewAuthService(repos, cfg),
		Holdings:      holdings,
		Snapshots:     NewSnapshotsService(repos, holdings),
		DebtEntries:   debtEntries,
		DebtSnapshots: NewDebtSnapshotsService(repos, debtEntries),
		Dashboard:     NewDashboardService(repos),
		Progress:      NewProgressService(repos),
		Targets:       NewTargetsService(repos),
	}
}

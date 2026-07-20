package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// SnapshotsService implements the snapshot mutability rule (only the
// latest snapshot, by snapshot_date, is editable) and snapshot creation.
type SnapshotsService struct {
	repos    *db.Repos
	holdings *HoldingsService
}

func NewSnapshotsService(repos *db.Repos, holdings *HoldingsService) *SnapshotsService {
	return &SnapshotsService{repos: repos, holdings: holdings}
}

// SnapshotSummary is the shape returned by GET /snapshots.
type SnapshotSummary struct {
	ID            uuid.UUID   `json:"id"`
	SnapshotDate  domain.Date `json:"snapshot_date"`
	IsEditable    bool        `json:"is_editable"`
	HoldingsCount int64       `json:"holdings_count"`
	NetEquityIdr  int64       `json:"net_equity_idr"`
}

// SnapshotDetail is the shape returned by GET /snapshots/latest,
// GET /snapshots/{date}, and POST /snapshots.
type SnapshotDetail struct {
	ID           uuid.UUID        `json:"id"`
	SnapshotDate domain.Date      `json:"snapshot_date"`
	IsEditable   bool             `json:"is_editable"`
	Holdings     []domain.Holding `json:"holdings"`
}

// ListSummaries returns every snapshot for the user, newest first, with
// is_editable true only for the single latest one.
func (s *SnapshotsService) ListSummaries(ctx context.Context, userID uuid.UUID) ([]SnapshotSummary, error) {
	aggs, err := s.repos.Snapshots.ListWithAgg(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]SnapshotSummary, 0, len(aggs))
	for i, a := range aggs {
		out = append(out, SnapshotSummary{
			ID:            a.Snapshot.ID,
			SnapshotDate:  a.Snapshot.SnapshotDate,
			IsEditable:    i == 0,
			HoldingsCount: a.HoldingsCount,
			NetEquityIdr:  a.NetEquityIdr,
		})
	}
	return out, nil
}

func (s *SnapshotsService) detailFromSnapshot(ctx context.Context, userID uuid.UUID, snap domain.Snapshot) (SnapshotDetail, error) {
	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return SnapshotDetail{}, err
	}
	holdings, err := s.repos.Holdings.ListBySnapshot(ctx, snap.ID)
	if err != nil {
		return SnapshotDetail{}, err
	}
	return SnapshotDetail{
		ID:           snap.ID,
		SnapshotDate: snap.SnapshotDate,
		IsEditable:   snap.ID == latest.ID,
		Holdings:     holdings,
	}, nil
}

// GetLatestDetail returns the user's latest snapshot with its holdings.
// Returns db.ErrNotFound if the user has no snapshots yet.
func (s *SnapshotsService) GetLatestDetail(ctx context.Context, userID uuid.UUID) (SnapshotDetail, error) {
	snap, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return SnapshotDetail{}, err
	}
	return s.detailFromSnapshot(ctx, userID, snap)
}

// GetByDateDetail returns the user's snapshot for a specific date with its
// holdings. Returns db.ErrNotFound if there isn't one.
func (s *SnapshotsService) GetByDateDetail(ctx context.Context, userID uuid.UUID, date domain.Date) (SnapshotDetail, error) {
	snap, err := s.repos.Snapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return SnapshotDetail{}, err
	}
	return s.detailFromSnapshot(ctx, userID, snap)
}

// ListHoldingsForDate returns just the holdings for the snapshot on the
// given date. Returns db.ErrNotFound if there isn't one.
func (s *SnapshotsService) ListHoldingsForDate(ctx context.Context, userID uuid.UUID, date domain.Date) ([]domain.Holding, error) {
	snap, err := s.repos.Snapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}
	return s.repos.Holdings.ListBySnapshot(ctx, snap.ID)
}

// Create makes a new snapshot for the user on the given date. The date may
// be anything not already used by an existing snapshot — including a date
// before the current latest, to backfill older history. "Latest"/editable
// status is always derived dynamically from MAX(snapshot_date), so backfilled
// snapshots correctly come back locked. If copyFromLatest is true and a
// previous latest snapshot exists, every one of its holdings is duplicated
// into the new snapshot. initialHoldings, if non-empty, are written directly
// into the new snapshot regardless of whether it ends up being the latest —
// this is the only way to populate a backfilled (non-latest) snapshot, since
// every other holdings write path requires the target snapshot to be latest.
func (s *SnapshotsService) Create(ctx context.Context, userID uuid.UUID, date domain.Date, copyFromLatest bool, initialHoldings []HoldingRequest) (SnapshotDetail, error) {
	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	hasLatest := true
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			hasLatest = false
		} else {
			return SnapshotDetail{}, err
		}
	}

	if _, err := s.repos.Snapshots.GetByDate(ctx, userID, date); err == nil {
		return SnapshotDetail{}, ErrSnapshotDateExists
	} else if !errors.Is(err, db.ErrNotFound) {
		return SnapshotDetail{}, err
	}

	newSnap, err := s.repos.Snapshots.Create(ctx, userID, date)
	if err != nil {
		return SnapshotDetail{}, err
	}

	if copyFromLatest && hasLatest {
		if err := s.repos.Holdings.CopyFromSnapshot(ctx, latest.ID, newSnap.ID); err != nil {
			return SnapshotDetail{}, err
		}
	}

	for _, req := range initialHoldings {
		if _, err := s.holdings.CreateUnlocked(ctx, userID, newSnap.ID, req); err != nil {
			return SnapshotDetail{}, err
		}
	}

	holdings, err := s.repos.Holdings.ListBySnapshot(ctx, newSnap.ID)
	if err != nil {
		return SnapshotDetail{}, err
	}

	isEditable := !hasLatest || date.Time.After(latest.SnapshotDate.Time)

	return SnapshotDetail{
		ID:           newSnap.ID,
		SnapshotDate: newSnap.SnapshotDate,
		IsEditable:   isEditable,
		Holdings:     holdings,
	}, nil
}

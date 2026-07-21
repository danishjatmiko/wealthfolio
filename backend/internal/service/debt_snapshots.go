package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// DebtSnapshotsService implements the debt-snapshot mutability rule (only
// the latest snapshot, by snapshot_date, is editable) and snapshot
// creation. Debt snapshots run on their own independent timeline from asset
// snapshots.
type DebtSnapshotsService struct {
	repos   *db.Repos
	entries *DebtEntriesService
}

func NewDebtSnapshotsService(repos *db.Repos, entries *DebtEntriesService) *DebtSnapshotsService {
	return &DebtSnapshotsService{repos: repos, entries: entries}
}

// DebtSnapshotSummary is the shape returned by GET /debt-snapshots.
type DebtSnapshotSummary struct {
	ID           uuid.UUID   `json:"id"`
	SnapshotDate domain.Date `json:"snapshot_date"`
	IsEditable   bool        `json:"is_editable"`
	EntriesCount int64       `json:"entries_count"`
	IOweIdr      int64       `json:"i_owe_idr"`
	OwedToMeIdr  int64       `json:"owed_to_me_idr"`
}

// DebtSnapshotDetail is the shape returned by GET /debt-snapshots/latest,
// GET /debt-snapshots/{date}, and POST /debt-snapshots.
type DebtSnapshotDetail struct {
	ID           uuid.UUID          `json:"id"`
	SnapshotDate domain.Date        `json:"snapshot_date"`
	IsEditable   bool               `json:"is_editable"`
	Entries      []domain.DebtEntry `json:"entries"`
}

// ListSummaries returns every debt snapshot for the user, newest first,
// with is_editable true only for the single latest one.
func (s *DebtSnapshotsService) ListSummaries(ctx context.Context, userID uuid.UUID) ([]DebtSnapshotSummary, error) {
	aggs, err := s.repos.DebtSnapshots.ListWithAgg(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]DebtSnapshotSummary, 0, len(aggs))
	for i, a := range aggs {
		out = append(out, DebtSnapshotSummary{
			ID:           a.Snapshot.ID,
			SnapshotDate: a.Snapshot.SnapshotDate,
			IsEditable:   i == 0,
			EntriesCount: a.EntriesCount,
			IOweIdr:      a.IOweIdr,
			OwedToMeIdr:  a.OwedToMeIdr,
		})
	}
	return out, nil
}

func (s *DebtSnapshotsService) detailFromSnapshot(ctx context.Context, userID uuid.UUID, snap domain.DebtSnapshot) (DebtSnapshotDetail, error) {
	latest, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}
	entries, err := s.repos.DebtEntries.ListByDebtSnapshot(ctx, snap.ID)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}
	return DebtSnapshotDetail{
		ID:           snap.ID,
		SnapshotDate: snap.SnapshotDate,
		IsEditable:   snap.ID == latest.ID,
		Entries:      entries,
	}, nil
}

// GetLatestDetail returns the user's latest debt snapshot with its entries.
// Returns db.ErrNotFound if the user has no debt snapshots yet.
func (s *DebtSnapshotsService) GetLatestDetail(ctx context.Context, userID uuid.UUID) (DebtSnapshotDetail, error) {
	snap, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}
	return s.detailFromSnapshot(ctx, userID, snap)
}

// GetByDateDetail returns the user's debt snapshot for a specific date with
// its entries. Returns db.ErrNotFound if there isn't one.
func (s *DebtSnapshotsService) GetByDateDetail(ctx context.Context, userID uuid.UUID, date domain.Date) (DebtSnapshotDetail, error) {
	snap, err := s.repos.DebtSnapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}
	return s.detailFromSnapshot(ctx, userID, snap)
}

// Create makes a new debt snapshot for the user on the given date. The date
// must be today or later (ErrSnapshotDateInPast otherwise) and not already
// used by an existing debt snapshot. "Latest"/editable status is always
// derived dynamically from MAX(snapshot_date). If copyFromLatest is true and
// a previous latest debt snapshot exists, every one of its entries is
// duplicated into the new snapshot. initialEntries, if non-empty, are
// written directly into the new snapshot regardless of whether it ends up
// being the latest.
func (s *DebtSnapshotsService) Create(ctx context.Context, userID uuid.UUID, date domain.Date, copyFromLatest bool, initialEntries []DebtEntryRequest) (DebtSnapshotDetail, error) {
	today := domain.NewDate(time.Now())
	if date.Time.Before(today.Time) {
		return DebtSnapshotDetail{}, ErrSnapshotDateInPast
	}

	latest, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	hasLatest := true
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			hasLatest = false
		} else {
			return DebtSnapshotDetail{}, err
		}
	}

	if _, err := s.repos.DebtSnapshots.GetByDate(ctx, userID, date); err == nil {
		return DebtSnapshotDetail{}, ErrSnapshotDateExists
	} else if !errors.Is(err, db.ErrNotFound) {
		return DebtSnapshotDetail{}, err
	}

	newSnap, err := s.repos.DebtSnapshots.Create(ctx, userID, date)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}

	if copyFromLatest && hasLatest {
		if err := s.repos.DebtEntries.CopyFromSnapshot(ctx, latest.ID, newSnap.ID); err != nil {
			return DebtSnapshotDetail{}, err
		}
	}

	for _, req := range initialEntries {
		if _, err := s.entries.CreateUnlocked(ctx, newSnap.ID, req); err != nil {
			return DebtSnapshotDetail{}, err
		}
	}

	entries, err := s.repos.DebtEntries.ListByDebtSnapshot(ctx, newSnap.ID)
	if err != nil {
		return DebtSnapshotDetail{}, err
	}

	isEditable := !hasLatest || date.Time.After(latest.SnapshotDate.Time)

	return DebtSnapshotDetail{
		ID:           newSnap.ID,
		SnapshotDate: newSnap.SnapshotDate,
		IsEditable:   isEditable,
		Entries:      entries,
	}, nil
}

// Delete soft-deletes a debt snapshot. Any snapshot may be deleted, not just
// the latest; deleting the current latest simply makes the next-most-recent
// remaining debt snapshot latest/editable again. Returns db.ErrNotFound if
// the snapshot doesn't exist or isn't owned by userID.
func (s *DebtSnapshotsService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.DebtSnapshots.Delete(ctx, userID, id)
}

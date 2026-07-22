package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// DebtEntriesService implements the debt-snapshot mutability rule for
// writes (create/update/delete only allowed against the user's latest debt
// snapshot).
type DebtEntriesService struct {
	repos *db.Repos
}

func NewDebtEntriesService(repos *db.Repos) *DebtEntriesService {
	return &DebtEntriesService{repos: repos}
}

// DebtEntryRequest is the parsed POST/PUT body for a debt entry write.
type DebtEntryRequest struct {
	Name      string
	Type      string
	ValueIdr  int64
	Direction string
}

func (r DebtEntryRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if r.Direction != "i_owe" && r.Direction != "owed_to_me" {
		return fmt.Errorf("%w: direction must be 'i_owe' or 'owed_to_me'", ErrInvalidInput)
	}
	return nil
}

func debtWriteFromRequest(debtSnapshotID uuid.UUID, req DebtEntryRequest) db.DebtEntryWrite {
	return db.DebtEntryWrite{
		DebtSnapshotID: debtSnapshotID,
		Name:           req.Name,
		Type:           req.Type,
		ValueIdr:       req.ValueIdr,
		Direction:      req.Direction,
	}
}

// Create adds a new debt entry to the debt snapshot on the given date.
// Returns db.ErrNotFound if there's no debt snapshot on that date,
// ErrSnapshotLocked if that snapshot isn't the latest.
func (s *DebtEntriesService) Create(ctx context.Context, userID uuid.UUID, date domain.Date, req DebtEntryRequest) (domain.DebtEntry, error) {
	if err := req.validate(); err != nil {
		return domain.DebtEntry{}, err
	}

	snap, err := s.repos.DebtSnapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return domain.DebtEntry{}, err
	}

	latest, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	if err != nil {
		return domain.DebtEntry{}, err
	}
	if snap.ID != latest.ID {
		return domain.DebtEntry{}, ErrSnapshotLocked
	}

	return s.CreateUnlocked(ctx, snap.ID, req)
}

// CreateUnlocked adds a debt entry directly to the given debt snapshot with
// no latest-snapshot check. Used by DebtSnapshotsService.Create to populate
// a freshly-created (possibly backfilled, non-latest) snapshot's initial
// entries — never exposed as a standalone endpoint.
func (s *DebtEntriesService) CreateUnlocked(ctx context.Context, debtSnapshotID uuid.UUID, req DebtEntryRequest) (domain.DebtEntry, error) {
	if err := req.validate(); err != nil {
		return domain.DebtEntry{}, err
	}
	return s.repos.DebtEntries.Create(ctx, debtWriteFromRequest(debtSnapshotID, req))
}

// Update overwrites an existing debt entry. Returns db.ErrNotFound if it
// doesn't exist, ErrSnapshotLocked if its snapshot isn't the latest.
func (s *DebtEntriesService) Update(ctx context.Context, userID uuid.UUID, entryID uuid.UUID, req DebtEntryRequest) (domain.DebtEntry, error) {
	if err := req.validate(); err != nil {
		return domain.DebtEntry{}, err
	}

	existing, err := s.repos.DebtEntries.GetByID(ctx, userID, entryID)
	if err != nil {
		return domain.DebtEntry{}, err
	}

	latest, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	if err != nil {
		return domain.DebtEntry{}, err
	}
	if existing.DebtSnapshotID != latest.ID {
		return domain.DebtEntry{}, ErrSnapshotLocked
	}

	return s.repos.DebtEntries.Update(ctx, userID, entryID, debtWriteFromRequest(existing.DebtSnapshotID, req))
}

// Delete removes a debt entry. Returns db.ErrNotFound if it doesn't exist,
// ErrSnapshotLocked if its snapshot isn't the latest.
func (s *DebtEntriesService) Delete(ctx context.Context, userID uuid.UUID, entryID uuid.UUID) error {
	existing, err := s.repos.DebtEntries.GetByID(ctx, userID, entryID)
	if err != nil {
		return err
	}

	latest, err := s.repos.DebtSnapshots.GetLatest(ctx, userID)
	if err != nil {
		return err
	}
	if existing.DebtSnapshotID != latest.ID {
		return ErrSnapshotLocked
	}

	return s.repos.DebtEntries.Delete(ctx, userID, entryID)
}

package service

import (
	"context"
	"errors"
	"time"

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

// holdingRequestFromHolding rebuilds a HoldingRequest from an existing
// holding so it can be re-submitted through CreateUnlocked. Routing the copy
// through the normal create path (rather than a raw SQL duplicate) means
// gold and USD-linked holdings get repriced against the latest rate entry
// instead of carrying over a stale value_idr.
func holdingRequestFromHolding(h domain.Holding) HoldingRequest {
	valueIdr := h.ValueIdr
	detail := h.Detail
	return HoldingRequest{
		CategoryID: h.CategoryID,
		Name:       h.Name,
		Gram:       h.Gram,
		Qty:        h.Qty,
		Brand:      h.Brand,
		UsdValue:   h.UsdValue,
		Currency:   h.Currency,
		ValueIdr:   &valueIdr,
		Detail:     &detail,
	}
}

// bondHoldingRequests turns the bond ledger into one holding request per
// bond name, as of the given date. Returns the bonds_usd category alongside
// them so callers can identify existing bond rows without a second lookup.
//
// Only purchases already bought and not yet matured count: a matured bond
// has returned its principal and is no longer a position. value_idr is left
// nil so ComputeHoldingValue derives it from usd_value at the latest rate,
// exactly like a hand-entered bond row.
func (s *SnapshotsService) bondHoldingRequests(ctx context.Context, userID uuid.UUID, asOf domain.Date) (domain.Category, []HoldingRequest, error) {
	bondCat, err := s.repos.Categories.GetByKey(ctx, "bonds_usd")
	if err != nil {
		return domain.Category{}, nil, err
	}
	active, err := s.repos.BondPurchases.ListActiveAsOf(ctx, userID, asOf)
	if err != nil {
		return domain.Category{}, nil, err
	}

	summaries := summarizeByName(active, asOf.Time, 0)
	reqs := make([]HoldingRequest, 0, len(summaries))
	for _, sum := range summaries {
		usd := sum.TotalUsd
		name := sum.BondName
		reqs = append(reqs, HoldingRequest{
			CategoryID: bondCat.ID,
			Name:       name,
			UsdValue:   &usd,
		})
	}
	return bondCat, reqs, nil
}

// SyncBonds rebuilds the snapshot's bonds_usd holdings from the ledger,
// replacing whatever is there. Restricted to the latest snapshot: past
// snapshots are frozen history and must not be rewritten from today's
// ledger.
//
// Destructive by design — every existing bonds_usd row in the snapshot goes,
// including a hand-typed one, which is why the UI confirms with counts
// first. Not wrapped in a transaction (consistent with Create's
// copy-forward); a partial failure is recoverable by running it again.
func (s *SnapshotsService) SyncBonds(ctx context.Context, userID uuid.UUID, date domain.Date) (SnapshotDetail, error) {
	snap, err := s.repos.Snapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return SnapshotDetail{}, err
	}
	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return SnapshotDetail{}, err
	}
	if snap.ID != latest.ID {
		return SnapshotDetail{}, ErrSnapshotLocked
	}

	bondCat, reqs, err := s.bondHoldingRequests(ctx, userID, date)
	if err != nil {
		return SnapshotDetail{}, err
	}
	if _, err := s.repos.Holdings.DeleteByCategory(ctx, userID, snap.ID, bondCat.ID); err != nil {
		return SnapshotDetail{}, err
	}
	for _, req := range reqs {
		if _, err := s.holdings.CreateUnlocked(ctx, userID, snap.ID, req); err != nil {
			return SnapshotDetail{}, err
		}
	}

	return s.detailFromSnapshot(ctx, userID, snap)
}

// SyncBondsForLatest re-syncs the bonds_usd holdings on the user's latest
// snapshot from the ledger. Called automatically after every bond_purchases
// write (BondPurchasesService.Create/Update/Delete) so Assets never drifts
// from the ledger without a manual "Sync bonds" click — the two are tied
// together at the application layer rather than by a literal foreign key,
// since one holding row is an aggregate of many purchases and a real FK
// can't express that.
//
// A user with no snapshot yet, or no rate entry logged yet, is not an
// error: the ledger write that triggered this must still succeed even
// though there's nothing to sync onto yet.
func (s *SnapshotsService) SyncBondsForLatest(ctx context.Context, userID uuid.UUID) error {
	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}
		return err
	}
	_, err = s.SyncBonds(ctx, userID, latest.SnapshotDate)
	if errors.Is(err, ErrNoRateEntry) {
		return nil
	}
	return err
}

// Create makes a new snapshot for the user on the given date. The date must
// be today or later (ErrSnapshotDateInPast otherwise) and not already used
// by an existing snapshot. "Latest"/editable status is always derived
// dynamically from MAX(snapshot_date). If copyFromLatest is true and a
// previous latest snapshot exists, every one of its holdings is duplicated
// into the new snapshot. initialHoldings, if non-empty, are written directly
// into the new snapshot regardless of whether it ends up being the latest —
// this is the only way to populate a snapshot dated ahead of an
// already-future latest, since every other holdings write path requires the
// target snapshot to be latest.
func (s *SnapshotsService) Create(ctx context.Context, userID uuid.UUID, date domain.Date, copyFromLatest bool, initialHoldings []HoldingRequest) (SnapshotDetail, error) {
	today := domain.NewDate(time.Now())
	if date.Time.Before(today.Time) {
		return SnapshotDetail{}, ErrSnapshotDateInPast
	}

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

	// The bond ledger, not the previous snapshot, is the source of truth for
	// bonds — but only once it has something in it. With an empty ledger
	// this whole branch is inert and copy-forward behaves exactly as it
	// always did, so a user with a hand-typed bond row can't silently lose
	// it just because the feature shipped.
	bondCat, bondReqs, err := s.bondHoldingRequests(ctx, userID, date)
	if err != nil {
		return SnapshotDetail{}, err
	}
	ledgerOwnsBonds := len(bondReqs) > 0

	if copyFromLatest && hasLatest {
		sourceHoldings, err := s.repos.Holdings.ListBySnapshot(ctx, latest.ID)
		if err != nil {
			return SnapshotDetail{}, err
		}
		for _, h := range sourceHoldings {
			if ledgerOwnsBonds && h.CategoryID == bondCat.ID {
				continue
			}
			if _, err := s.holdings.CreateUnlocked(ctx, userID, newSnap.ID, holdingRequestFromHolding(h)); err != nil {
				return SnapshotDetail{}, err
			}
		}
	}

	for _, req := range bondReqs {
		if _, err := s.holdings.CreateUnlocked(ctx, userID, newSnap.ID, req); err != nil {
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

// Delete soft-deletes a snapshot (and, implicitly, its holdings become
// unreachable — the rows aren't touched). Any snapshot may be deleted, not
// just the latest; deleting the current latest simply makes the
// next-most-recent remaining snapshot latest/editable again. Returns
// db.ErrNotFound if the snapshot doesn't exist or isn't owned by userID.
func (s *SnapshotsService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.Snapshots.Delete(ctx, userID, id)
}

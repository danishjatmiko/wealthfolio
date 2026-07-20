package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// HoldingsService implements holding value derivation plus the snapshot
// mutability rule for writes (create/update/delete only allowed against
// the user's latest snapshot).
type HoldingsService struct {
	repos *db.Repos
}

func NewHoldingsService(repos *db.Repos) *HoldingsService {
	return &HoldingsService{repos: repos}
}

// HoldingRequest is the parsed POST/PUT body for a holding write.
type HoldingRequest struct {
	CategoryID int16
	Name       string
	Gram       *float64
	Qty        *float64
	Brand      *string
	UsdValue   *float64
	Currency   *string
	ValueIdr   *float64
	Detail     *string
}

func derefF64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (r HoldingRequest) toValuationInput() HoldingInput {
	return HoldingInput{
		Gram:     derefF64(r.Gram),
		Qty:      derefF64(r.Qty),
		Brand:    derefStr(r.Brand),
		Currency: derefStr(r.Currency),
		UsdValue: derefF64(r.UsdValue),
		ValueIdr: derefF64(r.ValueIdr),
		Detail:   derefStr(r.Detail),
	}
}

// computeValue fetches the user's latest rate entry (if any) and derives
// value_idr/detail for the given category + input.
func (s *HoldingsService) computeValue(ctx context.Context, userID uuid.UUID, category domain.Category, input HoldingInput) (int64, string, error) {
	var rate *domain.RateEntry
	r, err := s.repos.Rates.GetLatest(ctx, userID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return 0, "", err
		}
		rate = nil
	} else {
		rate = &r
	}
	return ComputeHoldingValue(category.Key, input, rate)
}

func writeFromRequest(snapshotID uuid.UUID, category domain.Category, req HoldingRequest, valueIdr int64, detail string) db.HoldingWrite {
	return db.HoldingWrite{
		SnapshotID:    snapshotID,
		CategoryID:    category.ID,
		CategoryKey:   category.Key,
		CategoryLabel: category.Label,
		Name:          req.Name,
		Detail:        detail,
		ValueIdr:      valueIdr,
		IsLiability:   category.Kind == "liability",
		Gram:          req.Gram,
		Qty:           req.Qty,
		Brand:         req.Brand,
		UsdValue:      req.UsdValue,
		Currency:      req.Currency,
	}
}

// Create adds a new holding to the snapshot on the given date. Returns
// db.ErrNotFound if there's no snapshot on that date, ErrSnapshotLocked if
// that snapshot isn't the latest, ErrInvalidCategory if category_id doesn't
// exist, and ErrNoRateEntry if a rate-linked value was requested with no
// rate entry available.
func (s *HoldingsService) Create(ctx context.Context, userID uuid.UUID, date domain.Date, req HoldingRequest) (domain.Holding, error) {
	snap, err := s.repos.Snapshots.GetByDate(ctx, userID, date)
	if err != nil {
		return domain.Holding{}, err
	}

	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return domain.Holding{}, err
	}
	if snap.ID != latest.ID {
		return domain.Holding{}, ErrSnapshotLocked
	}

	return s.CreateUnlocked(ctx, userID, snap.ID, req)
}

// CreateUnlocked adds a holding directly to the given snapshot with no
// latest-snapshot check. It exists solely for SnapshotsService.Create to
// populate a freshly-created (possibly backfilled, non-latest) snapshot's
// initial holdings in the same request — it is never exposed as a
// standalone endpoint, since that would defeat the immutability guarantee
// that every other write path enforces.
func (s *HoldingsService) CreateUnlocked(ctx context.Context, userID uuid.UUID, snapshotID uuid.UUID, req HoldingRequest) (domain.Holding, error) {
	category, err := s.repos.Categories.GetByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return domain.Holding{}, ErrInvalidCategory
		}
		return domain.Holding{}, err
	}

	valueIdr, detail, err := s.computeValue(ctx, userID, category, req.toValuationInput())
	if err != nil {
		return domain.Holding{}, err
	}

	return s.repos.Holdings.Create(ctx, writeFromRequest(snapshotID, category, req, valueIdr, detail))
}

// Update recomputes and overwrites an existing holding. Returns
// db.ErrNotFound if the holding doesn't exist, ErrSnapshotLocked if its
// snapshot isn't the latest.
func (s *HoldingsService) Update(ctx context.Context, userID uuid.UUID, holdingID uuid.UUID, req HoldingRequest) (domain.Holding, error) {
	existing, err := s.repos.Holdings.GetByID(ctx, holdingID)
	if err != nil {
		return domain.Holding{}, err
	}

	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return domain.Holding{}, err
	}
	if existing.SnapshotID != latest.ID {
		return domain.Holding{}, ErrSnapshotLocked
	}

	category, err := s.repos.Categories.GetByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return domain.Holding{}, ErrInvalidCategory
		}
		return domain.Holding{}, err
	}

	valueIdr, detail, err := s.computeValue(ctx, userID, category, req.toValuationInput())
	if err != nil {
		return domain.Holding{}, err
	}

	return s.repos.Holdings.Update(ctx, holdingID, writeFromRequest(existing.SnapshotID, category, req, valueIdr, detail))
}

// Delete removes a holding. Returns db.ErrNotFound if it doesn't exist,
// ErrSnapshotLocked if its snapshot isn't the latest.
func (s *HoldingsService) Delete(ctx context.Context, userID uuid.UUID, holdingID uuid.UUID) error {
	existing, err := s.repos.Holdings.GetByID(ctx, holdingID)
	if err != nil {
		return err
	}

	latest, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		return err
	}
	if existing.SnapshotID != latest.ID {
		return ErrSnapshotLocked
	}

	return s.repos.Holdings.Delete(ctx, holdingID)
}

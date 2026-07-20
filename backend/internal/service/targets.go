package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// TargetsService computes the derived current_value/percent/lower_is_better
// fields for targets on top of the stored row.
type TargetsService struct {
	repos *db.Repos
}

func NewTargetsService(repos *db.Repos) *TargetsService {
	return &TargetsService{repos: repos}
}

var validMetricTypes = map[string]bool{
	"equity":         true,
	"gold_grams":     true,
	"passive_income": true,
	"debt_ratio":     true,
	"custom":         true,
}

// TargetRequest is the parsed POST/PUT body for a target write.
type TargetRequest struct {
	Name               string
	Year               int
	MetricType         string
	TargetValue        float64
	Unit               string
	ManualCurrentValue *float64
}

func (r TargetRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if r.Year == 0 {
		return fmt.Errorf("%w: year is required", ErrInvalidInput)
	}
	if !validMetricTypes[r.MetricType] {
		return fmt.Errorf("%w: metric_type must be one of equity, gold_grams, passive_income, debt_ratio, custom", ErrInvalidInput)
	}
	return nil
}

func (s *TargetsService) computeCurrentValue(ctx context.Context, userID uuid.UUID, metricType string, manual *float64) (float64, error) {
	switch metricType {
	case "equity":
		holdings, ok, err := s.latestHoldings(ctx, userID)
		if err != nil || !ok {
			return 0, err
		}
		return float64(netEquity(holdings)), nil

	case "gold_grams":
		holdings, ok, err := s.latestHoldings(ctx, userID)
		if err != nil || !ok {
			return 0, err
		}
		var sum float64
		for _, h := range holdings {
			if h.CategoryKey != "logam_mulia" || h.Gram == nil {
				continue
			}
			qty := 1.0
			if h.Qty != nil && *h.Qty != 0 {
				qty = *h.Qty
			}
			sum += *h.Gram * qty
		}
		return sum, nil

	case "passive_income":
		sum, err := s.repos.PassiveIncome.Sum(ctx, userID)
		return float64(sum), err

	case "debt_ratio":
		holdings, ok, err := s.latestHoldings(ctx, userID)
		if err != nil || !ok {
			return 0, err
		}
		netEq := netEquity(holdings)
		iOwe, _, err := s.repos.Debts.SumByDirection(ctx, userID)
		if err != nil {
			return 0, err
		}
		return percentOf(float64(iOwe), float64(netEq)), nil

	case "custom":
		if manual == nil {
			return 0, nil
		}
		return *manual, nil

	default:
		return 0, nil
	}
}

// latestHoldings returns the holdings of the user's latest snapshot, or
// ok=false (no error) if the user has no snapshots yet.
func (s *TargetsService) latestHoldings(ctx context.Context, userID uuid.UUID) ([]domain.Holding, bool, error) {
	snap, err := s.repos.Snapshots.GetLatest(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	holdings, err := s.repos.Holdings.ListBySnapshot(ctx, snap.ID)
	if err != nil {
		return nil, false, err
	}
	return holdings, true, nil
}

func (s *TargetsService) enrich(ctx context.Context, userID uuid.UUID, t domain.Target) (domain.Target, error) {
	cur, err := s.computeCurrentValue(ctx, userID, t.MetricType, t.ManualCurrentValue)
	if err != nil {
		return domain.Target{}, err
	}
	t.CurrentValue = cur
	t.Percent = percentOf(cur, t.TargetValue)
	t.LowerIsBetter = t.MetricType == "debt_ratio"
	return t, nil
}

// List returns every target for the user with current_value/percent/
// lower_is_better computed.
func (s *TargetsService) List(ctx context.Context, userID uuid.UUID) ([]domain.Target, error) {
	targets, err := s.repos.Targets.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Target, 0, len(targets))
	for _, t := range targets {
		enriched, err := s.enrich(ctx, userID, t)
		if err != nil {
			return nil, err
		}
		out = append(out, enriched)
	}
	return out, nil
}

// Create validates and inserts a new target, returning it with computed
// fields.
func (s *TargetsService) Create(ctx context.Context, userID uuid.UUID, req TargetRequest) (domain.Target, error) {
	if err := req.validate(); err != nil {
		return domain.Target{}, err
	}
	t, err := s.repos.Targets.Create(ctx, userID, req.Name, req.Year, req.MetricType, req.TargetValue, req.Unit, req.ManualCurrentValue)
	if err != nil {
		return domain.Target{}, err
	}
	return s.enrich(ctx, userID, t)
}

// Update validates and overwrites a target, returning it with computed
// fields. Returns db.ErrNotFound if the id doesn't exist.
func (s *TargetsService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req TargetRequest) (domain.Target, error) {
	if err := req.validate(); err != nil {
		return domain.Target{}, err
	}
	t, err := s.repos.Targets.Update(ctx, id, req.Name, req.Year, req.MetricType, req.TargetValue, req.Unit, req.ManualCurrentValue)
	if err != nil {
		return domain.Target{}, err
	}
	return s.enrich(ctx, userID, t)
}

// Delete removes a target by id. Returns db.ErrNotFound if it didn't exist.
func (s *TargetsService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repos.Targets.Delete(ctx, id)
}

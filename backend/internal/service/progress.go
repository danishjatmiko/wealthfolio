package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// ProgressService computes the net-equity time series behind GET
// /progress, at monthly/quarterly/yearly granularity.
type ProgressService struct {
	repos *db.Repos
}

func NewProgressService(repos *db.Repos) *ProgressService {
	return &ProgressService{repos: repos}
}

// ProgressPoint is one entry of progress.series.
type ProgressPoint struct {
	Label        string      `json:"label"`
	Date         domain.Date `json:"date"`
	NetEquityIdr int64       `json:"net_equity_idr"`
}

// ProgressDTO is the full GET /progress response.
type ProgressDTO struct {
	Granularity    string          `json:"granularity"`
	Series         []ProgressPoint `json:"series"`
	LatestValueIdr int64           `json:"latest_value_idr"`
	DeltaIdr       int64           `json:"delta_idr"`
	DeltaPct       float64         `json:"delta_pct"`
}

// Get computes the progress series for the user at the given granularity
// ("monthly", "quarterly", or "yearly"; anything else falls back to
// monthly).
func (s *ProgressService) Get(ctx context.Context, userID uuid.UUID, granularity string) (ProgressDTO, error) {
	aggs, err := s.repos.Snapshots.ListWithAggAsc(ctx, userID)
	if err != nil {
		return ProgressDTO{}, err
	}

	var series []ProgressPoint
	switch granularity {
	case "quarterly":
		series = quarterlySeries(aggs)
	case "yearly":
		series = yearlySeries(aggs)
	default:
		granularity = "monthly"
		series = monthlySeries(aggs)
	}

	out := ProgressDTO{Granularity: granularity, Series: series}
	if len(series) > 0 {
		out.LatestValueIdr = series[len(series)-1].NetEquityIdr
	}
	if len(series) >= 2 {
		prev := series[len(series)-2].NetEquityIdr
		out.DeltaIdr = out.LatestValueIdr - prev
		out.DeltaPct = percentOf(float64(out.DeltaIdr), float64(prev))
	}
	return out, nil
}

func monthlySeries(aggs []db.SnapshotAgg) []ProgressPoint {
	out := make([]ProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		out = append(out, ProgressPoint{
			Label:        a.Snapshot.SnapshotDate.Time.Format("Jan '06"),
			Date:         a.Snapshot.SnapshotDate,
			NetEquityIdr: a.NetEquityIdr,
		})
	}
	return out
}

// quarterOf returns 1-4 for the calendar quarter containing month m.
func quarterOf(m time.Month) int {
	return (int(m)-1)/3 + 1
}

// quarterlySeries groups snapshots (already ascending by date) by calendar
// quarter, keeping only the last snapshot seen in each quarter.
func quarterlySeries(aggs []db.SnapshotAgg) []ProgressPoint {
	out := make([]ProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		t := a.Snapshot.SnapshotDate.Time
		label := fmt.Sprintf("Q%d'%02d", quarterOf(t.Month()), t.Year()%100)
		point := ProgressPoint{Label: label, Date: a.Snapshot.SnapshotDate, NetEquityIdr: a.NetEquityIdr}
		if len(out) > 0 && out[len(out)-1].Label == label {
			out[len(out)-1] = point
		} else {
			out = append(out, point)
		}
	}
	return out
}

// yearlySeries groups snapshots (already ascending by date) by calendar
// year, keeping only the last snapshot seen in each year.
func yearlySeries(aggs []db.SnapshotAgg) []ProgressPoint {
	out := make([]ProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		t := a.Snapshot.SnapshotDate.Time
		label := fmt.Sprintf("%d", t.Year())
		point := ProgressPoint{Label: label, Date: a.Snapshot.SnapshotDate, NetEquityIdr: a.NetEquityIdr}
		if len(out) > 0 && out[len(out)-1].Label == label {
			out[len(out)-1] = point
		} else {
			out = append(out, point)
		}
	}
	return out
}

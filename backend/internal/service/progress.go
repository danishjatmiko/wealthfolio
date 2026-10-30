package service

import (
	"context"
	"fmt"
	"sort"
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

// DebtProgressPoint is one entry of debt-progress.series.
type DebtProgressPoint struct {
	Label       string      `json:"label"`
	Date        domain.Date `json:"date"`
	DebtIdr     int64       `json:"debt_idr"`
	OwedToMeIdr int64       `json:"owed_to_me_idr"`
	RatioPct    float64     `json:"ratio_pct"`
}

// DebtProgressDTO is the full GET /debt-progress response.
type DebtProgressDTO struct {
	Granularity    string              `json:"granularity"`
	Series         []DebtProgressPoint `json:"series"`
	LatestDebtIdr  int64               `json:"latest_debt_idr"`
	LatestRatioPct float64             `json:"latest_ratio_pct"`
	DeltaIdr       int64               `json:"delta_idr"`
	DeltaPct       float64             `json:"delta_pct"`
}

// GetDebtProgress computes the debt value + debt-to-equity ratio series for
// the user at the given granularity. Debt snapshots run on their own
// timeline, independent of asset snapshots, so the ratio at each point
// pairs that debt snapshot's totals with the most recent asset snapshot on
// or before its date (0 net equity, and so a 0% ratio, if none exists yet).
func (s *ProgressService) GetDebtProgress(ctx context.Context, userID uuid.UUID, granularity string) (DebtProgressDTO, error) {
	debtAggs, err := s.repos.DebtSnapshots.ListWithAggAsc(ctx, userID)
	if err != nil {
		return DebtProgressDTO{}, err
	}
	assetAggs, err := s.repos.Snapshots.ListWithAggAsc(ctx, userID)
	if err != nil {
		return DebtProgressDTO{}, err
	}

	netEquityAsOf := func(d domain.Date) int64 {
		var net int64
		for _, a := range assetAggs {
			if a.Snapshot.SnapshotDate.Time.After(d.Time) {
				break
			}
			net = a.NetEquityIdr
		}
		return net
	}

	var series []DebtProgressPoint
	switch granularity {
	case "quarterly":
		series = debtQuarterlySeries(debtAggs, netEquityAsOf)
	case "yearly":
		series = debtYearlySeries(debtAggs, netEquityAsOf)
	default:
		granularity = "monthly"
		series = debtMonthlySeries(debtAggs, netEquityAsOf)
	}

	out := DebtProgressDTO{Granularity: granularity, Series: series}
	if len(series) > 0 {
		last := series[len(series)-1]
		out.LatestDebtIdr = last.DebtIdr
		out.LatestRatioPct = last.RatioPct
	}
	if len(series) >= 2 {
		prev := series[len(series)-2]
		out.DeltaIdr = out.LatestDebtIdr - prev.DebtIdr
		out.DeltaPct = percentOf(float64(out.DeltaIdr), float64(prev.DebtIdr))
	}
	return out, nil
}

func debtMonthlySeries(aggs []db.DebtSnapshotAgg, netEquityAsOf func(domain.Date) int64) []DebtProgressPoint {
	out := make([]DebtProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		net := netEquityAsOf(a.Snapshot.SnapshotDate)
		out = append(out, DebtProgressPoint{
			Label:       a.Snapshot.SnapshotDate.Time.Format("Jan '06"),
			Date:        a.Snapshot.SnapshotDate,
			DebtIdr:     a.IOweIdr,
			OwedToMeIdr: a.OwedToMeIdr,
			RatioPct:    percentOf(float64(a.IOweIdr), float64(net)),
		})
	}
	return out
}

func debtQuarterlySeries(aggs []db.DebtSnapshotAgg, netEquityAsOf func(domain.Date) int64) []DebtProgressPoint {
	out := make([]DebtProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		t := a.Snapshot.SnapshotDate.Time
		label := fmt.Sprintf("Q%d'%02d", quarterOf(t.Month()), t.Year()%100)
		net := netEquityAsOf(a.Snapshot.SnapshotDate)
		point := DebtProgressPoint{
			Label: label, Date: a.Snapshot.SnapshotDate, DebtIdr: a.IOweIdr, OwedToMeIdr: a.OwedToMeIdr,
			RatioPct: percentOf(float64(a.IOweIdr), float64(net)),
		}
		if len(out) > 0 && out[len(out)-1].Label == label {
			out[len(out)-1] = point
		} else {
			out = append(out, point)
		}
	}
	return out
}

func debtYearlySeries(aggs []db.DebtSnapshotAgg, netEquityAsOf func(domain.Date) int64) []DebtProgressPoint {
	out := make([]DebtProgressPoint, 0, len(aggs))
	for _, a := range aggs {
		t := a.Snapshot.SnapshotDate.Time
		label := fmt.Sprintf("%d", t.Year())
		net := netEquityAsOf(a.Snapshot.SnapshotDate)
		point := DebtProgressPoint{
			Label: label, Date: a.Snapshot.SnapshotDate, DebtIdr: a.IOweIdr, OwedToMeIdr: a.OwedToMeIdr,
			RatioPct: percentOf(float64(a.IOweIdr), float64(net)),
		}
		if len(out) > 0 && out[len(out)-1].Label == label {
			out[len(out)-1] = point
		} else {
			out = append(out, point)
		}
	}
	return out
}

// ExpenseProgressPoint is one entry of expense-progress.series.
type ExpenseProgressPoint struct {
	Label        string      `json:"label"`
	Date         domain.Date `json:"date"`
	ActualIdr    int64       `json:"actual_idr"`
	CommittedIdr int64       `json:"committed_idr"`
}

// ExpenseProgressDTO is the full GET /expense-progress response.
type ExpenseProgressDTO struct {
	Granularity string                 `json:"granularity"`
	Series      []ExpenseProgressPoint `json:"series"`
}

// progressBucketLabel is the label a date gets at each granularity — the
// same format the net-equity series uses, so the charts read alike.
func progressBucketLabel(granularity string, t time.Time) string {
	switch granularity {
	case "quarterly":
		return fmt.Sprintf("Q%d'%02d", quarterOf(t.Month()), t.Year()%100)
	case "yearly":
		return fmt.Sprintf("%d", t.Year())
	default:
		return t.Format("Jan '06")
	}
}

// GetExpenseProgress computes spend and budget over time, one point per
// month/quarter/year that has at least one expense period — nothing before
// expense tracking began. Each point totals the periods named after it (a
// period is named after the month it ends in, e.g. 25 Sep – 24 Oct is
// October), labelled the same way as the net-equity series.
func (s *ProgressService) GetExpenseProgress(ctx context.Context, userID uuid.UUID, granularity string) (ExpenseProgressDTO, error) {
	if granularity != "quarterly" && granularity != "yearly" {
		granularity = "monthly"
	}
	periods, err := s.repos.ExpensePeriods.ListWithAgg(ctx, userID)
	if err != nil {
		return ExpenseProgressDTO{}, err
	}
	return ExpenseProgressDTO{Granularity: granularity, Series: expenseSeries(periods, granularity)}, nil
}

// expenseSeries buckets periods (in any order) by label, oldest bucket
// first; each point is dated by the latest period end it covers.
func expenseSeries(periods []db.ExpensePeriodAgg, granularity string) []ExpenseProgressPoint {
	sorted := make([]db.ExpensePeriodAgg, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Period.EndDate.Time.Before(sorted[j].Period.EndDate.Time) })

	out := make([]ExpenseProgressPoint, 0, len(sorted))
	for _, p := range sorted {
		label := progressBucketLabel(granularity, p.Period.EndDate.Time)
		if len(out) == 0 || out[len(out)-1].Label != label {
			out = append(out, ExpenseProgressPoint{Label: label})
		}
		last := &out[len(out)-1]
		last.Date = p.Period.EndDate
		last.ActualIdr += p.ActualTotalIdr
		last.CommittedIdr += p.CommittedTotalIdr
	}
	return out
}

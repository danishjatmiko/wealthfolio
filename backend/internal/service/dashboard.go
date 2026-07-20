package service

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
)

// DashboardService assembles the GET /dashboard payload from the latest
// snapshot plus current debts/passive-income/targets.
type DashboardService struct {
	repos *db.Repos
}

func NewDashboardService(repos *db.Repos) *DashboardService {
	return &DashboardService{repos: repos}
}

// CategoryBreakdown is one entry of equity.by_category / allocation.
type CategoryBreakdown struct {
	CategoryKey string  `json:"category_key"`
	Label       string  `json:"label"`
	ColorOKLCH  string  `json:"color_oklch"`
	ValueIdr    int64   `json:"value_idr"`
	Percent     float64 `json:"percent"`
}

// EquityDTO is dashboard.equity.
type EquityDTO struct {
	TotalIdr       int64               `json:"total_idr"`
	InvestedIdr    int64               `json:"invested_idr"`
	InclPassiveIdr int64               `json:"incl_passive_idr"`
	MomChangeIdr   int64               `json:"mom_change_idr"`
	MomChangePct   float64             `json:"mom_change_pct"`
	ByCategory     []CategoryBreakdown `json:"by_category"`
}

// DebtDTO is dashboard.debt.
type DebtDTO struct {
	TotalDebtIdr       int64   `json:"total_debt_idr"`
	TotalReceivableIdr int64   `json:"total_receivable_idr"`
	RatioPct           float64 `json:"ratio_pct"`
}

// PassiveDTO is dashboard.passive.
type PassiveDTO struct {
	PerYearIdr        int64   `json:"per_year_idr"`
	TargetPerYearIdr  int64   `json:"target_per_year_idr"`
	Percent           float64 `json:"percent"`
	PerMonthIdr       int64   `json:"per_month_idr"`
	PerMonthTargetIdr int64   `json:"per_month_target_idr"`
}

// DashboardDTO is the full GET /dashboard response.
type DashboardDTO struct {
	Equity     EquityDTO           `json:"equity"`
	Debt       DebtDTO             `json:"debt"`
	Passive    PassiveDTO          `json:"passive"`
	Allocation []CategoryBreakdown `json:"allocation"`
}

func emptyDashboard() DashboardDTO {
	return DashboardDTO{
		Equity:     EquityDTO{ByCategory: []CategoryBreakdown{}},
		Allocation: []CategoryBreakdown{},
	}
}

// Get computes the full dashboard payload for the user. If the user has no
// snapshots yet, every figure is zero and the category arrays are empty.
func (s *DashboardService) Get(ctx context.Context, userID uuid.UUID) (DashboardDTO, error) {
	out := emptyDashboard()

	aggs, err := s.repos.Snapshots.ListWithAgg(ctx, userID)
	if err != nil {
		return out, err
	}

	// Passive income and its target exist independently of snapshots.
	perYear, err := s.repos.PassiveIncome.Sum(ctx, userID)
	if err != nil {
		return out, err
	}
	targetPerYear, _, err := s.repos.Targets.FirstTargetValueByMetricType(ctx, userID, "passive_income")
	if err != nil {
		return out, err
	}
	targetPerYearIdr := round64(targetPerYear)

	out.Passive = PassiveDTO{
		PerYearIdr:        perYear,
		TargetPerYearIdr:  targetPerYearIdr,
		Percent:           percentOf(float64(perYear), targetPerYear),
		PerMonthIdr:       round64(float64(perYear) / 12),
		PerMonthTargetIdr: round64(float64(targetPerYearIdr) / 12),
	}

	if len(aggs) == 0 {
		return out, nil
	}

	latest := aggs[0]
	holdings, err := s.repos.Holdings.ListBySnapshot(ctx, latest.Snapshot.ID)
	if err != nil {
		return out, err
	}
	categories, err := s.repos.Categories.List(ctx)
	if err != nil {
		return out, err
	}

	colorByKey := map[string]string{}
	orderByKey := map[string]int16{}
	labelByKey := map[string]string{}
	for _, c := range categories {
		colorByKey[c.Key] = c.ColorOKLCH
		orderByKey[c.Key] = c.SortOrder
		labelByKey[c.Key] = c.Label
	}

	invested := investedTotal(holdings)
	liabilities := liabilityTotal(holdings)
	total := invested + liabilities

	var momIdr int64
	var momPct float64
	if len(aggs) >= 2 {
		prev := aggs[1].NetEquityIdr
		momIdr = total - prev
		momPct = percentOf(float64(momIdr), float64(prev))
	}

	iOwe, owedToMe, err := s.repos.Debts.SumByDirection(ctx, userID)
	if err != nil {
		return out, err
	}
	ratioPct := percentOf(float64(iOwe), float64(total))

	grouped := map[string]int64{}
	for _, h := range holdings {
		if h.IsLiability {
			continue
		}
		grouped[h.CategoryKey] += h.ValueIdr
	}
	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return orderByKey[keys[i]] < orderByKey[keys[j]] })

	byCategory := make([]CategoryBreakdown, 0, len(keys))
	for _, k := range keys {
		v := grouped[k]
		byCategory = append(byCategory, CategoryBreakdown{
			CategoryKey: k,
			Label:       labelByKey[k],
			ColorOKLCH:  colorByKey[k],
			ValueIdr:    v,
			Percent:     percentOf(float64(v), float64(invested)),
		})
	}

	out.Equity = EquityDTO{
		TotalIdr:       total,
		InvestedIdr:    invested,
		InclPassiveIdr: total + perYear,
		MomChangeIdr:   momIdr,
		MomChangePct:   momPct,
		ByCategory:     byCategory,
	}
	out.Debt = DebtDTO{
		TotalDebtIdr:       iOwe,
		TotalReceivableIdr: owedToMe,
		RatioPct:           ratioPct,
	}
	out.Allocation = byCategory

	return out, nil
}

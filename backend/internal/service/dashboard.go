package service

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
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
	AsOfDate       *domain.Date        `json:"as_of_date"`
}

// DebtDTO is dashboard.debt.
type DebtDTO struct {
	TotalDebtIdr       int64        `json:"total_debt_idr"`
	TotalReceivableIdr int64        `json:"total_receivable_idr"`
	RatioPct           float64      `json:"ratio_pct"`
	UpdatedAt          *domain.Date `json:"updated_at"`
}

// PassiveDTO is dashboard.passive.
type PassiveDTO struct {
	PerYearIdr        int64        `json:"per_year_idr"`
	TargetPerYearIdr  int64        `json:"target_per_year_idr"`
	Percent           float64      `json:"percent"`
	PerMonthIdr       int64        `json:"per_month_idr"`
	PerMonthTargetIdr int64        `json:"per_month_target_idr"`
	UpdatedAt         *domain.Date `json:"updated_at"`
}

// ExpenseDTO is dashboard.expense. ActualTotalIdr sums every fixed expense
// in the latest period; CommittedTotalIdr sums every envelope's target.
// ActualByEnvelope/CommittedByEnvelope break both down per envelope, for
// the two Dashboard pie charts.
type ExpenseDTO struct {
	PeriodLabel         string              `json:"period_label"`
	ActualTotalIdr      int64               `json:"actual_total_idr"`
	CommittedTotalIdr   int64               `json:"committed_total_idr"`
	ActualByEnvelope    []CategoryBreakdown `json:"actual_by_envelope"`
	CommittedByEnvelope []CategoryBreakdown `json:"committed_by_envelope"`
	UpdatedAt           *domain.Date        `json:"updated_at"`
}

// expenseEnvelopePalette assigns a chart color to each envelope, cycled in
// creation order (envelopes have no seeded color of their own, unlike
// Asset categories). Same OKLCH style/ranges as the seeded Asset palette
// (migrations/00002_seed.sql) for visual consistency.
var expenseEnvelopePalette = []string{
	"oklch(0.62 0.11 34)",
	"oklch(0.60 0.10 152)",
	"oklch(0.60 0.10 248)",
	"oklch(0.70 0.12 56)",
	"oklch(0.66 0.09 196)",
	"oklch(0.58 0.11 292)",
	"oklch(0.66 0.07 322)",
	"oklch(0.58 0.13 28)",
}

// DashboardDTO is the full GET /dashboard response.
type DashboardDTO struct {
	Equity     EquityDTO           `json:"equity"`
	Debt       DebtDTO             `json:"debt"`
	Passive    PassiveDTO          `json:"passive"`
	Expense    ExpenseDTO          `json:"expense"`
	Allocation []CategoryBreakdown `json:"allocation"`
}

func emptyDashboard() DashboardDTO {
	return DashboardDTO{
		Equity: EquityDTO{ByCategory: []CategoryBreakdown{}},
		Expense: ExpenseDTO{
			ActualByEnvelope:    []CategoryBreakdown{},
			CommittedByEnvelope: []CategoryBreakdown{},
		},
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

	passiveUpdatedAt, err := s.repos.PassiveIncome.MaxUpdatedAt(ctx, userID)
	if err != nil {
		return out, err
	}

	out.Passive = PassiveDTO{
		PerYearIdr:        perYear,
		TargetPerYearIdr:  targetPerYearIdr,
		Percent:           percentOf(float64(perYear), targetPerYear),
		PerMonthIdr:       round64(float64(perYear) / 12),
		PerMonthTargetIdr: round64(float64(targetPerYearIdr) / 12),
		UpdatedAt:         passiveUpdatedAt,
	}

	// Debt snapshots run on their own independent timeline from asset
	// snapshots, so this is computed regardless of whether the user has any
	// asset snapshots yet.
	var iOwe, owedToMe int64
	var debtUpdatedAt *domain.Date
	debtAggs, err := s.repos.DebtSnapshots.ListWithAgg(ctx, userID)
	if err != nil {
		return out, err
	}
	if len(debtAggs) > 0 {
		latestDebt := debtAggs[0]
		iOwe = latestDebt.IOweIdr
		owedToMe = latestDebt.OwedToMeIdr
		debtUpdatedAt, err = s.repos.DebtEntries.MaxUpdatedAt(ctx, latestDebt.Snapshot.ID)
		if err != nil {
			return out, err
		}
	}

	// Expense periods run on their own independent timeline too, computed
	// regardless of whether the user has any asset snapshots yet.
	// ActualTotalIdr/CommittedTotalIdr sum every fixed expense/envelope
	// target in the latest period; ActualByEnvelope/CommittedByEnvelope
	// break both down per envelope.
	expense := ExpenseDTO{
		ActualByEnvelope:    []CategoryBreakdown{},
		CommittedByEnvelope: []CategoryBreakdown{},
	}
	latestPeriod, err := s.repos.ExpensePeriods.GetLatest(ctx, userID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return out, err
	}
	if err == nil {
		fixedExpenses, err := s.repos.FixedExpenses.ListByPeriod(ctx, latestPeriod.ID)
		if err != nil {
			return out, err
		}
		envelopes, err := s.repos.BudgetEnvelopes.ListByPeriod(ctx, latestPeriod.ID)
		if err != nil {
			return out, err
		}

		actualByEnvelope := map[uuid.UUID]int64{}
		var actualTotal int64
		var expenseUpdatedAt *domain.Date
		for _, fe := range fixedExpenses {
			actualTotal += fe.AmountIdr
			actualByEnvelope[fe.EnvelopeID] += fe.AmountIdr
			d := domain.NewDate(fe.UpdatedAt)
			if expenseUpdatedAt == nil || d.Time.After(expenseUpdatedAt.Time) {
				expenseUpdatedAt = &d
			}
		}

		var committedTotal int64
		for _, env := range envelopes {
			committedTotal += env.CommittedAmountIdr
		}

		actualByEnv := make([]CategoryBreakdown, 0, len(envelopes))
		committedByEnv := make([]CategoryBreakdown, 0, len(envelopes))
		for i, env := range envelopes {
			color := expenseEnvelopePalette[i%len(expenseEnvelopePalette)]
			actualByEnv = append(actualByEnv, CategoryBreakdown{
				CategoryKey: env.ID.String(),
				Label:       env.Name,
				ColorOKLCH:  color,
				ValueIdr:    actualByEnvelope[env.ID],
				Percent:     percentOf(float64(actualByEnvelope[env.ID]), float64(actualTotal)),
			})
			committedByEnv = append(committedByEnv, CategoryBreakdown{
				CategoryKey: env.ID.String(),
				Label:       env.Name,
				ColorOKLCH:  color,
				ValueIdr:    env.CommittedAmountIdr,
				Percent:     percentOf(float64(env.CommittedAmountIdr), float64(committedTotal)),
			})
		}

		expense = ExpenseDTO{
			PeriodLabel:         periodLabel(latestPeriod.EndDate),
			ActualTotalIdr:      actualTotal,
			CommittedTotalIdr:   committedTotal,
			ActualByEnvelope:    actualByEnv,
			CommittedByEnvelope: committedByEnv,
			UpdatedAt:           expenseUpdatedAt,
		}
	}
	out.Expense = expense

	if len(aggs) == 0 {
		out.Debt = DebtDTO{
			TotalDebtIdr:       iOwe,
			TotalReceivableIdr: owedToMe,
			RatioPct:           0,
			UpdatedAt:          debtUpdatedAt,
		}
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

	asOf := latest.Snapshot.SnapshotDate
	out.Equity = EquityDTO{
		TotalIdr:       total,
		InvestedIdr:    invested,
		InclPassiveIdr: total + perYear,
		MomChangeIdr:   momIdr,
		MomChangePct:   momPct,
		ByCategory:     byCategory,
		AsOfDate:       &asOf,
	}
	out.Debt = DebtDTO{
		TotalDebtIdr:       iOwe,
		TotalReceivableIdr: owedToMe,
		RatioPct:           ratioPct,
		UpdatedAt:          debtUpdatedAt,
	}
	out.Allocation = byCategory

	return out, nil
}

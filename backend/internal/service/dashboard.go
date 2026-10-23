package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// DashboardService assembles the GET /dashboard payload from the latest
// snapshot plus current debts/passive-income/targets.
type DashboardService struct {
	repos       *db.Repos
	bonds       *BondPurchasesService
	receivables *ReceivableLoansService
}

func NewDashboardService(repos *db.Repos, bonds *BondPurchasesService, receivables *ReceivableLoansService) *DashboardService {
	return &DashboardService{repos: repos, bonds: bonds, receivables: receivables}
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

// ExpenseDTO is dashboard.expense (the latest period) and the body of
// GET /dashboard/expense (any one period). ActualTotalIdr sums every fixed
// expense in the period; CommittedTotalIdr sums every envelope's target.
// ActualByEnvelope/CommittedByEnvelope break both down per envelope, for
// the two Dashboard pie charts; Envelopes pairs the two figures per
// envelope for the budget-vs-actual section. PeriodID is nil when the user
// has no periods yet.
type ExpenseDTO struct {
	PeriodID            *uuid.UUID          `json:"period_id"`
	PeriodLabel         string              `json:"period_label"`
	StartDate           *domain.Date        `json:"start_date"`
	EndDate             *domain.Date        `json:"end_date"`
	ActualTotalIdr      int64               `json:"actual_total_idr"`
	CommittedTotalIdr   int64               `json:"committed_total_idr"`
	ActualByEnvelope    []CategoryBreakdown `json:"actual_by_envelope"`
	CommittedByEnvelope []CategoryBreakdown `json:"committed_by_envelope"`
	Envelopes           []EnvelopeBudgetDTO `json:"envelopes"`
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

// EnvelopeBudgetDTO is one envelope's target next to what was actually
// spent against it, in the same order and color as the pie charts.
type EnvelopeBudgetDTO struct {
	EnvelopeID   uuid.UUID `json:"envelope_id"`
	Label        string    `json:"label"`
	ColorOKLCH   string    `json:"color_oklch"`
	CommittedIdr int64     `json:"committed_idr"`
	ActualIdr    int64     `json:"actual_idr"`
}

// DashboardDTO is the full GET /dashboard response.
type DashboardDTO struct {
	Equity     EquityDTO           `json:"equity"`
	Debt       DebtDTO             `json:"debt"`
	Passive    PassiveDTO          `json:"passive"`
	Expense    ExpenseDTO          `json:"expense"`
	Allocation []CategoryBreakdown `json:"allocation"`
}

func emptyExpense() ExpenseDTO {
	return ExpenseDTO{
		ActualByEnvelope:    []CategoryBreakdown{},
		CommittedByEnvelope: []CategoryBreakdown{},
		Envelopes:           []EnvelopeBudgetDTO{},
	}
}

func emptyDashboard() DashboardDTO {
	return DashboardDTO{
		Equity:     EquityDTO{ByCategory: []CategoryBreakdown{}},
		Expense:    emptyExpense(),
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
	// Bond coupons and receivable payments both count alongside hand-entered
	// entries — TargetsService adds the same three summands for metric_type
	// "passive_income", so the two pages always report the same figure.
	//
	// The halves are measured differently on purpose: manual entries are
	// dated receipts, so only the current year's count, while coupons and
	// receivable payments stay the forward-looking full-year figure. That
	// makes this a floor early in the year — the dividends simply haven't
	// been paid yet — that fills in as the year runs, rather than a
	// projection of income never received.
	perYear, err := s.passivePerYear(ctx, userID)
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
	bondsUpdatedAt, err := s.repos.BondPurchases.MaxUpdatedAt(ctx, userID)
	if err != nil {
		return out, err
	}
	if bondsUpdatedAt != nil && (passiveUpdatedAt == nil || bondsUpdatedAt.Time.After(passiveUpdatedAt.Time)) {
		passiveUpdatedAt = bondsUpdatedAt
	}
	receivablesUpdatedAt, err := s.repos.ReceivableLoans.MaxUpdatedAt(ctx, userID)
	if err != nil {
		return out, err
	}
	if receivablesUpdatedAt != nil && (passiveUpdatedAt == nil || receivablesUpdatedAt.Time.After(passiveUpdatedAt.Time)) {
		passiveUpdatedAt = receivablesUpdatedAt
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

		// owedToMe is recomputed from the entries rather than trusted from
		// the aggregate, so it counts each receivable's remaining debt where
		// a loan reports one — the same figure the Debt & Loans page totals.
		latestEntries, err := s.repos.DebtEntries.ListByDebtSnapshot(ctx, latestDebt.Snapshot.ID)
		if err != nil {
			return out, err
		}
		latestDTOs, err := s.receivables.AttachRemainingDebt(ctx, userID, latestEntries)
		if err != nil {
			return out, err
		}
		for _, e := range latestDTOs {
			if e.Direction == "owed_to_me" {
				owedToMe += e.ShownIdr()
			}
		}

		debtUpdatedAt, err = s.repos.DebtEntries.MaxUpdatedAt(ctx, latestDebt.Snapshot.ID)
		if err != nil {
			return out, err
		}
	}

	// Expense periods run on their own independent timeline too, computed
	// regardless of whether the user has any asset snapshots yet.
	latestPeriod, err := s.repos.ExpensePeriods.GetLatest(ctx, userID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return out, err
	}
	if err == nil {
		out.Expense, err = s.expenseFor(ctx, latestPeriod)
		if err != nil {
			return out, err
		}
	}

	if len(aggs) == 0 {
		out.Debt = DebtDTO{
			TotalDebtIdr:       iOwe,
			TotalReceivableIdr: owedToMe,
			RatioPct:           0,
			UpdatedAt:          debtUpdatedAt,
		}
		return out, nil
	}

	out.Equity, out.Allocation, err = s.equityAt(ctx, aggs, 0, perYear)
	if err != nil {
		return out, err
	}
	out.Debt = DebtDTO{
		TotalDebtIdr:       iOwe,
		TotalReceivableIdr: owedToMe,
		RatioPct:           percentOf(float64(iOwe), float64(out.Equity.TotalIdr)),
		UpdatedAt:          debtUpdatedAt,
	}

	return out, nil
}

// passivePerYear is the year's passive income: hand-logged receipts for the
// current year plus the forward-looking bond coupons and receivable
// payments. See Get for why the halves are measured differently.
func (s *DashboardService) passivePerYear(ctx context.Context, userID uuid.UUID) (int64, error) {
	perYear, err := s.repos.PassiveIncome.SumForYear(ctx, userID, time.Now().UTC().Year())
	if err != nil {
		return 0, err
	}
	coupons, err := s.bonds.CouponPerYearIdr(ctx, userID)
	if err != nil {
		return 0, err
	}
	receivableIncome, err := s.receivables.MonthlyIncomePerYearIdr(ctx, userID)
	if err != nil {
		return 0, err
	}
	return perYear + coupons + receivableIncome, nil
}

// DashboardAssetsDTO is the snapshot-dependent half of the dashboard — the
// equity hero and the allocation breakdown — for GET /dashboard/assets.
type DashboardAssetsDTO struct {
	Equity     EquityDTO           `json:"equity"`
	Allocation []CategoryBreakdown `json:"allocation"`
}

// Assets returns the equity hero and allocation as of one of the user's
// asset snapshots, so the Home page can show any snapshot rather than only
// the latest. ErrNotFound if the user has no snapshot on that date.
func (s *DashboardService) Assets(ctx context.Context, userID uuid.UUID, date domain.Date) (DashboardAssetsDTO, error) {
	aggs, err := s.repos.Snapshots.ListWithAgg(ctx, userID)
	if err != nil {
		return DashboardAssetsDTO{}, err
	}
	idx := -1
	for i, a := range aggs {
		if a.Snapshot.SnapshotDate.Time.Equal(date.Time) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return DashboardAssetsDTO{}, db.ErrNotFound
	}
	perYear, err := s.passivePerYear(ctx, userID)
	if err != nil {
		return DashboardAssetsDTO{}, err
	}
	equity, allocation, err := s.equityAt(ctx, aggs, idx, perYear)
	if err != nil {
		return DashboardAssetsDTO{}, err
	}
	return DashboardAssetsDTO{Equity: equity, Allocation: allocation}, nil
}

// equityAt computes the equity hero and allocation for aggs[idx], where
// aggs is newest first (SnapshotsRepo.ListWithAgg); the month-on-month
// change compares against the snapshot just before it, aggs[idx+1].
func (s *DashboardService) equityAt(ctx context.Context, aggs []db.SnapshotAgg, idx int, perYear int64) (EquityDTO, []CategoryBreakdown, error) {
	snap := aggs[idx]
	holdings, err := s.repos.Holdings.ListBySnapshot(ctx, snap.Snapshot.ID)
	if err != nil {
		return EquityDTO{}, nil, err
	}
	categories, err := s.repos.Categories.List(ctx)
	if err != nil {
		return EquityDTO{}, nil, err
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
	if idx+1 < len(aggs) {
		prev := aggs[idx+1].NetEquityIdr
		momIdr = total - prev
		momPct = percentOf(float64(momIdr), float64(prev))
	}

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

	asOf := snap.Snapshot.SnapshotDate
	equity := EquityDTO{
		TotalIdr:       total,
		InvestedIdr:    invested,
		InclPassiveIdr: total + perYear,
		MomChangeIdr:   momIdr,
		MomChangePct:   momPct,
		ByCategory:     byCategory,
		AsOfDate:       &asOf,
	}
	return equity, byCategory, nil
}

// Expense returns the dashboard's expense block for one of the user's
// periods, so the Home page can show any period rather than only the
// latest. ErrNotFound if the period doesn't exist or isn't the user's.
func (s *DashboardService) Expense(ctx context.Context, userID, periodID uuid.UUID) (ExpenseDTO, error) {
	period, err := s.repos.ExpensePeriods.GetByID(ctx, userID, periodID)
	if err != nil {
		return ExpenseDTO{}, err
	}
	return s.expenseFor(ctx, period)
}

// expenseFor builds the expense block for one period: totals, the two
// per-envelope breakdowns for the pie charts, and each envelope's target
// next to its spend. Envelope colors cycle expenseEnvelopePalette in
// envelope order, so an envelope keeps its color across all three views.
func (s *DashboardService) expenseFor(ctx context.Context, period domain.ExpensePeriod) (ExpenseDTO, error) {
	fixedExpenses, err := s.repos.FixedExpenses.ListByPeriod(ctx, period.ID)
	if err != nil {
		return ExpenseDTO{}, err
	}
	envelopes, err := s.repos.BudgetEnvelopes.ListByPeriod(ctx, period.ID)
	if err != nil {
		return ExpenseDTO{}, err
	}

	actualByEnvelope := map[uuid.UUID]int64{}
	var actualTotal int64
	var updatedAt *domain.Date
	for _, fe := range fixedExpenses {
		actualTotal += fe.AmountIdr
		actualByEnvelope[fe.EnvelopeID] += fe.AmountIdr
		d := domain.NewDate(fe.UpdatedAt)
		if updatedAt == nil || d.Time.After(updatedAt.Time) {
			updatedAt = &d
		}
	}

	var committedTotal int64
	for _, env := range envelopes {
		committedTotal += env.CommittedAmountIdr
	}

	out := emptyExpense()
	for i, env := range envelopes {
		color := expenseEnvelopePalette[i%len(expenseEnvelopePalette)]
		actual := actualByEnvelope[env.ID]
		out.ActualByEnvelope = append(out.ActualByEnvelope, CategoryBreakdown{
			CategoryKey: env.ID.String(),
			Label:       env.Name,
			ColorOKLCH:  color,
			ValueIdr:    actual,
			Percent:     percentOf(float64(actual), float64(actualTotal)),
		})
		out.CommittedByEnvelope = append(out.CommittedByEnvelope, CategoryBreakdown{
			CategoryKey: env.ID.String(),
			Label:       env.Name,
			ColorOKLCH:  color,
			ValueIdr:    env.CommittedAmountIdr,
			Percent:     percentOf(float64(env.CommittedAmountIdr), float64(committedTotal)),
		})
		out.Envelopes = append(out.Envelopes, EnvelopeBudgetDTO{
			EnvelopeID:   env.ID,
			Label:        env.Name,
			ColorOKLCH:   color,
			CommittedIdr: env.CommittedAmountIdr,
			ActualIdr:    actual,
		})
	}

	id, start, end := period.ID, period.StartDate, period.EndDate
	out.PeriodID = &id
	out.PeriodLabel = periodLabel(period.EndDate)
	out.StartDate = &start
	out.EndDate = &end
	out.ActualTotalIdr = actualTotal
	out.CommittedTotalIdr = committedTotal
	out.UpdatedAt = updatedAt
	return out, nil
}

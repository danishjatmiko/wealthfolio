package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// BondPurchasesService owns the USD bond ledger: plain CRUD over
// bond_purchases, plus the three derived read shapes the UI needs (per-bond
// summary, whole-ledger summary, coupon calendar).
//
// Unlike passive income — which is repo-direct pass-through CRUD — this
// resource earns a service layer: most of its reads return shapes with no
// table behind them, and SnapshotsService reuses the same name-grouping to
// build Assets holdings.
type BondPurchasesService struct {
	repos     *db.Repos
	snapshots *SnapshotsService
}

func NewBondPurchasesService(repos *db.Repos, snapshots *SnapshotsService) *BondPurchasesService {
	return &BondPurchasesService{repos: repos, snapshots: snapshots}
}

// couponMonthPalette gives each calendar month a stable chart color. Owned
// by the backend for the same reason expenseEnvelopePalette is: months have
// no seeded color, and serving one lets the frontend consume color_oklch
// exactly as it already does for the allocation and expense donuts. Hues
// step 30° with alternating lightness, so neighbouring ring slices stay
// distinguishable.
var couponMonthPalette = [12]string{
	"oklch(0.66 0.10 30)",
	"oklch(0.58 0.10 60)",
	"oklch(0.66 0.10 90)",
	"oklch(0.58 0.10 120)",
	"oklch(0.66 0.10 150)",
	"oklch(0.58 0.10 180)",
	"oklch(0.66 0.10 210)",
	"oklch(0.58 0.10 240)",
	"oklch(0.66 0.10 270)",
	"oklch(0.58 0.10 300)",
	"oklch(0.66 0.10 330)",
	"oklch(0.58 0.10 360)",
}

// BondPurchaseDTO is one ledger row: the stored columns plus everything
// derivable from them.
type BondPurchaseDTO struct {
	ID                 uuid.UUID   `json:"id"`
	BondName           string      `json:"bond_name"`
	Platform           string      `json:"platform"`
	InterestRate       float64     `json:"interest_rate"`
	BuyDate            domain.Date `json:"buy_date"`
	MaturityDate       domain.Date `json:"maturity_date"`
	Quantity           float64     `json:"quantity"`
	FaceValueUsd       float64     `json:"face_value_usd"`
	PriceUsd           float64     `json:"price_usd"`
	AccruedInterestUsd float64     `json:"accrued_interest_usd"`
	UsdIdrAtPurchase   float64     `json:"usd_idr_at_purchase"`

	// TotalUsd is the invested amount (clean price, no accrued);
	// SettlementUsd is what actually left the account.
	TotalUsd          float64 `json:"total_usd"`
	TotalIdr          int64   `json:"total_idr"`
	SettlementUsd     float64 `json:"settlement_usd"`
	SettlementIdr     int64   `json:"settlement_idr"`
	CouponPerCycleUsd float64 `json:"coupon_per_cycle_usd"`
	CouponPerYearUsd  float64 `json:"coupon_per_year_usd"`
	PricePct          float64 `json:"price_pct"`
	YtmPct            float64 `json:"ytm_pct"`
	CouponMonths      []int   `json:"coupon_months"`
	CouponDay         int     `json:"coupon_day"`
	IsMatured         bool    `json:"is_matured"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// bondPurchaseDTO builds a ledger row. rate is the user's latest logged
// USD/IDR — every Rupiah figure converts at it, so the ledger agrees with
// the Assets page rather than mixing in historical rates. The rate actually
// paid stays visible as UsdIdrAtPurchase.
func bondPurchaseDTO(p domain.BondPurchase, today time.Time, rate float64) BondPurchaseDTO {
	return BondPurchaseDTO{
		ID: p.ID, BondName: p.BondName, Platform: p.Platform,
		InterestRate: p.InterestRate, BuyDate: p.BuyDate, MaturityDate: p.MaturityDate,
		Quantity: p.Quantity, FaceValueUsd: p.FaceValueUsd, PriceUsd: p.PriceUsd,
		AccruedInterestUsd: p.AccruedInterestUsd, UsdIdrAtPurchase: p.UsdIdrAtPurchase,

		TotalUsd:          bondTotalUsd(p),
		TotalIdr:          bondTotalIdr(p, rate),
		SettlementUsd:     bondSettlementUsd(p),
		SettlementIdr:     bondSettlementIdr(p, rate),
		CouponPerCycleUsd: bondCouponPerCycleUsd(p),
		CouponPerYearUsd:  bondCouponPerYearUsd(p),
		PricePct:          bondPricePct(p),
		YtmPct:            bondYtmPct(p),
		CouponMonths:      bondCouponMonths(p.MaturityDate),
		CouponDay:         p.MaturityDate.Time.Day(),
		IsMatured:         p.MaturityDate.Time.Before(today),

		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// BondNameSummaryDTO rolls every purchase filed under one bond name into a
// single position. Purchases are nested so the ledger UI can expand a row
// without a second request, the same way ExpensePeriodDetail nests its
// fixed expenses.
type BondNameSummaryDTO struct {
	BondName          string            `json:"bond_name"`
	Platforms         []string          `json:"platforms"`
	InterestRate      float64           `json:"interest_rate"`
	MaturityDate      domain.Date       `json:"maturity_date"`
	Quantity          float64           `json:"quantity"`
	TotalUsd          float64           `json:"total_usd"`
	TotalIdr          int64             `json:"total_idr"`
	SettlementUsd     float64           `json:"settlement_usd"`
	SettlementIdr     int64             `json:"settlement_idr"`
	AverageUsdIdr     float64           `json:"average_usd_idr"`
	FaceTotalUsd      float64           `json:"face_total_usd"`
	CouponPerCycleUsd float64           `json:"coupon_per_cycle_usd"`
	CouponPerYearUsd  float64           `json:"coupon_per_year_usd"`
	YtmPct            float64           `json:"ytm_pct"`
	CouponMonths      []int             `json:"coupon_months"`
	CouponDay         int               `json:"coupon_day"`
	IsMatured         bool              `json:"is_matured"`
	HasConflicts      bool              `json:"has_conflicts"`
	Purchases         []BondPurchaseDTO `json:"purchases"`
}

// BondLedgerSummaryDTO is the whole-portfolio view: every bond name plus
// the totals the ledger page's hero cards show.
type BondLedgerSummaryDTO struct {
	Bonds            []BondNameSummaryDTO `json:"bonds"`
	TotalUsd         float64              `json:"total_usd"`
	TotalIdr         int64                `json:"total_idr"`
	SettlementUsd    float64              `json:"settlement_usd"`
	SettlementIdr    int64                `json:"settlement_idr"`
	AverageUsdIdr    float64              `json:"average_usd_idr"`
	CouponPerYearUsd float64              `json:"coupon_per_year_usd"`
	CouponPerYearIdr int64                `json:"coupon_per_year_idr"`
	YtmPct           float64              `json:"ytm_pct"`
	LatestUsdIdr     float64              `json:"latest_usd_idr"`
}

// CouponEntryDTO is one payment landing on one date for one bond.
type CouponEntryDTO struct {
	BondName  string      `json:"bond_name"`
	Platform  string      `json:"platform"`
	PayDate   domain.Date `json:"pay_date"`
	AmountUsd float64     `json:"amount_usd"`
	AmountIdr int64       `json:"amount_idr"`
}

// CouponMonthDTO is one of the twelve calendar buckets.
type CouponMonthDTO struct {
	Month      int              `json:"month"`
	Label      string           `json:"label"`
	AmountUsd  float64          `json:"amount_usd"`
	AmountIdr  int64            `json:"amount_idr"`
	Percent    float64          `json:"percent"`
	ColorOKLCH string           `json:"color_oklch"`
	Entries    []CouponEntryDTO `json:"entries"`
}

// CouponCalendarDTO is the reworked Passive Income page's payload: bond
// coupons bucketed by month for one reference year, plus the manual
// passive-income figures so the page can show a combined total.
type CouponCalendarDTO struct {
	ReferenceYear      int              `json:"reference_year"`
	Months             []CouponMonthDTO `json:"months"`
	TotalUsd           float64          `json:"total_usd"`
	TotalIdr           int64            `json:"total_idr"`
	LatestUsdIdr       float64          `json:"latest_usd_idr"`
	ManualPerYearIdr   int64            `json:"manual_per_year_idr"`
	ManualPerMonthIdr  int64            `json:"manual_per_month_idr"`
	CombinedPerYearIdr int64            `json:"combined_per_year_idr"`
}

// Input is expected in (bond_name, buy_date) order, which every repo read
// provides; first-seen order is preserved so the UI ordering is stable.
// summarizeByName groups purchases into one position per bond name, with
// every Rupiah figure converted at rate (the latest logged USD/IDR).
// SnapshotsService passes 0 — it only reads names and USD totals, since a
// holding derives its own value_idr downstream.
func summarizeByName(purchases []domain.BondPurchase, today time.Time, rate float64) []BondNameSummaryDTO {
	type agg struct {
		sum         BondNameSummaryDTO
		platforms   map[string]bool
		ytmWeighted float64
		// The average rate paid is a historical fact, so it's blended from
		// the rates actually paid rather than from the displayed figures.
		costIdrAtPurchase int64
	}
	byName := map[string]*agg{}
	order := []string{}

	for _, p := range purchases {
		a, ok := byName[p.BondName]
		if !ok {
			a = &agg{
				sum: BondNameSummaryDTO{
					BondName: p.BondName,
					// The earliest purchase defines the bond's terms; a
					// later one disagreeing is flagged via HasConflicts.
					InterestRate: p.InterestRate,
					MaturityDate: p.MaturityDate,
					CouponMonths: bondCouponMonths(p.MaturityDate),
					CouponDay:    p.MaturityDate.Time.Day(),
					IsMatured:    p.MaturityDate.Time.Before(today),
					Purchases:    []BondPurchaseDTO{},
				},
				platforms: map[string]bool{},
			}
			byName[p.BondName] = a
			order = append(order, p.BondName)
		}

		if p.InterestRate != a.sum.InterestRate || !p.MaturityDate.Time.Equal(a.sum.MaturityDate.Time) {
			a.sum.HasConflicts = true
		}
		if p.Platform != "" {
			a.platforms[p.Platform] = true
		}

		a.sum.Quantity += p.Quantity
		a.sum.TotalUsd += bondTotalUsd(p)
		a.sum.TotalIdr += bondTotalIdr(p, rate)
		a.sum.SettlementUsd += bondSettlementUsd(p)
		a.sum.SettlementIdr += bondSettlementIdr(p, rate)
		a.costIdrAtPurchase += bondCostIdrAtPurchase(p)
		a.sum.FaceTotalUsd += bondFaceTotalUsd(p)
		// Summed rather than recomputed from the group, so a name holding
		// lots at different rates still totals correctly.
		a.sum.CouponPerCycleUsd += bondCouponPerCycleUsd(p)
		a.sum.CouponPerYearUsd += bondCouponPerYearUsd(p)
		// Yields don't average — a big lot bought cheap moves the position's
		// return more than a small one bought dear, so weight by cash in.
		a.ytmWeighted += bondYtmPct(p) * bondSettlementUsd(p)
		a.sum.Purchases = append(a.sum.Purchases, bondPurchaseDTO(p, today, rate))
	}

	out := make([]BondNameSummaryDTO, 0, len(order))
	for _, name := range order {
		a := byName[name]
		a.sum.AverageUsdIdr = averageUsdIdr(a.costIdrAtPurchase, a.sum.SettlementUsd)
		if a.sum.SettlementUsd > 0 {
			a.sum.YtmPct = a.ytmWeighted / a.sum.SettlementUsd
		}
		a.sum.Platforms = sortedKeys(a.platforms)
		out = append(out, a.sum)
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// latestUsdIdr returns the user's most recent USD/IDR rate, or 0 if they
// have none. Deliberately not an error: a user who hasn't logged a rate yet
// should still see their ledger in dollars rather than a 422.
func (s *BondPurchasesService) latestUsdIdr(ctx context.Context, userID uuid.UUID) (float64, error) {
	rate, err := s.repos.Rates.GetLatest(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return rate.UsdIdr, nil
}

// BondPurchaseRequest is the parsed POST/PUT body for a ledger write.
type BondPurchaseRequest struct {
	BondName           string
	Platform           string
	InterestRate       float64
	BuyDate            domain.Date
	MaturityDate       domain.Date
	Quantity           float64
	FaceValueUsd       float64
	PriceUsd           float64
	AccruedInterestUsd float64
	UsdIdrAtPurchase   float64
}

func (r BondPurchaseRequest) validate() error {
	if r.BondName == "" {
		return fmt.Errorf("%w: bond_name is required", ErrInvalidInput)
	}
	if r.InterestRate <= 0 || r.InterestRate > 100 {
		return fmt.Errorf("%w: interest_rate must be a percent between 0 and 100", ErrInvalidInput)
	}
	if r.BuyDate.Time.IsZero() {
		return fmt.Errorf("%w: buy_date is required", ErrInvalidInput)
	}
	if r.MaturityDate.Time.IsZero() {
		return fmt.Errorf("%w: maturity_date is required", ErrInvalidInput)
	}
	if !r.MaturityDate.Time.After(r.BuyDate.Time) {
		return fmt.Errorf("%w: maturity_date must be after buy_date", ErrInvalidInput)
	}
	if r.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
	}
	if r.FaceValueUsd <= 0 {
		return fmt.Errorf("%w: face_value_usd must be greater than 0", ErrInvalidInput)
	}
	if r.PriceUsd <= 0 {
		return fmt.Errorf("%w: price_usd must be greater than 0", ErrInvalidInput)
	}
	if r.AccruedInterestUsd < 0 {
		return fmt.Errorf("%w: accrued_interest_usd cannot be negative", ErrInvalidInput)
	}
	if r.UsdIdrAtPurchase <= 0 {
		return fmt.Errorf("%w: usd_idr_at_purchase must be greater than 0", ErrInvalidInput)
	}
	return nil
}

// defaultPlatform is filled in whenever a purchase arrives with no platform
// named — MyBCA is where most of this ledger's history actually happened,
// so it's a better default than leaving the field blank.
const defaultPlatform = "MyBCA"

func (r BondPurchaseRequest) toWrite() db.BondPurchaseWrite {
	platform := r.Platform
	if platform == "" {
		platform = defaultPlatform
	}
	return db.BondPurchaseWrite{
		BondName: r.BondName, Platform: platform, InterestRate: r.InterestRate,
		BuyDate: r.BuyDate, MaturityDate: r.MaturityDate, Quantity: r.Quantity,
		FaceValueUsd: r.FaceValueUsd, PriceUsd: r.PriceUsd,
		AccruedInterestUsd: r.AccruedInterestUsd, UsdIdrAtPurchase: r.UsdIdrAtPurchase,
	}
}

// List returns every purchase, newest bond name first in ledger order.
func (s *BondPurchasesService) List(ctx context.Context, userID uuid.UUID) ([]BondPurchaseDTO, error) {
	purchases, err := s.repos.BondPurchases.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return nil, err
	}
	today := time.Now().UTC()
	out := make([]BondPurchaseDTO, 0, len(purchases))
	for _, p := range purchases {
		out = append(out, bondPurchaseDTO(p, today, rate))
	}
	return out, nil
}

func (s *BondPurchasesService) Create(ctx context.Context, userID uuid.UUID, req BondPurchaseRequest) (BondPurchaseDTO, error) {
	if err := req.validate(); err != nil {
		return BondPurchaseDTO{}, err
	}
	p, err := s.repos.BondPurchases.Create(ctx, userID, req.toWrite())
	if err != nil {
		return BondPurchaseDTO{}, err
	}
	if err := s.snapshots.SyncBondsForLatest(ctx, userID); err != nil {
		return BondPurchaseDTO{}, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return BondPurchaseDTO{}, err
	}
	return bondPurchaseDTO(p, time.Now().UTC(), rate), nil
}

func (s *BondPurchasesService) Update(ctx context.Context, userID, id uuid.UUID, req BondPurchaseRequest) (BondPurchaseDTO, error) {
	if err := req.validate(); err != nil {
		return BondPurchaseDTO{}, err
	}
	p, err := s.repos.BondPurchases.Update(ctx, userID, id, req.toWrite())
	if err != nil {
		return BondPurchaseDTO{}, err
	}
	if err := s.snapshots.SyncBondsForLatest(ctx, userID); err != nil {
		return BondPurchaseDTO{}, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return BondPurchaseDTO{}, err
	}
	return bondPurchaseDTO(p, time.Now().UTC(), rate), nil
}

// Delete removes a purchase and immediately re-syncs Assets so a deleted
// lot can't keep inflating the Bonds USD holding until the next manual
// sync or snapshot rollover.
func (s *BondPurchasesService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.repos.BondPurchases.Delete(ctx, userID, id); err != nil {
		return err
	}
	return s.snapshots.SyncBondsForLatest(ctx, userID)
}

// Summary rolls the whole ledger up per bond name, with portfolio totals.
func (s *BondPurchasesService) Summary(ctx context.Context, userID uuid.UUID) (BondLedgerSummaryDTO, error) {
	purchases, err := s.repos.BondPurchases.List(ctx, userID)
	if err != nil {
		return BondLedgerSummaryDTO{}, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return BondLedgerSummaryDTO{}, err
	}

	out := BondLedgerSummaryDTO{
		Bonds:        summarizeByName(purchases, time.Now().UTC(), rate),
		LatestUsdIdr: rate,
	}
	var ytmWeighted, ytmBase float64
	for _, b := range out.Bonds {
		out.TotalUsd += b.TotalUsd
		out.TotalIdr += b.TotalIdr
		out.SettlementUsd += b.SettlementUsd
		out.SettlementIdr += b.SettlementIdr
		// Matured bonds have stopped paying, so they contribute neither
		// forward income nor a forward yield, even though their cost stays
		// in the totals.
		if !b.IsMatured {
			out.CouponPerYearUsd += b.CouponPerYearUsd
			ytmWeighted += b.YtmPct * b.SettlementUsd
			ytmBase += b.SettlementUsd
		}
	}
	// Blended from the rates actually paid, not from the displayed Rupiah —
	// those now convert at the latest rate, which would make this trivially
	// echo that rate back instead of telling you what you bought dollars at.
	var costIdr int64
	var costUsd float64
	for _, p := range purchases {
		costIdr += bondCostIdrAtPurchase(p)
		costUsd += bondSettlementUsd(p)
	}
	out.AverageUsdIdr = averageUsdIdr(costIdr, costUsd)
	out.CouponPerYearIdr = round64(out.CouponPerYearUsd * rate)
	if ytmBase > 0 {
		out.YtmPct = ytmWeighted / ytmBase
	}
	return out, nil
}

// CouponCalendar buckets the year's coupon payments into twelve months.
// Always returns exactly twelve entries — a month with no coupons still
// needs a clickable row in the UI, even though the donut won't draw a slice
// for it.
func (s *BondPurchasesService) CouponCalendar(ctx context.Context, userID uuid.UUID, year int) (CouponCalendarDTO, error) {
	purchases, err := s.repos.BondPurchases.List(ctx, userID)
	if err != nil {
		return CouponCalendarDTO{}, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return CouponCalendarDTO{}, err
	}
	manual, err := s.repos.PassiveIncome.Sum(ctx, userID)
	if err != nil {
		return CouponCalendarDTO{}, err
	}

	// Coupons are collected per (bond name, pay date) rather than per
	// purchase, so three lots of the same bond read as one payment.
	type key struct {
		name string
		date string
	}
	merged := map[key]*CouponEntryDTO{}
	perMonth := map[int][]key{}

	for _, p := range purchases {
		amount := bondCouponPerCycleUsd(p)
		for _, d := range bondCouponDatesInYear(p, year) {
			k := key{name: p.BondName, date: d.String()}
			e, ok := merged[k]
			if !ok {
				e = &CouponEntryDTO{BondName: p.BondName, Platform: p.Platform, PayDate: d}
				merged[k] = e
				m := int(d.Time.Month())
				perMonth[m] = append(perMonth[m], k)
			}
			e.AmountUsd += amount
		}
	}

	out := CouponCalendarDTO{
		ReferenceYear:    year,
		Months:           make([]CouponMonthDTO, 0, 12),
		LatestUsdIdr:     rate,
		ManualPerYearIdr: manual,
	}

	for m := 1; m <= 12; m++ {
		month := CouponMonthDTO{
			Month:      m,
			Label:      time.Month(m).String(),
			ColorOKLCH: couponMonthPalette[m-1],
			Entries:    []CouponEntryDTO{},
		}
		for _, k := range perMonth[m] {
			e := merged[k]
			e.AmountIdr = round64(e.AmountUsd * rate)
			month.AmountUsd += e.AmountUsd
			month.Entries = append(month.Entries, *e)
		}
		sort.Slice(month.Entries, func(i, j int) bool {
			if !month.Entries[i].PayDate.Time.Equal(month.Entries[j].PayDate.Time) {
				return month.Entries[i].PayDate.Time.Before(month.Entries[j].PayDate.Time)
			}
			return month.Entries[i].BondName < month.Entries[j].BondName
		})
		month.AmountIdr = round64(month.AmountUsd * rate)
		out.TotalUsd += month.AmountUsd
		out.Months = append(out.Months, month)
	}

	out.TotalIdr = round64(out.TotalUsd * rate)
	for i := range out.Months {
		out.Months[i].Percent = percentOf(out.Months[i].AmountUsd, out.TotalUsd)
	}
	out.ManualPerMonthIdr = round64(float64(manual) / 12)
	out.CombinedPerYearIdr = out.TotalIdr + manual
	return out, nil
}

// CouponPerYearIdr is the user's forward-looking annual bond coupon income
// in Rupiah, converted at their latest logged rate (0 when they have none).
// Shared by the dashboard and the targets service so the two always agree
// on what "passive income" includes.
func (s *BondPurchasesService) CouponPerYearIdr(ctx context.Context, userID uuid.UUID) (int64, error) {
	purchases, err := s.repos.BondPurchases.List(ctx, userID)
	if err != nil {
		return 0, err
	}
	rate, err := s.latestUsdIdr(ctx, userID)
	if err != nil {
		return 0, err
	}

	today := time.Now().UTC()
	var perYearUsd float64
	for _, p := range purchases {
		if p.MaturityDate.Time.Before(today) {
			continue
		}
		perYearUsd += bondCouponPerYearUsd(p)
	}
	return round64(perYearUsd * rate), nil
}

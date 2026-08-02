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

// PassiveIncomeService owns the Passive Income page: CRUD over
// passive_income_entries, plus the per-year calendar that merges those
// hand-logged receipts with the bond coupons falling in the same months.
//
// The merge is why this resource outgrew the repo-direct pass-through it
// used to be — the page's whole point is one combined picture of income, so
// something has to own combining the two sources, and neither the bond
// ledger nor a bare repo is the right home for it.
type PassiveIncomeService struct {
	repos *db.Repos
	bonds *BondPurchasesService
}

func NewPassiveIncomeService(repos *db.Repos, bonds *BondPurchasesService) *PassiveIncomeService {
	return &PassiveIncomeService{repos: repos, bonds: bonds}
}

// Income entry kinds. Coupons are derived from bond_purchases and have no
// row of their own, so only "manual" entries carry an ID the UI can edit.
const (
	incomeKindCoupon = "coupon"
	incomeKindManual = "manual"
)

// defaultIncomeType is filled in whenever an entry arrives with no type
// named, so every row always has one to group by.
const defaultIncomeType = "Other"

// PassiveIncomeEntryDTO is one ledger row — a flat re-declaration of the
// stored columns, house style.
type PassiveIncomeEntryDTO struct {
	ID            uuid.UUID   `json:"id"`
	CategoryID    int16       `json:"category_id"`
	CategoryKey   string      `json:"category_key"`
	CategoryLabel string      `json:"category_label"`
	Name          string      `json:"name"`
	AmountIdr     int64       `json:"amount_idr"`
	ReceivedDate  domain.Date `json:"received_date"`
	IncomeType    string      `json:"income_type"`
	Note          string      `json:"note"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func passiveIncomeEntryDTO(p domain.PassiveIncomeEntry) PassiveIncomeEntryDTO {
	return PassiveIncomeEntryDTO{
		ID: p.ID, CategoryID: p.CategoryID, CategoryKey: p.CategoryKey,
		CategoryLabel: p.CategoryLabel, Name: p.Name, AmountIdr: p.AmountIdr,
		ReceivedDate: p.ReceivedDate, IncomeType: p.IncomeType, Note: p.Note,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// IncomeEntryDTO is one payment landing on one date, from either source.
// Kind tells them apart: a "coupon" is derived from the bond ledger and
// carries a dollar amount; a "manual" entry is a row in
// passive_income_entries, IDR-only, and its ID is what the UI edits.
//
// ID is a string rather than a uuid.UUID precisely so a coupon can leave it
// empty — a zero uuid would serialize as a real-looking all-zeros id, which
// every coupon in a month would then share.
//
// The category fields repeat what the month's slices already say in
// aggregate, but per entry: without them the detail list can't be grouped
// or sorted by asset class, since Source carries the income type (Dividen,
// Bunga) for manual entries and the platform for coupons — neither of which
// is the category.
type IncomeEntryDTO struct {
	Kind          string      `json:"kind"`
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Source        string      `json:"source"`
	CategoryKey   string      `json:"category_key"`
	CategoryLabel string      `json:"category_label"`
	ColorOKLCH    string      `json:"color_oklch"`
	PayDate       domain.Date `json:"pay_date"`
	AmountUsd     float64     `json:"amount_usd"`
	AmountIdr     int64       `json:"amount_idr"`
	Note          string      `json:"note"`
}

// IncomeMonthSliceDTO is one category's contribution to a single month —
// what the month's bar is stacked out of. Emitted in the same order for
// every month (see Calendar), so a category keeps its position along the
// bar and months stay visually comparable.
type IncomeMonthSliceDTO struct {
	CategoryKey string `json:"category_key"`
	Label       string `json:"label"`
	ColorOKLCH  string `json:"color_oklch"`
	AmountIdr   int64  `json:"amount_idr"`
}

// IncomeMonthDTO is one of the twelve calendar buckets. The two sources
// stay separately visible alongside the combined AmountIdr — coupons are
// forward-looking and dollar-denominated, manual entries are realized and
// in Rupiah, and collapsing that distinction would hide it.
type IncomeMonthDTO struct {
	Month     int                   `json:"month"`
	Label     string                `json:"label"`
	CouponUsd float64               `json:"coupon_usd"`
	CouponIdr int64                 `json:"coupon_idr"`
	ManualIdr int64                 `json:"manual_idr"`
	AmountIdr int64                 `json:"amount_idr"`
	Percent   float64               `json:"percent"`
	Slices    []IncomeMonthSliceDTO `json:"slices"`
	Entries   []IncomeEntryDTO      `json:"entries"`
}

// IncomeCategoryDTO is one asset class's share of the year's income —
// which part of the portfolio actually paid out. Bond coupons file under
// the bonds_usd category so the ring covers every source, not just the
// hand-logged ones.
type IncomeCategoryDTO struct {
	CategoryKey string  `json:"category_key"`
	Label       string  `json:"label"`
	ColorOKLCH  string  `json:"color_oklch"`
	AmountIdr   int64   `json:"amount_idr"`
	Percent     float64 `json:"percent"`
	Count       int     `json:"count"`
}

// IncomeCalendarDTO is the Passive Income page's payload: every source of
// income for one reference year, bucketed by month and by category.
type IncomeCalendarDTO struct {
	ReferenceYear  int                 `json:"reference_year"`
	Months         []IncomeMonthDTO    `json:"months"`
	Categories     []IncomeCategoryDTO `json:"categories"`
	CouponTotalUsd float64             `json:"coupon_total_usd"`
	CouponTotalIdr int64               `json:"coupon_total_idr"`
	ManualTotalIdr int64               `json:"manual_total_idr"`
	TotalIdr       int64               `json:"total_idr"`
	LatestUsdIdr   float64             `json:"latest_usd_idr"`
}

// bondIncomeCategoryKey is the category coupons are filed under. Coupons
// are derived from bond_purchases, which has no category_id of its own —
// the Assets page already surfaces bonds under this same key, so reusing it
// keeps the ring's colors matching the allocation donut.
const bondIncomeCategoryKey = "bonds_usd"

// resolveCategory looks a category up by key, falling back to a neutral
// stand-in named after the key itself. A category row deleted out from
// under an entry would otherwise drop that entry's income off the ring and
// out of the stacked bars entirely; showing it unstyled beats losing it.
func resolveCategory(byKey map[string]domain.Category, key string) domain.Category {
	if cat, ok := byKey[key]; ok {
		return cat
	}
	return domain.Category{Key: key, Label: key, ColorOKLCH: "oklch(0.70 0.02 260)"}
}

// List returns every passive income entry for the user, newest first.
func (s *PassiveIncomeService) List(ctx context.Context, userID uuid.UUID) ([]PassiveIncomeEntryDTO, error) {
	entries, err := s.repos.PassiveIncome.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]PassiveIncomeEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, passiveIncomeEntryDTO(e))
	}
	return out, nil
}

// PassiveIncomeRequest is the parsed POST/PUT body for a ledger write.
type PassiveIncomeRequest struct {
	CategoryID   int16
	Name         string
	AmountIdr    int64
	ReceivedDate domain.Date
	IncomeType   string
	Note         string
}

func (r PassiveIncomeRequest) validate() error {
	if r.CategoryID == 0 {
		return fmt.Errorf("%w: category_id is required", ErrInvalidInput)
	}
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	// Deliberately not "> 0", unlike big expenses: a realized capital loss
	// or a cut-loss redemption belongs in this ledger at its true sign.
	if r.AmountIdr == 0 {
		return fmt.Errorf("%w: amount_idr cannot be 0", ErrInvalidInput)
	}
	if r.ReceivedDate.Time.IsZero() {
		return fmt.Errorf("%w: received_date is required", ErrInvalidInput)
	}
	return nil
}

func (r PassiveIncomeRequest) toWrite() db.PassiveIncomeWrite {
	incomeType := r.IncomeType
	if incomeType == "" {
		incomeType = defaultIncomeType
	}
	return db.PassiveIncomeWrite{
		CategoryID: r.CategoryID, Name: r.Name, AmountIdr: r.AmountIdr,
		ReceivedDate: r.ReceivedDate, IncomeType: incomeType, Note: r.Note,
	}
}

// validateCategory rejects a category_id with no matching row up front, so
// the caller gets ErrInvalidCategory rather than a raw FK violation.
func (s *PassiveIncomeService) validateCategory(ctx context.Context, id int16) error {
	if _, err := s.repos.Categories.GetByID(ctx, id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrInvalidCategory
		}
		return err
	}
	return nil
}

func (s *PassiveIncomeService) Create(ctx context.Context, userID uuid.UUID, req PassiveIncomeRequest) (PassiveIncomeEntryDTO, error) {
	if err := req.validate(); err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	if err := s.validateCategory(ctx, req.CategoryID); err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	e, err := s.repos.PassiveIncome.Create(ctx, userID, req.toWrite())
	if err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	return passiveIncomeEntryDTO(e), nil
}

func (s *PassiveIncomeService) Update(ctx context.Context, userID, id uuid.UUID, req PassiveIncomeRequest) (PassiveIncomeEntryDTO, error) {
	if err := req.validate(); err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	if err := s.validateCategory(ctx, req.CategoryID); err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	e, err := s.repos.PassiveIncome.Update(ctx, userID, id, req.toWrite())
	if err != nil {
		return PassiveIncomeEntryDTO{}, err
	}
	return passiveIncomeEntryDTO(e), nil
}

func (s *PassiveIncomeService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.PassiveIncome.Delete(ctx, userID, id)
}

// Calendar buckets the year's income into twelve months and by asset
// category, merging bond coupons with hand-logged receipts so one row
// answers "what did this month pay me" and one ring answers "which part of
// the portfolio paid it". Always returns exactly twelve months — a month
// with nothing in it still needs a clickable row in the UI, even though the
// donut won't draw a slice for it.
func (s *PassiveIncomeService) Calendar(ctx context.Context, userID uuid.UUID, year int) (IncomeCalendarDTO, error) {
	coupons, rate, err := s.bonds.couponEntriesInYear(ctx, userID, year)
	if err != nil {
		return IncomeCalendarDTO{}, err
	}
	manual, err := s.repos.PassiveIncome.ListByYear(ctx, userID, year)
	if err != nil {
		return IncomeCalendarDTO{}, err
	}
	categories, err := s.repos.Categories.List(ctx)
	if err != nil {
		return IncomeCalendarDTO{}, err
	}
	byKey := make(map[string]domain.Category, len(categories))
	for _, c := range categories {
		byKey[c.Key] = c
	}

	// Accumulated per category key, then ordered by size below — a category
	// keeps its seeded color either way, so unlike the month palette there's
	// no index to keep stable.
	type catAgg struct {
		amountIdr int64
		count     int
	}
	catTotals := map[string]*catAgg{}
	// Same accumulation one level down, so each month's bar can be stacked
	// out of the categories that actually paid into it: [month index][key].
	monthCatTotals := make([]map[string]int64, 12)
	for i := range monthCatTotals {
		monthCatTotals[i] = map[string]int64{}
	}
	addToCategory := func(key string, monthIdx int, amountIdr int64) {
		agg, ok := catTotals[key]
		if !ok {
			agg = &catAgg{}
			catTotals[key] = agg
		}
		agg.amountIdr += amountIdr
		agg.count++
		monthCatTotals[monthIdx][key] += amountIdr
	}

	out := IncomeCalendarDTO{
		ReferenceYear: year,
		Months:        make([]IncomeMonthDTO, 12),
		Categories:    []IncomeCategoryDTO{},
		LatestUsdIdr:  rate,
	}
	for i := range out.Months {
		out.Months[i] = IncomeMonthDTO{
			Month:   i + 1,
			Label:   time.Month(i + 1).String(),
			Slices:  []IncomeMonthSliceDTO{},
			Entries: []IncomeEntryDTO{},
		}
	}

	bondCat := resolveCategory(byKey, bondIncomeCategoryKey)
	for _, c := range coupons {
		idx := int(c.PayDate.Time.Month()) - 1
		m := &out.Months[idx]
		c.CategoryKey = bondCat.Key
		c.CategoryLabel = bondCat.Label
		c.ColorOKLCH = bondCat.ColorOKLCH
		m.CouponUsd += c.AmountUsd
		m.CouponIdr += c.AmountIdr
		m.Entries = append(m.Entries, c)
		out.CouponTotalUsd += c.AmountUsd
		out.CouponTotalIdr += c.AmountIdr
		addToCategory(bondIncomeCategoryKey, idx, c.AmountIdr)
	}

	for _, e := range manual {
		idx := int(e.ReceivedDate.Time.Month()) - 1
		m := &out.Months[idx]
		cat := resolveCategory(byKey, e.CategoryKey)
		m.ManualIdr += e.AmountIdr
		m.Entries = append(m.Entries, IncomeEntryDTO{
			Kind:          incomeKindManual,
			ID:            e.ID.String(),
			Name:          e.Name,
			Source:        e.IncomeType,
			CategoryKey:   cat.Key,
			CategoryLabel: cat.Label,
			ColorOKLCH:    cat.ColorOKLCH,
			PayDate:       e.ReceivedDate,
			AmountIdr:     e.AmountIdr,
			Note:          e.Note,
		})
		out.ManualTotalIdr += e.AmountIdr
		addToCategory(e.CategoryKey, idx, e.AmountIdr)
	}

	out.TotalIdr = out.CouponTotalIdr + out.ManualTotalIdr
	// Percentages are shares of the year's positive income only: a month
	// that nets negative has no share of what came in, and letting losses
	// into the denominator would inflate every other month's slice.
	var positiveTotal float64
	for i := range out.Months {
		m := &out.Months[i]
		m.AmountIdr = m.CouponIdr + m.ManualIdr
		if m.AmountIdr > 0 {
			positiveTotal += float64(m.AmountIdr)
		}
		sort.Slice(m.Entries, func(a, b int) bool {
			if !m.Entries[a].PayDate.Time.Equal(m.Entries[b].PayDate.Time) {
				return m.Entries[a].PayDate.Time.Before(m.Entries[b].PayDate.Time)
			}
			return m.Entries[a].Name < m.Entries[b].Name
		})
	}
	for i := range out.Months {
		if out.Months[i].AmountIdr <= 0 {
			continue
		}
		out.Months[i].Percent = percentOf(float64(out.Months[i].AmountIdr), positiveTotal)
	}

	// Same positive-only rule as the months, for the same reason: a category
	// that nets a loss over the year has no share of what came in.
	var positiveCatTotal float64
	for _, agg := range catTotals {
		if agg.amountIdr > 0 {
			positiveCatTotal += float64(agg.amountIdr)
		}
	}
	for key, agg := range catTotals {
		cat := resolveCategory(byKey, key)
		percent := 0.0
		if agg.amountIdr > 0 {
			percent = percentOf(float64(agg.amountIdr), positiveCatTotal)
		}
		out.Categories = append(out.Categories, IncomeCategoryDTO{
			CategoryKey: cat.Key,
			Label:       cat.Label,
			ColorOKLCH:  cat.ColorOKLCH,
			AmountIdr:   agg.amountIdr,
			Percent:     percent,
			Count:       agg.count,
		})
	}
	// Biggest payer first — map iteration order is random, so without this
	// the ring's slice order would shuffle between requests.
	sort.Slice(out.Categories, func(i, j int) bool {
		if out.Categories[i].AmountIdr != out.Categories[j].AmountIdr {
			return out.Categories[i].AmountIdr > out.Categories[j].AmountIdr
		}
		return out.Categories[i].Label < out.Categories[j].Label
	})

	// Stack every month's bar in that same year-level order, so a category
	// occupies the same position in every bar and the twelve rows read as
	// one chart rather than twelve unrelated ones. Only categories that
	// actually paid into the month get a slice.
	for i := range out.Months {
		for _, cat := range out.Categories {
			amount, ok := monthCatTotals[i][cat.CategoryKey]
			if !ok || amount == 0 {
				continue
			}
			out.Months[i].Slices = append(out.Months[i].Slices, IncomeMonthSliceDTO{
				CategoryKey: cat.CategoryKey,
				Label:       cat.Label,
				ColorOKLCH:  cat.ColorOKLCH,
				AmountIdr:   amount,
			})
		}
	}

	return out, nil
}

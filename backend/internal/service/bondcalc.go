package service

import (
	"sort"
	"time"

	"wealthfolio/backend/internal/domain"
)

// Everything a bond purchase "is" beyond the columns stored in
// bond_purchases is computed here: what was paid, what it cost in Rupiah,
// how big each coupon is, and which dates those coupons land on. Pure
// functions with no ctx and no repos, so SnapshotsService can reuse the
// name-grouping without taking a dependency on BondPurchasesService — and
// so the arithmetic is testable without a database.

// bondCouponsPerYear is the assumed payment frequency. Every bond in the
// ledger so far is a standard semiannual bullet. Supporting quarterly bonds
// means adding a coupon_frequency column and generalizing the /2 here and
// the +6 month offset in bondCouponMonths to 12/frequency.
const bondCouponsPerYear = 2

// bondTotalUsd is what actually left the account at settlement: the clean
// lot price plus the accrued interest owed to the seller.
func bondTotalUsd(p domain.BondPurchase) float64 {
	return p.PriceUsd + p.AccruedInterestUsd
}

// bondTotalIdr converts the settlement amount at the rate frozen on the
// purchase — deliberately not the latest rate, since this is a historical
// cost, not a current valuation.
func bondTotalIdr(p domain.BondPurchase) int64 {
	return round64(bondTotalUsd(p) * p.UsdIdrAtPurchase)
}

// bondCouponPerCycleUsd is one payment for the whole lot: face value times
// quantity gives the principal the coupon rate applies to, and the annual
// rate is split across the year's payments.
func bondCouponPerCycleUsd(p domain.BondPurchase) float64 {
	return p.FaceValueUsd * p.Quantity * (p.InterestRate / 100) / bondCouponsPerYear
}

func bondCouponPerYearUsd(p domain.BondPurchase) float64 {
	return bondCouponPerCycleUsd(p) * bondCouponsPerYear
}

// bondPricePct expresses the lot price as a percentage of par, the way a
// broker quotes it — $1,608.00 for 2 lots of $1,000 face is 80.4.
func bondPricePct(p domain.BondPurchase) float64 {
	par := p.FaceValueUsd * p.Quantity
	if par == 0 {
		return 0
	}
	return p.PriceUsd / par * 100
}

// bondCouponMonths returns the two months a bond pays in, ascending. A
// bullet bond is issued a whole number of years before it matures, so the
// maturity month is a coupon month and the other is six months away.
func bondCouponMonths(maturity domain.Date) []int {
	m1 := int(maturity.Time.Month())
	m2 := ((m1-1)+6)%12 + 1
	if m1 > m2 {
		m1, m2 = m2, m1
	}
	return []int{m1, m2}
}

// clampDayOfMonth pins a day-of-month to a month that may be shorter, so a
// bond maturing on the 31st still pays on the 28th (or 29th) of February.
func clampDayOfMonth(year int, month time.Month, day int) int {
	// Day 0 of the following month is the last day of this one.
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > last {
		return last
	}
	return day
}

// bondCouponDatesInYear returns the coupons this purchase actually pays out
// during the given calendar year, ascending.
//
// A date counts when it falls strictly after the buy date and on or before
// maturity. Strictly after, because the accrued interest paid at settlement
// is the seller's share of the current period — the first coupon the buyer
// collects is the next one. On or before maturity, because the final coupon
// is paid together with the principal; anything past maturity doesn't
// exist, which is what drops a matured bond out of the calendar.
func bondCouponDatesInYear(p domain.BondPurchase, year int) []domain.Date {
	day := p.MaturityDate.Time.Day()
	out := []domain.Date{}
	for _, m := range bondCouponMonths(p.MaturityDate) {
		month := time.Month(m)
		d := time.Date(year, month, clampDayOfMonth(year, month, day), 0, 0, 0, 0, time.UTC)
		if d.After(p.BuyDate.Time) && !d.After(p.MaturityDate.Time) {
			out = append(out, domain.NewDate(d))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out
}

// averageUsdIdr is the blended rate actually paid across a set of
// purchases — total Rupiah spent per dollar acquired.
func averageUsdIdr(totalIdr int64, totalUsd float64) float64 {
	if totalUsd == 0 {
		return 0
	}
	return float64(totalIdr) / totalUsd
}

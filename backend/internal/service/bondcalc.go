package service

import (
	"math"
	"sort"
	"time"

	"wealthfolio/backend/internal/domain"
)

// Everything a bond purchase "is" beyond the columns stored in
// bond_purchases is computed here: what was paid, what it cost in Rupiah,
// how big each coupon is, which dates those coupons land on, and what the
// position actually yields if held to maturity. Pure functions with no ctx
// and no repos, so SnapshotsService can reuse the name-grouping without
// taking a dependency on BondPurchasesService — and so the arithmetic is
// testable without a database.

// bondCouponsPerYear is the assumed payment frequency. Every bond in the
// ledger so far is a standard semiannual bullet. Supporting quarterly bonds
// means adding a coupon_frequency column and generalizing the /2 here and
// the 6-month offsets below to 12/frequency.
const bondCouponsPerYear = 2

// bondTotalUsd is the invested amount — the clean price, excluding accrued
// interest. Accrued is money advanced to the seller for coupon days you
// didn't hold, and it comes straight back at the first coupon, so it isn't
// part of what you have invested in the bond.
func bondTotalUsd(p domain.BondPurchase) float64 {
	return p.PriceUsd
}

// bondSettlementUsd is what actually left the account on settlement day:
// the clean price plus the accrued interest owed to the seller.
func bondSettlementUsd(p domain.BondPurchase) float64 {
	return p.PriceUsd + p.AccruedInterestUsd
}

// bondTotalIdr converts the invested amount at the given rate — callers
// pass the user's latest logged rate, so the ledger's Rupiah figures agree
// with the Assets page, which values every USD holding the same way.
func bondTotalIdr(p domain.BondPurchase, rate float64) int64 {
	return round64(bondTotalUsd(p) * rate)
}

// bondSettlementIdr converts the settlement amount at the given rate.
func bondSettlementIdr(p domain.BondPurchase, rate float64) int64 {
	return round64(bondSettlementUsd(p) * rate)
}

// bondCostIdrAtPurchase is the Rupiah that actually left the account on
// settlement day, at the rate frozen on the purchase. Only feeds
// averageUsdIdr — the displayed Rupiah figures use the latest rate instead.
func bondCostIdrAtPurchase(p domain.BondPurchase) int64 {
	return round64(bondSettlementUsd(p) * p.UsdIdrAtPurchase)
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

// bondFaceTotalUsd is what the issuer redeems at maturity.
func bondFaceTotalUsd(p domain.BondPurchase) float64 {
	return p.FaceValueUsd * p.Quantity
}

// bondPricePct expresses the lot price as a percentage of par, the way a
// broker quotes it — $1,608.00 for 2 lots of $1,000 face is 80.4.
func bondPricePct(p domain.BondPurchase) float64 {
	par := bondFaceTotalUsd(p)
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

// addMonthsAnchored steps a date by n months while keeping the anchor
// day-of-month, re-clamping for short months. Stepping in whole months from
// the anchor (rather than accumulating) means a schedule that passes
// through February doesn't permanently lose the 30th.
func addMonthsAnchored(t time.Time, months int, anchorDay int) time.Time {
	y, m := t.Year(), int(t.Month())+months
	for m > 12 {
		m -= 12
		y++
	}
	for m < 1 {
		m += 12
		y--
	}
	month := time.Month(m)
	return time.Date(y, month, clampDayOfMonth(y, month, anchorDay), 0, 0, 0, 0, time.UTC)
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

// bondRemainingCouponDates lists every coupon still to come, from just
// after the buy date through maturity inclusive. Walks backwards from
// maturity so the anchor day stays exact rather than drifting.
func bondRemainingCouponDates(p domain.BondPurchase) []time.Time {
	anchorDay := p.MaturityDate.Time.Day()
	step := 12 / bondCouponsPerYear

	out := []time.Time{}
	d := p.MaturityDate.Time
	// Guard the loop independently of the date maths: a 100-year bond pays
	// 200 coupons, so anything past that means bad input, not a long bond.
	for i := 0; d.After(p.BuyDate.Time) && i < 400; i++ {
		out = append(out, d)
		d = addMonthsAnchored(d, -step, anchorDay)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// bondYtmPct is the annualised yield to maturity as a percent: the discount
// rate at which the remaining coupons plus the redemption payment are worth
// exactly the clean price. It answers "what does this actually return if I
// hold it to the end", which is the number worth comparing across bonds
// bought at very different prices.
//
// Solved by bisection rather than the common shortcut
// ((C + (F-P)/n) / ((F+P)/2)), which drifts badly on deep-discount
// long-dated bonds — an INDON52 at 80% of par with 26 years to run is
// exactly where that approximation is worst.
//
// Priced clean against whole remaining periods, the standard quoting
// convention, rather than dirty against exact day counts. That deliberately
// ignores the stub between settlement and the next coupon: doing so needs
// the real coupon dates, and the accrued interest a broker reports often
// disagrees with the schedule derived from the maturity date. (One ledger
// row pays $0.47 of accrued on a $28.45 coupon — three days into a period —
// while its maturity-anchored schedule puts the next coupon a week away.
// Modelling that stub would have priced in a $28 windfall that isn't real.)
// The cost is a few basis points of precision; the benefit is a figure that
// stays comparable across rows and can't be thrown by one odd accrued cell.
func bondYtmPct(p domain.BondPurchase) float64 {
	price := bondTotalUsd(p)
	face := bondFaceTotalUsd(p)
	coupon := bondCouponPerCycleUsd(p)
	n := len(bondRemainingCouponDates(p))
	if price <= 0 || face <= 0 || n == 0 {
		return 0
	}

	// Present value of n remaining periods at annual yield y, compounded
	// semiannually per bond convention.
	pv := func(y float64) float64 {
		periodic := 1 + y/bondCouponsPerYear
		if periodic <= 0 {
			return math.Inf(1)
		}
		var sum, df float64
		df = 1
		for i := 0; i < n; i++ {
			df *= periodic
			sum += coupon / df
		}
		return sum + face/df
	}

	// pv is strictly decreasing in y, so bisection converges without needing
	// a derivative or a good initial guess.
	lo, hi := -0.9, 5.0
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		if pv(mid) > price {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2 * 100
}

// averageUsdIdr is the blended rate actually paid across a set of
// purchases — total Rupiah spent per dollar acquired. Computed on
// settlement, since accrued interest was bought with Rupiah too.
func averageUsdIdr(settlementIdr int64, settlementUsd float64) float64 {
	if settlementUsd == 0 {
		return 0
	}
	return float64(settlementIdr) / settlementUsd
}

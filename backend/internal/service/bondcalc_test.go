package service

import (
	"testing"
	"time"

	"wealthfolio/backend/internal/domain"
)

func mustDate(t *testing.T, s string) domain.Date {
	t.Helper()
	d, err := domain.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// The fixtures below are real rows from the user's bond spreadsheet, so
// these tests assert against figures a broker actually produced rather than
// numbers derived from the same formulas they're checking.

func TestBondMoneyMath(t *testing.T) {
	tests := []struct {
		name            string
		purchase        domain.BondPurchase
		wantTotalUsd    float64
		wantTotalIdr    int64
		wantCouponCycle float64
		wantPricePct    float64
	}{
		{
			// INDON36NEWNEW, bought 22 May 2026.
			name: "single lot at a premium",
			purchase: domain.BondPurchase{
				InterestRate: 5.69, Quantity: 1, FaceValueUsd: 1000,
				PriceUsd: 1014.50, AccruedInterestUsd: 0.47, UsdIdrAtPurchase: 17690,
			},
			wantTotalUsd:    1014.97,
			wantTotalIdr:    17954819,
			wantCouponCycle: 28.45,
			wantPricePct:    101.45,
		},
		{
			// PLN 48 6.15 OCBC — five lots, the row that proves the coupon
			// formula scales with quantity.
			name: "five lots near par",
			purchase: domain.BondPurchase{
				InterestRate: 6.15, Quantity: 5, FaceValueUsd: 1000,
				PriceUsd: 4925.00, AccruedInterestUsd: 34.17, UsdIdrAtPurchase: 17860,
			},
			wantTotalUsd:    4959.17,
			wantTotalIdr:    88570776,
			wantCouponCycle: 153.75,
			wantPricePct:    98.5,
		},
		{
			// INDON52 — two lots deep below par, the row that proves
			// price_usd is an aggregate and not a per-unit quote.
			name: "two lots below par",
			purchase: domain.BondPurchase{
				InterestRate: 4.30, Quantity: 2, FaceValueUsd: 1000,
				PriceUsd: 1608.00, AccruedInterestUsd: 19.59, UsdIdrAtPurchase: 17694,
			},
			wantTotalUsd:    1627.59,
			wantTotalIdr:    28798577,
			wantCouponCycle: 43.00,
			wantPricePct:    80.4,
		},
	}

	const eps = 0.005 // half a cent — these are money figures, not floats-in-general

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bondTotalUsd(tt.purchase); !nearly(got, tt.wantTotalUsd, eps) {
				t.Errorf("bondTotalUsd = %v, want %v", got, tt.wantTotalUsd)
			}
			if got := bondTotalIdr(tt.purchase); got != tt.wantTotalIdr {
				t.Errorf("bondTotalIdr = %d, want %d", got, tt.wantTotalIdr)
			}
			if got := bondCouponPerCycleUsd(tt.purchase); !nearly(got, tt.wantCouponCycle, eps) {
				t.Errorf("bondCouponPerCycleUsd = %v, want %v", got, tt.wantCouponCycle)
			}
			if got := bondCouponPerYearUsd(tt.purchase); !nearly(got, tt.wantCouponCycle*2, eps) {
				t.Errorf("bondCouponPerYearUsd = %v, want %v", got, tt.wantCouponCycle*2)
			}
			if got := bondPricePct(tt.purchase); !nearly(got, tt.wantPricePct, 0.01) {
				t.Errorf("bondPricePct = %v, want %v", got, tt.wantPricePct)
			}
		})
	}
}

func nearly(got, want, eps float64) bool {
	d := got - want
	return d < eps && d > -eps
}

func TestAverageUsdIdr(t *testing.T) {
	// The spreadsheet's first two rows: $1014.97 bought at 17690 and
	// $1138.88 bought at 17770. The blend must land between those two rates
	// and lean toward the larger row. (The sheet's own "Average IDR we buy"
	// cell reads Rp17,924 because it spans all 17 rows, not just these two.)
	got := averageUsdIdr(17954819+20237898, 1014.97+1138.88)
	if !nearly(got, 17732.30, 0.01) {
		t.Errorf("averageUsdIdr = %v, want ~17732.30", got)
	}
	if got <= 17690 || got >= 17770 {
		t.Errorf("averageUsdIdr = %v, want a blend strictly between 17690 and 17770", got)
	}
	if got := averageUsdIdr(0, 0); got != 0 {
		t.Errorf("averageUsdIdr(0,0) = %v, want 0 (no division by zero)", got)
	}
}

func TestBondCouponMonths(t *testing.T) {
	tests := []struct {
		maturity string
		want     []int
	}{
		{"2054-07-02", []int{1, 7}},  // the user's worked example
		{"2036-05-29", []int{5, 11}}, // INDON36NEWNEW
		{"2048-01-11", []int{1, 7}},
		{"2050-10-15", []int{4, 10}},
		{"2052-03-31", []int{3, 9}},
		{"2044-12-15", []int{6, 12}}, // December wraps to June, not month 18
	}
	for _, tt := range tests {
		t.Run(tt.maturity, func(t *testing.T) {
			d, err := domain.ParseDate(tt.maturity)
			if err != nil {
				t.Fatal(err)
			}
			got := bondCouponMonths(d)
			if len(got) != 2 || got[0] != tt.want[0] || got[1] != tt.want[1] {
				t.Errorf("bondCouponMonths(%s) = %v, want %v", tt.maturity, got, tt.want)
			}
		})
	}
}

func TestClampDayOfMonth(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		day   int
		want  int
	}{
		{"day fits", 2027, time.July, 31, 31},
		{"31st into a 30-day month", 2027, time.September, 31, 30},
		{"31st into February, non-leap", 2027, time.February, 31, 28},
		{"31st into February, leap", 2028, time.February, 31, 29},
		{"29th into February, non-leap", 2027, time.February, 29, 28},
		{"29th into February, leap", 2028, time.February, 29, 29},
		{"December is not out of range", 2027, time.December, 31, 31},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampDayOfMonth(tt.year, tt.month, tt.day); got != tt.want {
				t.Errorf("clampDayOfMonth(%d, %v, %d) = %d, want %d", tt.year, tt.month, tt.day, got, tt.want)
			}
		})
	}
}

func TestBondCouponDatesInYear(t *testing.T) {
	// INDON36NEWNEW: bought 22 May 2026, matures 29 May 2036, so it pays
	// every 29 May and 29 November in between.
	indon36 := domain.BondPurchase{
		BuyDate:      mustDate(t, "2026-05-22"),
		MaturityDate: mustDate(t, "2036-05-29"),
	}
	// Matures 31 Aug, so its other coupon lands on the last day of February.
	monthEnd := domain.BondPurchase{
		BuyDate:      mustDate(t, "2020-01-01"),
		MaturityDate: mustDate(t, "2040-08-31"),
	}

	tests := []struct {
		name     string
		purchase domain.BondPurchase
		year     int
		want     []string
	}{
		{
			name: "full steady-state year", purchase: indon36, year: 2027,
			want: []string{"2027-05-29", "2027-11-29"},
		},
		{
			// Bought 22 May, so the 29 May coupon still counts but the
			// preceding November of that year is in the past.
			name: "purchase year is partial", purchase: indon36, year: 2026,
			want: []string{"2026-05-29", "2026-11-29"},
		},
		{
			// The final coupon is paid with the principal on the maturity
			// date, so May counts and November does not.
			name: "maturity year drops the later coupon", purchase: indon36, year: 2036,
			want: []string{"2036-05-29"},
		},
		{
			name: "after maturity pays nothing", purchase: indon36, year: 2037,
			want: []string{},
		},
		{
			name: "before purchase pays nothing", purchase: indon36, year: 2025,
			want: []string{},
		},
		{
			name: "month-end clamps into February", purchase: monthEnd, year: 2027,
			want: []string{"2027-02-28", "2027-08-31"},
		},
		{
			name: "month-end clamps into a leap February", purchase: monthEnd, year: 2028,
			want: []string{"2028-02-29", "2028-08-31"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bondCouponDatesInYear(tt.purchase, tt.year)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d dates %v, want %d %v", len(got), datesToStrings(got), len(tt.want), tt.want)
			}
			for i, w := range tt.want {
				if got[i].String() != w {
					t.Errorf("date[%d] = %s, want %s", i, got[i].String(), w)
				}
			}
		})
	}
}

func datesToStrings(ds []domain.Date) []string {
	out := make([]string, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.String())
	}
	return out
}

func TestBondCouponDateBuyDateIsStrict(t *testing.T) {
	// Buying on a coupon date means that coupon belongs to the seller.
	p := domain.BondPurchase{
		BuyDate:      mustDate(t, "2026-05-29"),
		MaturityDate: mustDate(t, "2036-05-29"),
	}
	got := bondCouponDatesInYear(p, 2026)
	if len(got) != 1 || got[0].String() != "2026-11-29" {
		t.Errorf("got %v, want only 2026-11-29 (the 29 May coupon goes to the seller)", datesToStrings(got))
	}
}

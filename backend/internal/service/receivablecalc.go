package service

import (
	"time"

	"wealthfolio/backend/internal/domain"
)

// maxReceivableTermMonths bounds term_months the same way
// bondRemainingCouponDates caps its own walk at 400 iterations — an
// informal personal loan doesn't need more than a few decades of runway,
// and this keeps the schedule loop from being abused into an unbounded one.
const maxReceivableTermMonths = 600

// receivableIsActiveOn reports whether this loan is still paying on the
// given date: strictly after it started, and not yet past its agreed end.
func receivableIsActiveOn(loan domain.ReceivableLoan, on time.Time) bool {
	end := addMonthsAnchored(loan.StartDate.Time, loan.TermMonths, loan.StartDate.Time.Day())
	return on.After(loan.StartDate.Time) && !on.After(end)
}

// receivableMonthlyAmountIdr is the level monthly installment implied by
// spreading the total interest evenly across the term — never stored,
// always derived, so editing interest_idr or term_months can't leave a
// stale monthly amount behind.
func receivableMonthlyAmountIdr(loan domain.ReceivableLoan) int64 {
	if loan.TermMonths == 0 {
		return 0
	}
	return loan.InterestIdr / int64(loan.TermMonths)
}

// receivablePaymentDatesInYear returns every monthly installment date this
// loan pays out during the given calendar year, ascending. A date counts
// when it falls strictly after start_date and within term_months of it —
// the same "> buy, <= maturity" inclusion rule bondCouponDatesInYear uses,
// generalized from a semiannual step to monthly. The first payment lands
// the month after the loan starts (the start date itself is when the money
// went out, not a receipt); the term_months-th payment lands on the agreed
// end date, and nothing projects past it.
func receivablePaymentDatesInYear(loan domain.ReceivableLoan, year int) []domain.Date {
	anchorDay := loan.StartDate.Time.Day()
	out := []domain.Date{}
	for k := 1; k <= loan.TermMonths; k++ {
		d := addMonthsAnchored(loan.StartDate.Time, k, anchorDay)
		if d.Year() > year {
			break // dates are monotonic in k
		}
		if d.Year() == year {
			out = append(out, domain.NewDate(d))
		}
	}
	return out
}

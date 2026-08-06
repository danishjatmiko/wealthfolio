package service

import (
	"testing"

	"wealthfolio/backend/internal/domain"
)

func assertDates(t *testing.T, got []domain.Date, want []string) {
	t.Helper()
	gotStr := datesToStrings(got)
	if len(gotStr) != len(want) {
		t.Fatalf("got %v, want %v", gotStr, want)
	}
	for i := range want {
		if gotStr[i] != want[i] {
			t.Fatalf("got %v, want %v", gotStr, want)
		}
	}
}

func TestReceivablePaymentDatesInYear(t *testing.T) {
	t.Run("mid-year start splits across the year boundary", func(t *testing.T) {
		loan := domain.ReceivableLoan{
			StartDate:  mustDate(t, "2026-03-15"),
			TermMonths: 12,
		}
		// First payment is the month AFTER start (start itself is when the
		// money went out, not a receipt): Apr 2026 .. Mar 2027, 12 total.
		assertDates(t, receivablePaymentDatesInYear(loan, 2026), []string{
			"2026-04-15", "2026-05-15", "2026-06-15", "2026-07-15",
			"2026-08-15", "2026-09-15", "2026-10-15", "2026-11-15", "2026-12-15",
		})
		assertDates(t, receivablePaymentDatesInYear(loan, 2027), []string{
			"2027-01-15", "2027-02-15", "2027-03-15",
		})
		// Nothing projects past the agreed end.
		assertDates(t, receivablePaymentDatesInYear(loan, 2028), []string{})
	})

	t.Run("month-end anchor day clamps through February and April", func(t *testing.T) {
		loan := domain.ReceivableLoan{
			StartDate:  mustDate(t, "2026-01-31"),
			TermMonths: 3,
		}
		// 2026 is not a leap year: Feb has 28 days. April has 30.
		assertDates(t, receivablePaymentDatesInYear(loan, 2026), []string{
			"2026-02-28", "2026-03-31", "2026-04-30",
		})
	})

	t.Run("leap year clamps to Feb 29", func(t *testing.T) {
		loan := domain.ReceivableLoan{
			StartDate:  mustDate(t, "2027-12-31"),
			TermMonths: 2,
		}
		assertDates(t, receivablePaymentDatesInYear(loan, 2028), []string{
			"2028-01-31", "2028-02-29",
		})
	})

	t.Run("single-month term pays exactly once", func(t *testing.T) {
		loan := domain.ReceivableLoan{
			StartDate:  mustDate(t, "2026-06-01"),
			TermMonths: 1,
		}
		assertDates(t, receivablePaymentDatesInYear(loan, 2026), []string{"2026-07-01"})
	})
}

func TestReceivableIsActiveOn(t *testing.T) {
	loan := domain.ReceivableLoan{
		StartDate:  mustDate(t, "2026-03-15"),
		TermMonths: 12,
	}
	tests := []struct {
		name string
		on   string
		want bool
	}{
		{"on the start date itself is not yet active", "2026-03-15", false},
		{"the day after start is active", "2026-03-16", true},
		{"mid-term is active", "2026-09-01", true},
		{"exactly on the end date is still active", "2027-03-15", true},
		{"the day after the end date is no longer active", "2027-03-16", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := receivableIsActiveOn(loan, mustDate(t, tt.on).Time)
			if got != tt.want {
				t.Errorf("receivableIsActiveOn(%s) = %v, want %v", tt.on, got, tt.want)
			}
		})
	}
}

package service

import (
	"errors"
	"testing"
	"time"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

func TestCashflowMonthValidation(t *testing.T) {
	for _, raw := range []string{"2026-08", "2026-12", "2027-01", "9998-12"} {
		month, err := ParseCashflowMonth(raw)
		if err != nil || month.Day() != 1 {
			t.Errorf("valid month %q: %v, %v", raw, month, err)
		}
	}
	for _, raw := range []string{"", "2026-07", "2025-12", "2026-8", "2026-13", "2026-08-01", "9999-01", "2026-00", "foo"} {
		if _, err := ParseCashflowMonth(raw); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q: expected invalid input, got %v", raw, err)
		}
	}
	for _, year := range []int{0, 2025, 9999} {
		if !errors.Is(validateCashflowYear(year), ErrInvalidInput) {
			t.Errorf("accepted year %d", year)
		}
	}
}

func TestActiveIncomeValidation(t *testing.T) {
	for _, amount := range []int64{0, 25_000_001, 9007199254740991} {
		if err := (ActiveIncomeRequest{AmountIdr: &amount}).validate(); err != nil {
			t.Errorf("amount %d: %v", amount, err)
		}
	}
	for _, amount := range []int64{-1, 9007199254740992} {
		if err := (ActiveIncomeRequest{AmountIdr: &amount}).validate(); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("accepted amount %d", amount)
		}
	}
	if err := (ActiveIncomeRequest{}).validate(); !errors.Is(err, ErrInvalidInput) {
		t.Error("missing amount must not save a zero")
	}
}

func TestAggregateCashflow(t *testing.T) {
	date := func(raw string) domain.Date {
		d, err := domain.ParseDate(raw)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	out := aggregateCashflow(2026,
		[]db.MonthlyActiveIncome{
			{Month: date("2026-07-01"), AmountIdr: 999_000_000},
			{Month: date("2026-08-01"), AmountIdr: 20_000_001},
			{Month: date("2026-09-01"), AmountIdr: 0},
			{Month: date("2027-08-01"), AmountIdr: 999_000_000},
		},
		[]IncomeMonthDTO{
			{Month: 7, AmountIdr: 999_000_000},
			{Month: 8, AmountIdr: 3_000_000, CouponIdr: 1_000_000, ReceivableIdr: 500_000},
			{Month: 9, AmountIdr: -1_000_000},
		},
		[]db.ExpensePeriodAgg{
			{Period: domain.ExpensePeriod{StartDate: date("2026-07-25"), EndDate: date("2026-08-24")}, ActualTotalIdr: 10_000_000, CommittedTotalIdr: 30_000_000},
			{Period: domain.ExpensePeriod{StartDate: date("2026-06-25"), EndDate: date("2026-07-24")}, ActualTotalIdr: 999_000_000},
			{Period: domain.ExpensePeriod{EndDate: date("2027-08-24")}, ActualTotalIdr: 999_000_000},
		},
		[]domain.BigExpense{
			{ExpenseDate: date("2026-08-01"), AmountIdr: 2_000_000},
			{ExpenseDate: date("2026-08-31"), AmountIdr: 4_000_000},
			{ExpenseDate: date("2026-09-01"), AmountIdr: 2_000_000},
			{ExpenseDate: date("2026-07-31"), AmountIdr: 999_000_000},
			{ExpenseDate: date("2027-08-01"), AmountIdr: 999_000_000},
		})
	if len(out.Months) != 5 || out.Months[0].Month != 8 || out.Months[4].Month != 12 {
		t.Fatalf("expected August–December, got %+v", out)
	}
	aug, sep, oct := out.Months[0], out.Months[1], out.Months[2]
	if aug.TotalIncomeIdr != 23_000_001 || aug.TotalExpensesIdr != 16_000_000 || aug.NetIdr != 7_000_001 {
		t.Errorf("wrong August balance (actual expenses only, exact IDR): %+v", aug)
	}
	if !aug.HasExpensePeriod || aug.ProjectedPassiveIncomeIdr != 1_500_000 {
		t.Errorf("missing source metadata: %+v", aug)
	}
	if sep.ActiveIncome == nil || sep.ActiveIncome.AmountIdr != 0 || sep.NetIdr != -3_000_000 || sep.HasExpensePeriod {
		t.Errorf("zero income / negative passive income lost: %+v", sep)
	}
	if oct.ActiveIncome != nil || oct.NetIdr != 0 || oct.HasExpensePeriod {
		t.Errorf("empty month should remain distinguishable: %+v", oct)
	}
}

func TestCashflowJanuaryUsesEndMonthAcrossYears(t *testing.T) {
	start, end := boundsForPeriodMonth(2027, time.January)
	out := aggregateCashflow(2027, nil, nil,
		[]db.ExpensePeriodAgg{{Period: domain.ExpensePeriod{StartDate: start, EndDate: end}, ActualTotalIdr: 123}}, nil)
	if len(out.Months) != 12 || out.Months[0].Month != 1 || out.Months[0].NetIdr != -123 || !out.Months[0].HasExpensePeriod {
		t.Fatalf("January period misclassified: %+v", out)
	}
}

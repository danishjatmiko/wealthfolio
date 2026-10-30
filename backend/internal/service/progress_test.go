package service

import (
	"testing"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// The expense chart labels its points the way the net-equity chart does, so
// the two must label the same date identically at every granularity.
func TestProgressBucketLabelMatchesEquitySeries(t *testing.T) {
	d := mustDate(t, "2026-10-24")
	aggs := []db.SnapshotAgg{{Snapshot: domain.Snapshot{SnapshotDate: d}}}

	tests := []struct {
		granularity string
		equity      []ProgressPoint
		want        string
	}{
		{"monthly", monthlySeries(aggs), "Oct '26"},
		{"quarterly", quarterlySeries(aggs), "Q4'26"},
		{"yearly", yearlySeries(aggs), "2026"},
	}
	for _, tt := range tests {
		got := progressBucketLabel(tt.granularity, d.Time)
		if got != tt.want || got != tt.equity[0].Label {
			t.Errorf("%s: got %q, want %q (equity series says %q)", tt.granularity, got, tt.want, tt.equity[0].Label)
		}
	}
	if got := progressBucketLabel("bogus", d.Time); got != "Oct '26" {
		t.Errorf("unknown granularity: got %q, want the monthly label", got)
	}
}

func TestExpenseSeries(t *testing.T) {
	period := func(start, end string, actual, committed int64) db.ExpensePeriodAgg {
		return db.ExpensePeriodAgg{
			Period:         domain.ExpensePeriod{StartDate: mustDate(t, start), EndDate: mustDate(t, end)},
			ActualTotalIdr: actual, CommittedTotalIdr: committed,
		}
	}
	// Newest first, the way ListWithAgg returns them.
	periods := []db.ExpensePeriodAgg{
		period("2026-10-25", "2026-11-24", 10, 40),
		period("2026-09-25", "2026-10-24", 41, 39),
		period("2026-08-25", "2026-09-24", 46, 37),
		period("2026-07-25", "2026-08-24", 521, 39),
	}

	tests := []struct {
		granularity string
		want        []ExpenseProgressPoint
	}{
		{"monthly", []ExpenseProgressPoint{
			{Label: "Aug '26", Date: mustDate(t, "2026-08-24"), ActualIdr: 521, CommittedIdr: 39},
			{Label: "Sep '26", Date: mustDate(t, "2026-09-24"), ActualIdr: 46, CommittedIdr: 37},
			{Label: "Oct '26", Date: mustDate(t, "2026-10-24"), ActualIdr: 41, CommittedIdr: 39},
			{Label: "Nov '26", Date: mustDate(t, "2026-11-24"), ActualIdr: 10, CommittedIdr: 40},
		}},
		{"quarterly", []ExpenseProgressPoint{
			{Label: "Q3'26", Date: mustDate(t, "2026-09-24"), ActualIdr: 567, CommittedIdr: 76},
			{Label: "Q4'26", Date: mustDate(t, "2026-11-24"), ActualIdr: 51, CommittedIdr: 79},
		}},
		{"yearly", []ExpenseProgressPoint{
			{Label: "2026", Date: mustDate(t, "2026-11-24"), ActualIdr: 618, CommittedIdr: 155},
		}},
	}
	for _, tt := range tests {
		got := expenseSeries(periods, tt.granularity)
		if len(got) != len(tt.want) {
			t.Fatalf("%s: got %d points, want %d: %+v", tt.granularity, len(got), len(tt.want), got)
		}
		for i := range got {
			if got[i].Label != tt.want[i].Label || !got[i].Date.Time.Equal(tt.want[i].Date.Time) ||
				got[i].ActualIdr != tt.want[i].ActualIdr || got[i].CommittedIdr != tt.want[i].CommittedIdr {
				t.Errorf("%s[%d]: got %+v, want %+v", tt.granularity, i, got[i], tt.want[i])
			}
		}
	}
	if got := expenseSeries(nil, "monthly"); len(got) != 0 {
		t.Errorf("no periods: got %+v, want empty", got)
	}
}

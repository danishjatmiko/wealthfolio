package service

import (
	"testing"
	"time"

	"wealthfolio/backend/internal/domain"
)

func TestBoundsForPeriodMonth(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		month     time.Month
		wantStart string
		wantEnd   string
	}{
		{"August 2026 spans 25 Jul - 24 Aug", 2026, time.August, "2026-07-25", "2026-08-24"},
		{"July 2026 spans 25 Jun - 24 Jul", 2026, time.July, "2026-06-25", "2026-07-24"},
		{"January rolls back into the previous December", 2026, time.January, "2025-12-25", "2026-01-24"},
		{"December stays within the same year", 2026, time.December, "2026-11-25", "2026-12-24"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := boundsForPeriodMonth(tt.year, tt.month)
			if got := start.String(); got != tt.wantStart {
				t.Errorf("start = %s, want %s", got, tt.wantStart)
			}
			if got := end.String(); got != tt.wantEnd {
				t.Errorf("end = %s, want %s", got, tt.wantEnd)
			}
		})
	}
}

func TestPeriodLabel(t *testing.T) {
	end, err := time.Parse("2006-01-02", "2026-08-24")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := periodLabel(domain.NewDate(end))
	if want := "August 2026"; got != want {
		t.Errorf("periodLabel = %q, want %q", got, want)
	}
}

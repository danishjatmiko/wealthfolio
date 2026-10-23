package service

import (
	"testing"
	"time"

	"wealthfolio/backend/internal/domain"
)

func TestPickIngestionPeriod(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	period := func(start, end string, created time.Time) *domain.ExpensePeriod {
		return &domain.ExpensePeriod{
			StartDate: mustDate(t, start), EndDate: mustDate(t, end), CreatedAt: created,
		}
	}
	// The real case: October was started a month earlier, November was
	// started on 23 Oct at 17:48 WIB, a day before October ends.
	october := period("2026-09-25", "2026-10-24", time.Date(2026, 9, 24, 22, 38, 0, 0, wib))
	november := period("2026-10-25", "2026-11-24", time.Date(2026, 10, 23, 17, 48, 24, 0, wib))

	tests := []struct {
		name       string
		dateStart  string
		covering   *domain.ExpensePeriod
		latest     *domain.ExpensePeriod
		occurredAt time.Time
		want       *domain.ExpensePeriod
	}{
		{
			name:      "paid after the next period was started goes to it",
			dateStart: "2026-09-25",
			covering:  october, latest: november,
			occurredAt: time.Date(2026, 10, 23, 17, 51, 58, 0, wib),
			want:       november,
		},
		{
			name:      "paid before the next period was started keeps its own period",
			dateStart: "2026-09-25",
			covering:  october, latest: november,
			occurredAt: time.Date(2026, 10, 23, 9, 0, 0, 0, wib),
			want:       october,
		},
		{
			name:      "covering period is already the latest",
			dateStart: "2026-10-25",
			covering:  november, latest: november,
			occurredAt: time.Date(2026, 10, 30, 12, 0, 0, 0, wib),
			want:       november,
		},
		{
			name:      "a late-syncing old payment keeps its own, older period",
			dateStart: "2026-09-25",
			covering:  october, latest: november,
			occurredAt: time.Date(2026, 10, 10, 12, 0, 0, 0, wib),
			want:       october,
		},
		{
			name:      "no period for the date, but a later one started before the payment",
			dateStart: "2026-09-25",
			covering:  nil, latest: november,
			occurredAt: time.Date(2026, 10, 24, 8, 0, 0, 0, wib),
			want:       november,
		},
		{
			name:      "no period for the date and nothing newer — retry later",
			dateStart: "2026-10-25",
			covering:  nil, latest: october,
			occurredAt: time.Date(2026, 10, 26, 8, 0, 0, 0, wib),
			want:       nil,
		},
		{
			name:      "no periods at all",
			dateStart: "2026-10-25",
			covering:  nil, latest: nil,
			occurredAt: time.Date(2026, 10, 26, 8, 0, 0, 0, wib),
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickIngestionPeriod(mustDate(t, tt.dateStart), tt.covering, tt.latest, tt.occurredAt); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// Package domain contains plain data structures shared across the service,
// db, and httpapi layers.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category mirrors the categories table.
type Category struct {
	ID          int16  `json:"id"`
	Key         string `json:"key"`
	Label       string `json:"label"`
	ColorOKLCH  string `json:"color_oklch"`
	Kind        string `json:"kind"`
	PriceLinked bool   `json:"price_linked"`
	SortOrder   int16  `json:"sort_order"`
}

// RateEntry mirrors the rate_entries table.
//
// NOTE on units: Antam/Kinghalim/Ubs are Rp-per-gram expressed in THOUSANDS
// of IDR (matching the general *_idr unit convention). UsdIdr is the sole
// exception in the schema: it is full IDR per 1 USD, not thousands.
type RateEntry struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	EntryDate Date      `json:"entry_date"`
	Antam     float64   `json:"antam"`
	Kinghalim float64   `json:"kinghalim"`
	Ubs       float64   `json:"ubs"`
	UsdIdr    float64   `json:"usd_idr"`
	CreatedAt time.Time `json:"created_at"`
}

// Snapshot mirrors the snapshots table.
type Snapshot struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"-"`
	SnapshotDate Date      `json:"snapshot_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// Holding mirrors the holdings table, joined with its category for
// convenience fields (category_key / category_label).
type Holding struct {
	ID            uuid.UUID `json:"id"`
	SnapshotID    uuid.UUID `json:"snapshot_id"`
	CategoryID    int16     `json:"category_id"`
	CategoryKey   string    `json:"category_key"`
	CategoryLabel string    `json:"category_label"`
	Name          string    `json:"name"`
	Detail        string    `json:"detail"`
	ValueIdr      int64     `json:"value_idr"`
	IsLiability   bool      `json:"is_liability"`
	Gram          *float64  `json:"gram"`
	Qty           *float64  `json:"qty"`
	Brand         *string   `json:"brand"`
	UsdValue      *float64  `json:"usd_value"`
	Currency      *string   `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DebtSnapshot mirrors the debt_snapshots table.
type DebtSnapshot struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"-"`
	SnapshotDate Date      `json:"snapshot_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// DebtEntry mirrors the debt_entries table.
type DebtEntry struct {
	ID             uuid.UUID `json:"id"`
	DebtSnapshotID uuid.UUID `json:"debt_snapshot_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	ValueIdr       int64     `json:"value_idr"`
	Direction      string    `json:"direction"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PassiveIncomeSource mirrors the passive_income_sources table.
type PassiveIncomeSource struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"-"`
	CategoryID    int16     `json:"category_id"`
	CategoryKey   string    `json:"category_key"`
	CategoryLabel string    `json:"category_label"`
	Name          string    `json:"name"`
	PerYearIdr    int64     `json:"per_year_idr"`
}

// Target mirrors the targets table plus the server-computed fields.
type Target struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"-"`
	Name               string    `json:"name"`
	Year               int       `json:"year"`
	MetricType         string    `json:"metric_type"`
	TargetValue        float64   `json:"target_value"`
	Unit               string    `json:"unit"`
	ManualCurrentValue *float64  `json:"-"`
	CurrentValue       float64   `json:"current_value"`
	Percent            float64   `json:"percent"`
	LowerIsBetter      bool      `json:"lower_is_better"`
}

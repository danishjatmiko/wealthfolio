// Package domain contains plain data structures shared across the service,
// db, and httpapi layers.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User mirrors the users table. Email/AvatarURL/GoogleSub are nullable at
// the column level only for the pre-auth seed user before it's claimed by
// the first Google login (see db.UsersRepo.ClaimSeedUser); every user
// reachable through a session has all three set. PasswordHash is nullable
// for everyone except the one account currently opted into email/password
// sign-in (see migrations/00006_password_auth.sql) — never serialized.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        *string   `json:"-"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    *string   `json:"-"`
	GoogleSub    *string   `json:"-"`
	PasswordHash *string   `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

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

// ExpensePeriod mirrors the expense_periods table: a custom pay-cycle
// window (25th of one month through the 24th of the next) rather than a
// calendar month. Never locks — every period stays editable indefinitely.
type ExpensePeriod struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	StartDate Date      `json:"start_date"`
	EndDate   Date      `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
}

// ExpenseCategory mirrors the expense_categories table: a free-form,
// user-created grouping for budget envelopes (1 envelope has 1 category;
// 1 category can have many envelopes).
type ExpenseCategory struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// BudgetEnvelope mirrors the budget_envelopes table: a monthly spending
// target that bundles multiple FixedExpense rows. CategoryName is joined
// for convenience, same pattern as Holding.CategoryKey/CategoryLabel.
type BudgetEnvelope struct {
	ID                 uuid.UUID `json:"id"`
	PeriodID           uuid.UUID `json:"period_id"`
	CategoryID         uuid.UUID `json:"category_id"`
	CategoryName       string    `json:"category_name"`
	Name               string    `json:"name"`
	CommittedAmountIdr int64     `json:"committed_amount_idr"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FixedExpense mirrors the fixed_expenses table. Every fixed expense
// belongs to a BudgetEnvelope — there's no more "standalone" expense.
type FixedExpense struct {
	ID         uuid.UUID `json:"id"`
	PeriodID   uuid.UUID `json:"period_id"`
	EnvelopeID uuid.UUID `json:"envelope_id"`
	Name       string    `json:"name"`
	AmountIdr  int64     `json:"amount_idr"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

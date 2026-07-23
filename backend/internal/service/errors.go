package service

import "errors"

// Sentinel errors surfaced by the service layer. httpapi handlers map these
// to specific HTTP status codes via errors.Is; their Error() text is used
// verbatim in the JSON error body where the API spec requires an exact
// message.
var (
	// ErrNoRateEntry: a rate-linked holding value was requested but the
	// user has no rate_entries row yet and gave no manual value_idr
	// fallback. Maps to HTTP 422.
	ErrNoRateEntry = errors.New("no rate entry available — add one on the Rates page first")

	// ErrSnapshotLocked: a holdings write was attempted against a
	// snapshot that isn't the user's latest. Maps to HTTP 409.
	ErrSnapshotLocked = errors.New("snapshot is locked; only the latest snapshot is editable")

	// ErrSnapshotDateExists: a snapshot already exists for the requested
	// date. Maps to HTTP 409.
	ErrSnapshotDateExists = errors.New("a snapshot already exists for this date")

	// ErrSnapshotDateInPast: an attempt to create a snapshot dated before
	// today. New snapshots may only be created today or in the future.
	// Maps to HTTP 400.
	ErrSnapshotDateInPast = errors.New("snapshot date must be today or later")

	// ErrPeriodMonthExists: an expense period already exists for the
	// requested year/month. Maps to HTTP 409.
	ErrPeriodMonthExists = errors.New("a period already exists for this month")

	// ErrInvalidCategory: the category_id on a holdings/passive-income
	// write doesn't exist. Maps to HTTP 400.
	ErrInvalidCategory = errors.New("invalid category_id")

	// ErrNoActivePeriod: a notification-expense ingestion landed on a date
	// with no expense period covering it yet. Retryable — resolves once
	// the user starts that period as usual. Maps to HTTP 422.
	ErrNoActivePeriod = errors.New("no expense period covers this date yet")

	// ErrNoSourceMapping: a notification-expense ingestion arrived for a
	// source with no configured envelope mapping. Maps to HTTP 422.
	ErrNoSourceMapping = errors.New("no envelope mapping configured for this source")

	// ErrEnvelopeNotFound: a notification-expense ingestion's mapped
	// envelope name doesn't exist in the current period. Maps to HTTP 422.
	ErrEnvelopeNotFound = errors.New("mapped envelope not found in the current period")

	// ErrInvalidInput: generic request validation failure. Maps to HTTP
	// 400; wrap with fmt.Errorf("%w: ...", ErrInvalidInput) for a more
	// specific message.
	ErrInvalidInput = errors.New("invalid input")
)

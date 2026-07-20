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
	// date (snapshots can be created for any date — past, to backfill
	// history, or future — as long as it isn't a duplicate). Maps to
	// HTTP 409.
	ErrSnapshotDateExists = errors.New("a snapshot already exists for this date")

	// ErrInvalidCategory: the category_id on a holdings/passive-income
	// write doesn't exist. Maps to HTTP 400.
	ErrInvalidCategory = errors.New("invalid category_id")

	// ErrInvalidInput: generic request validation failure. Maps to HTTP
	// 400; wrap with fmt.Errorf("%w: ...", ErrInvalidInput) for a more
	// specific message.
	ErrInvalidInput = errors.New("invalid input")
)

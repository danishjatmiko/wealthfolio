package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// NotificationExpenseEventsRepo manages notification_expense_events rows —
// the idempotent record of every notification-ingestion attempt.
type NotificationExpenseEventsRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationExpenseEventsRepo(pool *pgxpool.Pool) *NotificationExpenseEventsRepo {
	return &NotificationExpenseEventsRepo{pool: pool}
}

const notificationExpenseEventSelectCols = `id, user_id, idempotency_key, source, raw_title, raw_text, raw_big_text, occurred_at, parse_status, amount_idr, merchant_name, envelope_id, fixed_expense_id, created_at`

func scanNotificationExpenseEvent(row interface{ Scan(dest ...any) error }) (domain.NotificationExpenseEvent, error) {
	var e domain.NotificationExpenseEvent
	err := row.Scan(&e.ID, &e.UserID, &e.IdempotencyKey, &e.Source, &e.RawTitle, &e.RawText, &e.RawBigText,
		&e.OccurredAt, &e.ParseStatus, &e.AmountIdr, &e.MerchantName, &e.EnvelopeID, &e.FixedExpenseID, &e.CreatedAt)
	if err != nil {
		return domain.NotificationExpenseEvent{}, err
	}
	return e, nil
}

// GetByIdempotencyKey returns the stored outcome of a prior ingestion
// attempt for this key, if any. ErrNotFound if this key hasn't been seen
// before.
func (r *NotificationExpenseEventsRepo) GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (domain.NotificationExpenseEvent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+notificationExpenseEventSelectCols+`
		FROM notification_expense_events WHERE user_id = $1 AND idempotency_key = $2`, userID, key)
	e, err := scanNotificationExpenseEvent(row)
	if err != nil {
		return domain.NotificationExpenseEvent{}, wrapNotFound(err)
	}
	return e, nil
}

// RawFields are the notification's raw, unparsed fields — common to both
// a "created" and "ignored" outcome.
type RawFields struct {
	UserID         uuid.UUID
	IdempotencyKey string
	Source         string
	RawTitle       *string
	RawText        *string
	RawBigText     *string
	OccurredAt     time.Time
}

// CreateIgnored records a parse failure: the notification didn't match any
// known transaction pattern for its source, so no fixed_expense is
// created. Idempotent on (user_id, idempotency_key) — a retry of an
// already-ignored key returns the original row rather than erroring.
func (r *NotificationExpenseEventsRepo) CreateIgnored(ctx context.Context, f RawFields) (domain.NotificationExpenseEvent, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notification_expense_events
			(user_id, idempotency_key, source, raw_title, raw_text, raw_big_text, occurred_at, parse_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'ignored')
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING `+notificationExpenseEventSelectCols,
		f.UserID, f.IdempotencyKey, f.Source, f.RawTitle, f.RawText, f.RawBigText, f.OccurredAt)
	event, err := scanNotificationExpenseEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.GetByIdempotencyKey(ctx, f.UserID, f.IdempotencyKey)
		}
		return domain.NotificationExpenseEvent{}, err
	}
	return event, nil
}

// CreateExpenseWrite is everything needed to create a fixed_expense plus
// its audit event for a successfully-parsed notification.
type CreateExpenseWrite struct {
	RawFields
	PeriodID   uuid.UUID
	EnvelopeID uuid.UUID
	AmountIdr  int64
	Merchant   string
}

// CreateExpense inserts a fixed_expenses row and its notification_expense_
// events audit row (parse_status='created') in a single transaction, or —
// if (user_id, idempotency_key) was already processed, including by a
// concurrent request that raced this one — returns that existing event
// untouched, with no duplicate fixed_expense left behind.
func (r *NotificationExpenseEventsRepo) CreateExpense(ctx context.Context, w CreateExpenseWrite) (domain.NotificationExpenseEvent, error) {
	if existing, err := r.GetByIdempotencyKey(ctx, w.UserID, w.IdempotencyKey); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return domain.NotificationExpenseEvent{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.NotificationExpenseEvent{}, err
	}
	defer tx.Rollback(ctx)

	var fixedExpenseID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO fixed_expenses (period_id, envelope_id, name, amount_idr)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		w.PeriodID, w.EnvelopeID, w.Merchant, w.AmountIdr).Scan(&fixedExpenseID)
	if err != nil {
		return domain.NotificationExpenseEvent{}, err
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO notification_expense_events
			(user_id, idempotency_key, source, raw_title, raw_text, raw_big_text, occurred_at,
			 parse_status, amount_idr, merchant_name, envelope_id, fixed_expense_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'created', $8, $9, $10, $11)
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING `+notificationExpenseEventSelectCols,
		w.UserID, w.IdempotencyKey, w.Source, w.RawTitle, w.RawText, w.RawBigText, w.OccurredAt,
		w.AmountIdr, w.Merchant, w.EnvelopeID, fixedExpenseID)
	event, err := scanNotificationExpenseEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// A concurrent request recorded this key first — our
			// fixed_expense insert above rolls back via the deferred
			// Rollback, and we hand back the winner's result instead.
			return r.GetByIdempotencyKey(ctx, w.UserID, w.IdempotencyKey)
		}
		return domain.NotificationExpenseEvent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.NotificationExpenseEvent{}, err
	}
	return event, nil
}

-- +goose Up
-- expense_source_mappings: per-user default envelope for each notification
-- source (GoPay/DANA/BCA), configured once in the Android app's Settings
-- screen. Stored by envelope *name*, not id — envelopes are period-scoped
-- (a new period's envelopes get fresh ids via CopyFromPeriod), so the
-- mapping is resolved against whichever period is current at ingestion
-- time rather than pinned to one period's row.
CREATE TABLE expense_source_mappings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    source text NOT NULL,
    envelope_name text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, source)
);

-- notification_expense_events: one row per notification-ingestion attempt,
-- keyed by a client-generated idempotency_key so retries never duplicate.
-- Doubles as the audit trail for created expenses and as the sample data
-- the notificationparse parsers get developed against (raw_* columns are
-- kept even when parse_status = 'ignored').
CREATE TABLE notification_expense_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    idempotency_key text NOT NULL,
    source text NOT NULL,
    raw_title text,
    raw_text text,
    raw_big_text text,
    occurred_at timestamptz NOT NULL,
    parse_status text NOT NULL,
    amount_idr bigint,
    merchant_name text,
    envelope_id uuid REFERENCES budget_envelopes (id),
    fixed_expense_id uuid REFERENCES fixed_expenses (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);
CREATE INDEX idx_notification_expense_events_user ON notification_expense_events (user_id);

-- +goose Down
DROP TABLE notification_expense_events;
DROP TABLE expense_source_mappings;

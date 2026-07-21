-- +goose Up
-- Replaces the flat, unversioned `debts` table with a snapshot/entry model
-- identical in spirit to snapshots/holdings: one immutable snapshot per
-- date, only the latest (by snapshot_date) editable. Debt snapshots run on
-- their own independent timeline, not tied to asset snapshot dates.
CREATE TABLE debt_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    snapshot_date date NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, snapshot_date)
);
CREATE INDEX idx_debt_snapshots_user_date ON debt_snapshots (user_id, snapshot_date DESC);

CREATE TABLE debt_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    debt_snapshot_id uuid NOT NULL REFERENCES debt_snapshots (id) ON DELETE CASCADE,
    name text NOT NULL,
    type text NOT NULL DEFAULT '',
    value_idr bigint NOT NULL,
    direction text NOT NULL CHECK (direction IN ('i_owe', 'owed_to_me')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_debt_entries_snapshot ON debt_entries (debt_snapshot_id);

-- Migrate any existing debts rows into one new snapshot dated today. This is
-- a no-op on any database where `debts` is already empty.
INSERT INTO debt_snapshots (user_id, snapshot_date)
SELECT DISTINCT user_id, CURRENT_DATE FROM debts
ON CONFLICT (user_id, snapshot_date) DO NOTHING;

INSERT INTO debt_entries (debt_snapshot_id, name, type, value_idr, direction, created_at, updated_at)
SELECT ds.id, d.name, d.type, d.value_idr, d.direction, d.created_at, d.updated_at
FROM debts d
JOIN debt_snapshots ds ON ds.user_id = d.user_id AND ds.snapshot_date = CURRENT_DATE;

DROP TABLE debts;

-- +goose Down
CREATE TABLE debts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    type text NOT NULL DEFAULT '',
    value_idr bigint NOT NULL,
    direction text NOT NULL CHECK (direction IN ('i_owe', 'owed_to_me')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_debts_user ON debts (user_id);

DROP TABLE IF EXISTS debt_entries;
DROP TABLE IF EXISTS debt_snapshots;

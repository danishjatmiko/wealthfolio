-- +goose Up
-- UNIT CONVENTION (applies to every *_idr / value_idr / target_value / manual_current_value
-- column below): the integer stored is IDR in THOUSANDS, e.g. 3750000 means Rp 3,750,000,000.
-- This matches the design prototype's internal `val` unit exactly (see design_handoff_portfolio_app
-- README "Currency formatting" section), so the money()/goldPrice()/mdComputedVal() formulas can be
-- ported to Go/TS verbatim with no unit-conversion factor. rate_entries.usd_idr is the ONLY exception:
-- it is full IDR per 1 USD (not thousands), also matching the prototype.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text,
    display_name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE categories (
    id smallserial PRIMARY KEY,
    key text NOT NULL UNIQUE,
    label text NOT NULL,
    color_oklch text NOT NULL,
    kind text NOT NULL CHECK (kind IN ('asset', 'liability')),
    price_linked boolean NOT NULL DEFAULT false,
    sort_order smallint NOT NULL DEFAULT 0
);

CREATE TABLE rate_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    entry_date date NOT NULL,
    antam numeric(14, 2) NOT NULL,
    kinghalim numeric(14, 2) NOT NULL,
    ubs numeric(14, 2) NOT NULL,
    usd_idr numeric(14, 2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, entry_date)
);
CREATE INDEX idx_rate_entries_user_date ON rate_entries (user_id, entry_date DESC);

CREATE TABLE snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    snapshot_date date NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, snapshot_date)
);
CREATE INDEX idx_snapshots_user_date ON snapshots (user_id, snapshot_date DESC);

CREATE TABLE holdings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id uuid NOT NULL REFERENCES snapshots (id) ON DELETE CASCADE,
    category_id smallint NOT NULL REFERENCES categories (id),
    name text NOT NULL,
    detail text NOT NULL DEFAULT '',
    value_idr bigint NOT NULL,
    is_liability boolean NOT NULL DEFAULT false,
    gram numeric(14, 3),
    qty numeric(14, 3),
    brand text,
    usd_value numeric(14, 2),
    currency text CHECK (currency IN ('IDR', 'USD')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_holdings_snapshot ON holdings (snapshot_id);

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

CREATE TABLE passive_income_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id smallint NOT NULL REFERENCES categories (id),
    name text NOT NULL DEFAULT '',
    per_year_idr bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_passive_income_user ON passive_income_sources (user_id);

CREATE TABLE targets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    year int NOT NULL,
    metric_type text NOT NULL CHECK (metric_type IN ('equity', 'gold_grams', 'passive_income', 'debt_ratio', 'custom')),
    target_value numeric(18, 2) NOT NULL,
    unit text NOT NULL DEFAULT '',
    manual_current_value numeric(18, 2),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_targets_user ON targets (user_id);

-- +goose Down
DROP TABLE IF EXISTS targets;
DROP TABLE IF EXISTS passive_income_sources;
DROP TABLE IF EXISTS debts;
DROP TABLE IF EXISTS holdings;
DROP TABLE IF EXISTS snapshots;
DROP TABLE IF EXISTS rate_entries;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

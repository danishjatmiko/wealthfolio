-- +goose Up
-- Monthly expense tracking, organized around a custom pay-cycle "period"
-- (25th of one month through the 24th of the next, named after the month
-- it ends in) rather than calendar months. Periods never lock — unlike
-- snapshots/debt_snapshots, this is an ongoing log, not a point-in-time
-- record, so every period stays editable indefinitely.
CREATE TABLE expense_periods (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    start_date date NOT NULL,
    end_date date NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, start_date)
);
CREATE INDEX idx_expense_periods_user_start ON expense_periods (user_id, start_date DESC);

-- A budget envelope is a monthly target (e.g. "Kebutuhan Keluarga Inti")
-- that bundles multiple real fixed_expenses; its realized total (sum of
-- its children) is compared against committed_amount_idr to show
-- over/under, but that target itself never counts toward the dashboard
-- summary — only fixed_expenses do.
CREATE TABLE budget_envelopes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id uuid NOT NULL REFERENCES expense_periods (id) ON DELETE CASCADE,
    name text NOT NULL,
    committed_amount_idr bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_budget_envelopes_period ON budget_envelopes (period_id);

-- envelope_id is nullable: NULL means a standalone fixed expense not
-- bundled under any budget envelope (e.g. a fixed monthly obligation with
-- no need for budget-vs-actual tracking).
CREATE TABLE fixed_expenses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id uuid NOT NULL REFERENCES expense_periods (id) ON DELETE CASCADE,
    envelope_id uuid REFERENCES budget_envelopes (id) ON DELETE CASCADE,
    name text NOT NULL,
    amount_idr bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_fixed_expenses_period ON fixed_expenses (period_id);
CREATE INDEX idx_fixed_expenses_envelope ON fixed_expenses (envelope_id);

-- +goose Down
DROP TABLE IF EXISTS fixed_expenses;
DROP TABLE IF EXISTS budget_envelopes;
DROP TABLE IF EXISTS expense_periods;

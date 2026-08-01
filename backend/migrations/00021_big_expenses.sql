-- +goose Up
-- Big Expense: a manually-kept ledger of large one-off purchases through
-- the year (a bond, a laptop, a hotel stay) — separate from the recurring
-- monthly envelope budget in fixed_expenses. Permanent and user-scoped,
-- never copies forward and never locks — same family as bond_purchases,
-- debts, and passive_income_sources.
CREATE TABLE big_expenses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    amount_idr bigint NOT NULL,
    expense_date date NOT NULL,
    -- Free text, no CHECK: same reasoning as bond_purchases.platform — the
    -- frontend offers a curated list (Travel, Medical, ...) plus "Other",
    -- but nothing in Go branches on a closed set of category values.
    category text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_big_expenses_user_date ON big_expenses (user_id, expense_date DESC);

-- +goose Down
DROP TABLE IF EXISTS big_expenses;

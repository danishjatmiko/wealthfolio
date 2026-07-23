-- +goose Up
-- Categories group budget envelopes for reporting (1 envelope has 1
-- category; 1 category can have many envelopes). Free-form and user-
-- created — unlike the fixed, seeded Asset categories table, there's no
-- predefined taxonomy here.
CREATE TABLE expense_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);
CREATE INDEX idx_expense_categories_user ON expense_categories (user_id);

-- Backfill: every existing envelope needs a category. Seed a "General"
-- category per user who already has envelopes and point every existing
-- envelope at it.
INSERT INTO expense_categories (user_id, name)
SELECT DISTINCT p.user_id, 'General'
FROM budget_envelopes be
JOIN expense_periods p ON p.id = be.period_id;

ALTER TABLE budget_envelopes ADD COLUMN category_id uuid REFERENCES expense_categories (id);

UPDATE budget_envelopes be
SET category_id = ec.id
FROM expense_periods p
JOIN expense_categories ec ON ec.user_id = p.user_id AND ec.name = 'General'
WHERE be.period_id = p.id;

ALTER TABLE budget_envelopes ALTER COLUMN category_id SET NOT NULL;
CREATE INDEX idx_budget_envelopes_category ON budget_envelopes (category_id);

-- Backfill: every standalone (envelope_id IS NULL) fixed expense needs an
-- envelope now. Seed an "Others" category + one "Others" envelope per
-- period that has any standalone expenses, and reassign those expenses to
-- it, so no historical data is lost.
INSERT INTO expense_categories (user_id, name)
SELECT DISTINCT p.user_id, 'Others'
FROM fixed_expenses fe
JOIN expense_periods p ON p.id = fe.period_id
WHERE fe.envelope_id IS NULL
ON CONFLICT (user_id, name) DO NOTHING;

INSERT INTO budget_envelopes (period_id, category_id, name, committed_amount_idr)
SELECT DISTINCT fe.period_id, ec.id, 'Others', 0
FROM fixed_expenses fe
JOIN expense_periods p ON p.id = fe.period_id
JOIN expense_categories ec ON ec.user_id = p.user_id AND ec.name = 'Others'
WHERE fe.envelope_id IS NULL;

UPDATE fixed_expenses fe
SET envelope_id = be.id
FROM budget_envelopes be
WHERE fe.envelope_id IS NULL AND be.period_id = fe.period_id AND be.name = 'Others';

ALTER TABLE fixed_expenses ALTER COLUMN envelope_id SET NOT NULL;

-- +goose Down
ALTER TABLE fixed_expenses ALTER COLUMN envelope_id DROP NOT NULL;
ALTER TABLE budget_envelopes DROP COLUMN category_id;
DROP TABLE expense_categories;

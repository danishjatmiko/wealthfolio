-- +goose Up
-- Categories turned out not to earn their keep as a layer above budget
-- envelopes — every envelope stands on its own now, nothing groups them.
DROP INDEX IF EXISTS idx_budget_envelopes_category;
ALTER TABLE budget_envelopes DROP COLUMN category_id;
DROP TABLE expense_categories;

-- +goose Down
CREATE TABLE expense_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);
CREATE INDEX idx_expense_categories_user ON expense_categories (user_id);

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

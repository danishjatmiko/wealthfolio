-- +goose Up
-- The date a fixed expense was actually paid — editable, so an expense
-- logged a few days late can still carry the day it happened rather than
-- the day it was typed in. Backfilled from the capturing notification's
-- occurred_at where there is one, else from created_at, both read as
-- Jakarta-local calendar days.
ALTER TABLE fixed_expenses ADD COLUMN expense_date date;

UPDATE fixed_expenses fe
SET expense_date = COALESCE(
    (SELECT (ne.occurred_at AT TIME ZONE 'Asia/Jakarta')::date
     FROM notification_expense_events ne
     WHERE ne.fixed_expense_id = fe.id
     LIMIT 1),
    (fe.created_at AT TIME ZONE 'Asia/Jakarta')::date
);

ALTER TABLE fixed_expenses ALTER COLUMN expense_date SET NOT NULL;

-- +goose Down
ALTER TABLE fixed_expenses DROP COLUMN expense_date;

-- +goose Up
-- Passive income stops being a set of standing per-year estimates and
-- becomes a dated ledger: one row per payment actually received, the same
-- shape as big_expenses. That's what makes it mergeable with the bond
-- coupon calendar on the Passive Income page — a coupon lands on a date, so
-- a dividend has to as well, or the two can't share a month bucket.
--
-- The "sources" name goes with the old semantics; a row is now an entry.
ALTER TABLE passive_income_sources RENAME TO passive_income_entries;
ALTER TABLE passive_income_entries RENAME COLUMN per_year_idr TO amount_idr;

-- Existing rows were annual estimates with no date of their own. Their
-- created_at is the only date they carry, so it's what they get — the
-- amount survives intact and stays editable, which beats dropping them.
ALTER TABLE passive_income_entries ADD COLUMN received_date date;
UPDATE passive_income_entries SET received_date = created_at::date;
ALTER TABLE passive_income_entries ALTER COLUMN received_date SET NOT NULL;

-- Free text, no CHECK: same reasoning as big_expenses.category — the
-- frontend offers a curated list (Dividen, Capital, IPO, ...) plus "Other",
-- but nothing in Go branches on a closed set of values. Orthogonal to
-- category_id, which stays the asset class the income came from (Saham,
-- Reksa Dana): a dividend and a realized capital gain are both Saham.
ALTER TABLE passive_income_entries ADD COLUMN income_type text NOT NULL DEFAULT '';

-- The source spreadsheet annotates rows with why a figure came out the way
-- it did ("cut loss", "harusnya dapet 8jtan"). Without somewhere to put it
-- that context is lost on import.
ALTER TABLE passive_income_entries ADD COLUMN note text NOT NULL DEFAULT '';

ALTER INDEX idx_passive_income_user RENAME TO idx_passive_income_entries_user;
CREATE INDEX idx_passive_income_entries_user_date
    ON passive_income_entries (user_id, received_date DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_passive_income_entries_user_date;
ALTER INDEX idx_passive_income_entries_user RENAME TO idx_passive_income_user;
ALTER TABLE passive_income_entries DROP COLUMN note;
ALTER TABLE passive_income_entries DROP COLUMN income_type;
ALTER TABLE passive_income_entries DROP COLUMN received_date;
ALTER TABLE passive_income_entries RENAME COLUMN amount_idr TO per_year_idr;
ALTER TABLE passive_income_entries RENAME TO passive_income_sources;

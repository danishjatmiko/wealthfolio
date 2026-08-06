-- +goose Up
-- Loan terms behind a receivable — how long it runs, at what rate, and how
-- much comes in each month. Deliberately NOT modeled on debt_entries: a
-- loan's terms are a one-time historical fact set when the money went out,
-- not a monthly re-statement, so — same reasoning as bond_purchases
-- (migration 00019) — this must never copy forward and never lock.
--
-- Linked to debt_entries by matching borrower_name to that row's name,
-- cosmetically only, no foreign key — identical to how bond_purchases.
-- bond_name links to Assets holdings with no FK back to any holding row.
-- Passive Income's calendar reads this table directly; it never touches
-- debt_entries.
--
-- interest_idr is the only amount reported here — the principal itself
-- already lives on the matching debt_entries row as its balance, so it's
-- not duplicated. interest_idr is what was actually agreed, reported
-- directly rather than derived, the same "record what was actually
-- reported" rule bond_purchases.price_usd follows. The monthly amount that
-- feeds Passive Income is never stored: it's interest_idr / term_months —
-- the total interest spread evenly across the term — computed at read time
-- so an edit can't leave a stale monthly amount behind.
CREATE TABLE receivable_loans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    borrower_name text NOT NULL,
    start_date date NOT NULL,
    -- Fixed duration set once from start_date, like a bond's maturity date —
    -- not a countdown the user decrements as payments come in.
    term_months smallint NOT NULL,
    interest_idr bigint NOT NULL,
    note text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- user_id leads for plain per-user lookups; borrower_name second because
-- DebtModal looks a loan up by name to show/edit its terms.
CREATE INDEX idx_receivable_loans_user_name ON receivable_loans (user_id, borrower_name);

-- +goose Down
DROP TABLE IF EXISTS receivable_loans;

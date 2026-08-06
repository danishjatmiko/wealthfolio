-- +goose Up
-- remaining_debt_idr tracks how much principal is still owed on a
-- receivable — purely informational, shown next to the interest-derived
-- monthly amount but never fed into any Passive Income calculation
-- (interest_idr / term_months stays the only figure that does that). It's
-- expected to be updated by hand as the borrower pays the principal down,
-- independent of the fixed start_date/term_months/interest_idr agreement.
ALTER TABLE receivable_loans ADD COLUMN remaining_debt_idr bigint NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE receivable_loans DROP COLUMN remaining_debt_idr;

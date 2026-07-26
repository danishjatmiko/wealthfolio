-- +goose Up
-- Removes the "IDR in THOUSANDS" storage convention documented at the top
-- of 00001_init.sql. From this migration forward, every *_idr / value_idr
-- bigint column (plus rate_entries.antam/kinghalim/ubs, and targets rows
-- typed equity/passive_income) store RAW, WHOLE RUPIAH — "1 = Rp 1", not
-- "1 = Rp 1,000". No cents/decimal support is introduced; this migration
-- only rescales existing data x1000 so every already-stored amount
-- continues to display as the exact same Rupiah figure it did before,
-- just re-expressed in the new unit. Unaffected: holdings.usd_value,
-- rate_entries.usd_idr (already full IDR/foreign-currency units), targets
-- rows with metric_type IN ('gold_grams', 'debt_ratio', 'custom'), and
-- targets.manual_current_value (never IDR-typed for the rescaled metric
-- types). Historical migration headers describing the old convention are
-- left as-is.

UPDATE rate_entries SET
    antam     = antam * 1000,
    kinghalim = kinghalim * 1000,
    ubs       = ubs * 1000;

UPDATE holdings SET value_idr = value_idr * 1000;

UPDATE debt_entries SET value_idr = value_idr * 1000;

UPDATE passive_income_sources SET per_year_idr = per_year_idr * 1000;

UPDATE budget_envelopes SET committed_amount_idr = committed_amount_idr * 1000;

UPDATE fixed_expenses SET amount_idr = amount_idr * 1000;

-- amount_idr is nullable; multiplying NULL stays NULL, no WHERE needed.
UPDATE notification_expense_events SET amount_idr = amount_idr * 1000;

-- Only equity/passive_income targets are IDR-denominated in the old
-- "thousands" sense (see service/targets.go's computeCurrentValue and
-- service/dashboard.go's passive-income handling) — gold_grams/debt_ratio/
-- custom must NOT be rescaled.
UPDATE targets SET target_value = target_value * 1000
WHERE metric_type IN ('equity', 'passive_income');

-- +goose Down
-- Only safe to run immediately after Up, before any new data is entered
-- under the new (raw-Rupiah) convention — integer division here is lossy
-- for any post-migration write that isn't an exact multiple of 1000.
UPDATE targets SET target_value = target_value / 1000
WHERE metric_type IN ('equity', 'passive_income');

UPDATE notification_expense_events SET amount_idr = amount_idr / 1000;

UPDATE fixed_expenses SET amount_idr = amount_idr / 1000;

UPDATE budget_envelopes SET committed_amount_idr = committed_amount_idr / 1000;

UPDATE passive_income_sources SET per_year_idr = per_year_idr / 1000;

UPDATE debt_entries SET value_idr = value_idr / 1000;

UPDATE holdings SET value_idr = value_idr / 1000;

UPDATE rate_entries SET
    antam     = antam / 1000,
    kinghalim = kinghalim / 1000,
    ubs       = ubs / 1000;

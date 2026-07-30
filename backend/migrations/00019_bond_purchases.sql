-- +goose Up
-- Individual USD bond purchases — the permanent ledger behind what used to
-- be a single hand-typed `bonds_usd` holding. Deliberately NOT modeled on
-- snapshots/holdings: a purchase is a one-time historical fact, not a
-- monthly re-statement, so it must never copy forward and never lock. The
-- structural precedent is expense_periods / passive_income_sources —
-- user-scoped, permanent, hard-deleted.
--
-- STORED vs DERIVED: only the facts a broker statement actually reports are
-- stored here. total_usd (= price + accrued), total_idr (= round(total_usd
-- * usd_idr_at_purchase)), coupon_per_cycle_usd (= face_value_usd *
-- quantity * interest_rate/100 / 2), the two semiannual pay-dates, and
-- every per-bond-name aggregate are pure functions of these columns,
-- computed at read time in internal/service/bondcalc.go. Storing them would
-- only create drift when a row is edited. (Contrast holdings.value_idr,
-- which IS stored — but only because it freezes a market rate at a point in
-- time. Here usd_idr_at_purchase is the frozen input and the rest is
-- arithmetic.)
--
-- Coupons are assumed semiannual, anchored on maturity_date: a bullet
-- bond's issue date is a whole number of years before maturity, so the
-- maturity month/day IS the coupon month/day. No coupon_frequency column —
-- every bond held so far pays twice a year, and adding
-- `smallint NOT NULL DEFAULT 2` later is a zero-risk migration.
CREATE TABLE bond_purchases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    bond_name text NOT NULL,
    -- Free text, no CHECK: every CHECK in this schema guards a closed enum
    -- that Go actually branches on, and nothing branches on platform. The
    -- frontend offers OCBC/MyBCA/Mandiri in a select but allows others.
    -- holdings.brand is the precedent.
    platform text NOT NULL DEFAULT '',
    -- Annual coupon rate as a PERCENT, not a fraction: 5.69 means 5.69%,
    -- matching the broker statement and the /100 in the coupon formula.
    -- 4 decimals covers odd-lot coupons like 4.7625%.
    interest_rate numeric(6, 4) NOT NULL,
    buy_date date NOT NULL,
    maturity_date date NOT NULL,
    -- Number of lots bought. numeric(14,3) matches holdings.qty/gram.
    quantity numeric(14, 3) NOT NULL DEFAULT 1,
    -- Redemption value per lot; $1000 for every bond so far.
    face_value_usd numeric(14, 2) NOT NULL DEFAULT 1000,
    -- Clean settlement amount for the WHOLE lot, not a per-unit price:
    -- $1,608.00 for 2 x $1,000 face is a price of 80.4% of par.
    price_usd numeric(14, 2) NOT NULL,
    accrued_interest_usd numeric(14, 2) NOT NULL DEFAULT 0,
    -- IDR per 1 USD on the settlement date, same unit as
    -- rate_entries.usd_idr. Frozen here rather than looked up later,
    -- because the rate that matters is the one actually paid.
    usd_idr_at_purchase numeric(14, 2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- user_id leads so this also serves plain per-user lookups; bond_name
-- second because every read path groups by it.
CREATE INDEX idx_bond_purchases_user_name ON bond_purchases (user_id, bond_name);

-- +goose Down
DROP TABLE IF EXISTS bond_purchases;

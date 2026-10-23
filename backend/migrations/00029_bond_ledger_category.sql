-- +goose Up
-- The bond ledger now feeds two Assets categories instead of one: each bond
-- is either a government bond (bonds_usd, relabelled "Bonds Gov.") or a
-- corporate one (bonds_corp). The key stays bonds_usd so valuation.go and
-- existing holdings need no rewrite.
UPDATE categories SET label = 'Bonds Gov.' WHERE key = 'bonds_usd';

-- Stored per purchase, but kept uniform across every purchase of the same
-- bond_name by BondPurchasesService (a bond is one Assets row, so it can
-- only sit in one category). CHECK rather than an FK to categories.key:
-- Go branches on exactly these two values, the same closed-enum rule the
-- rest of this schema follows. Every existing purchase starts as a
-- government bond, matching where it already sits on Assets.
ALTER TABLE bond_purchases
    ADD COLUMN category_key text NOT NULL DEFAULT 'bonds_usd'
        CHECK (category_key IN ('bonds_usd', 'bonds_corp'));

-- +goose Down
ALTER TABLE bond_purchases DROP COLUMN category_key;
UPDATE categories SET label = 'Bonds USD' WHERE key = 'bonds_usd';

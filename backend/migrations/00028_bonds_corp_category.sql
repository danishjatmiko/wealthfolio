-- +goose Up
-- "Bonds Corp.": USD corporate bonds, entered by hand on Assets. Valued
-- like bonds_usd/us_etf (usd_value × usd_idr, see valuation.go), but not
-- fed by the bond ledger — that only feeds bonds_usd.
-- Slotted in right after Bonds USD, so everything after it shifts down one.
UPDATE categories SET sort_order = sort_order + 1 WHERE sort_order >= 4;
INSERT INTO categories (key, label, color_oklch, kind, price_linked, sort_order) VALUES
    ('bonds_corp', 'Bonds Corp.', 'oklch(0.56 0.09 222)', 'asset', true, 4)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM categories WHERE key = 'bonds_corp';
UPDATE categories SET sort_order = sort_order - 1 WHERE sort_order >= 5;

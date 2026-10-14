-- +goose Up
-- "US ETF" now also holds individual US stocks, so relabel it. The key stays
-- us_etf — valuation.go keys its USD conversion off it.
UPDATE categories SET label = 'US ETF / Stocks' WHERE key = 'us_etf';

-- "ESOP": employee stock options/grants. Plain pass-through category like
-- properti/crypto — not price-linked, value_idr used as-is. Slotted in right
-- after US ETF / Stocks, so everything after it shifts down one.
UPDATE categories SET sort_order = sort_order + 1 WHERE sort_order >= 6;
INSERT INTO categories (key, label, color_oklch, kind, price_linked, sort_order) VALUES
    ('esop', 'ESOP', 'oklch(0.68 0.11 108)', 'asset', false, 6)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM categories WHERE key = 'esop';
UPDATE categories SET sort_order = sort_order - 1 WHERE sort_order >= 7;
UPDATE categories SET label = 'US ETF' WHERE key = 'us_etf';

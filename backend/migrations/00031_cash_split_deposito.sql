-- +goose Up
-- Cash splits by currency: uang_tunai becomes "IDR Cash" (the key stays, so
-- nothing keyed on it needs rewriting), and dollar balances get their own
-- "USD Cash" category, valued like bonds_usd/us_etf (usd_value × usd_idr,
-- see valuation.go). Time deposits get "Deposito", a plain IDR pass-through
-- like esop/properti. Both slot in right after IDR Cash, so everything
-- after it shifts down two.
UPDATE categories SET label = 'IDR Cash' WHERE key = 'uang_tunai';

UPDATE categories SET sort_order = sort_order + 2
WHERE sort_order > (SELECT sort_order FROM categories WHERE key = 'uang_tunai');
INSERT INTO categories (key, label, color_oklch, kind, price_linked, sort_order)
SELECT v.key, v.label, v.color, 'asset', v.price_linked, c.sort_order + v.offset_
FROM categories c,
     (VALUES ('usd_cash', 'USD Cash', 'oklch(0.52 0.09 180)', true, 1),
             ('deposito', 'Deposito', 'oklch(0.72 0.12 258)', false, 2))
         AS v(key, label, color, price_linked, offset_)
WHERE c.key = 'uang_tunai'
ON CONFLICT (key) DO NOTHING;

-- Cash rows already held in dollars (the old IDR/USD toggle set
-- currency = 'USD') are USD cash by definition, so they move over in every
-- snapshot, keeping each category's history consistent. Values are
-- untouched — a moved row keeps the IDR figure it was frozen at.
UPDATE holdings
SET category_id = (SELECT id FROM categories WHERE key = 'usd_cash')
WHERE category_id = (SELECT id FROM categories WHERE key = 'uang_tunai')
  AND currency = 'USD';

-- +goose Down
UPDATE holdings
SET category_id = (SELECT id FROM categories WHERE key = 'uang_tunai')
WHERE category_id = (SELECT id FROM categories WHERE key = 'usd_cash');
UPDATE holdings
SET category_id = (SELECT id FROM categories WHERE key = 'uang_tunai')
WHERE category_id = (SELECT id FROM categories WHERE key = 'deposito');
UPDATE passive_income_entries
SET category_id = (SELECT id FROM categories WHERE key = 'uang_tunai')
WHERE category_id IN (SELECT id FROM categories WHERE key IN ('usd_cash', 'deposito'));
DELETE FROM categories WHERE key IN ('usd_cash', 'deposito');
UPDATE categories SET sort_order = sort_order - 2
WHERE sort_order > (SELECT sort_order FROM categories WHERE key = 'uang_tunai');
UPDATE categories SET label = 'Uang Tunai' WHERE key = 'uang_tunai';

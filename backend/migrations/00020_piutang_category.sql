-- +goose Up
-- "Piutang" (receivable): a debt entry with direction='owed_to_me' is
-- itself an asset from the user's side, so it gets mirrored onto the
-- Assets page under this category (see DebtEntriesService.Create). Plain
-- pass-through category like properti/crypto — not price-linked, value_idr
-- used as-is.
INSERT INTO categories (key, label, color_oklch, kind, price_linked, sort_order) VALUES
    ('piutang', 'Piutang', 'oklch(0.64 0.10 130)', 'asset', false, 10)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM categories WHERE key = 'piutang';

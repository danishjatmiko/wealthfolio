-- +goose Up
INSERT INTO users (id, email, display_name)
VALUES ('00000000-0000-0000-0000-000000000001', NULL, 'Default User')
ON CONFLICT (id) DO NOTHING;

INSERT INTO categories (key, label, color_oklch, kind, price_linked, sort_order) VALUES
    ('logam_mulia', 'Logam Mulia', 'oklch(0.76 0.11 84)', 'asset', true, 1),
    ('saham', 'Saham', 'oklch(0.60 0.10 152)', 'asset', false, 2),
    ('bonds_usd', 'Bonds USD', 'oklch(0.60 0.10 248)', 'asset', true, 3),
    ('uang_tunai', 'Uang Tunai', 'oklch(0.66 0.09 196)', 'asset', true, 4),
    ('us_etf', 'US ETF', 'oklch(0.58 0.11 292)', 'asset', false, 5),
    ('properti', 'Properti', 'oklch(0.62 0.11 34)', 'asset', false, 6),
    ('crypto', 'Crypto', 'oklch(0.70 0.12 56)', 'asset', false, 7),
    ('reksa_dana', 'Reksa Dana', 'oklch(0.66 0.07 322)', 'asset', false, 8),
    ('liabilitas', 'Liabilitas', 'oklch(0.58 0.13 28)', 'liability', false, 9)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM categories WHERE key IN
    ('logam_mulia','saham','bonds_usd','uang_tunai','us_etf','properti','crypto','reksa_dana','liabilitas');
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';

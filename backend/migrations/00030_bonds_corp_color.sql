-- +goose Up
-- Bonds Corp.'s original blue-teal sat only 26° of hue from Bonds Gov.'s
-- blue, so the two blurred together side by side on the Passive Income ring
-- and the allocation donut. Rose sits well clear of Bonds Gov. and of every
-- other category, at the same lightness/chroma as the rest of the palette.
UPDATE categories SET color_oklch = 'oklch(0.62 0.12 355)' WHERE key = 'bonds_corp';

-- +goose Down
UPDATE categories SET color_oklch = 'oklch(0.56 0.09 222)' WHERE key = 'bonds_corp';

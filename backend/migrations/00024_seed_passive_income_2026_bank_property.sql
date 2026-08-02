-- +goose Up
-- The "Bunga bank" and "Properti" sections of the same 2026 spreadsheet
-- imported by 00023, which only covered Stock and Reksa Dana.
--
-- Bank interest and cashback file under Uang Tunai — the money sits in a
-- savings account, so that's the asset class paying it out. Rent files
-- under Properti. Both categories are already seeded (00002_seed.sql), so
-- these show up as their own slices on the Passive Income ring.
--
-- "Another sewa apartemen" carried no date in the sheet; dated 15 Sep 2026
-- per the owner rather than guessed at.
INSERT INTO passive_income_entries (user_id, category_id, name, amount_idr, received_date, income_type, note)
SELECT
    '00000000-0000-0000-0000-000000000001',
    c.id,
    v.name,
    v.amount_idr,
    v.received_date::date,
    v.income_type,
    v.note
FROM (
    VALUES
        ('uang_tunai', 'Superbank', 1200000, '2026-02-28', 'Bunga', ''),
        ('uang_tunai', 'Superbank', 1100000, '2026-03-25', 'Bunga', ''),
        ('uang_tunai', 'Superbank', 540000, '2026-04-28', 'Bunga', ''),
        ('uang_tunai', 'Superbank', 800000, '2026-05-30', 'Bunga', ''),
        ('uang_tunai', 'Superbank', 850000, '2026-06-30', 'Bunga', ''),
        ('uang_tunai', 'Cashback OCBC', 350000, '2026-07-31', 'Cashback', ''),
        ('uang_tunai', 'Superbank', 500000, '2026-07-31', 'Bunga', ''),
        ('properti', 'Sewa Apartemen Tahap 1', 6000000, '2026-06-01', 'Sewa', ''),
        ('properti', 'Another sewa apartemen', 24000000, '2026-09-15', 'Sewa', 'No date in the source sheet; filed on 15 Sep 2026.')
) AS v (category_key, name, amount_idr, received_date, income_type, note)
JOIN categories c ON c.key = v.category_key
WHERE EXISTS (SELECT 1 FROM users WHERE id = '00000000-0000-0000-0000-000000000001');

-- +goose Down
-- Matched on exact (name, date, amount) triples, same as 00023 — a broader
-- delete would take any bank/rent income logged since with it.
DELETE FROM passive_income_entries p
USING (
    VALUES
        ('Superbank', '2026-02-28'::date, 1200000::bigint),
        ('Superbank', '2026-03-25', 1100000),
        ('Superbank', '2026-04-28', 540000),
        ('Superbank', '2026-05-30', 800000),
        ('Superbank', '2026-06-30', 850000),
        ('Cashback OCBC', '2026-07-31', 350000),
        ('Superbank', '2026-07-31', 500000),
        ('Sewa Apartemen Tahap 1', '2026-06-01', 6000000),
        ('Another sewa apartemen', '2026-09-15', 24000000)
) AS v (name, received_date, amount_idr)
WHERE p.user_id = '00000000-0000-0000-0000-000000000001'
  AND p.name = v.name
  AND p.received_date = v.received_date
  AND p.amount_idr = v.amount_idr;

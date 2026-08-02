-- +goose Up
-- The 2026 "Realized P/L" ledger, imported from the spreadsheet this app
-- replaced. Seeds the pre-auth default user (00002_seed.sql) — the same
-- account the first Google login claims in place, so these land on the
-- real user either way.
--
-- Every row from the sheet is here, including the negative ones: a realized
-- capital loss and a cut-loss redemption are part of what the year actually
-- paid out, and netting them out silently would overstate the total. The
-- sheet's own two sections map to category_id (Saham / Reksa Dana) and its
-- type column to income_type; the two rows it left untyped ("Cleo", "BCA")
-- come across as "Other" rather than being guessed at.
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
        ('saham', 'Dividen BBRI', 11000000, '2026-01-15', 'Dividen', ''),
        ('saham', 'Dividen BMRI', 700000, '2026-01-14', 'Dividen', ''),
        ('saham', 'Capital BBNI', 1800000, '2026-01-21', 'Capital', ''),
        ('saham', 'Capital BBCA', -2400000, '2026-01-30', 'Capital', 'ini sbenernya beli 33 lot seharga 23.3, ku jual 24.1, tapi karena kena harga rata2, stock bit ngitungnya rugi'),
        ('saham', 'Capital BBRI', -1900000, '2026-01-30', 'Capital', 'ini sbenernya beli 84 lot seharga 30.07, ku jual 31.8, tapi karena kena harga rata2, stock bit ngitungnya rugi'),
        ('saham', 'Capital Mandiri ajaib', 1000000, '2026-03-04', 'Capital', ''),
        ('saham', 'Dividen BBCA', 5000000, '2026-04-08', 'Dividen', ''),
        ('saham', 'Dividen BBRI', 21500000, '2026-05-08', 'Dividen', ''),
        ('saham', 'Dividen BTPS', 1900000, '2026-05-22', 'Dividen', ''),
        ('saham', 'Dividen Mandiri', 22000000, '2026-05-25', 'Dividen', ''),
        ('saham', 'BCA', 346000, '2026-06-29', 'Other', ''),
        ('saham', 'Dividen BIRD', 21000000, '2026-07-10', 'Dividen', ''),
        ('saham', 'ERAL', 2400000, '2026-07-17', 'Dividen', ''),
        ('saham', 'IPO Hunter', 1000000, '2026-07-21', 'IPO', ''),
        ('saham', 'Cleo', 100000, '2026-07-22', 'Other', ''),
        ('reksa_dana', 'I-Hajj', 7600000, '2026-02-09', 'Reksa Dana', '7.6jt itu bonus, which 3%'),
        ('reksa_dana', 'KIM Fixed Income', 9500000, '2026-02-09', 'Reksa Dana', 'kim modal 100, dpt nya 9.5%, 9.5jt'),
        ('reksa_dana', 'Avrist Emerald', 13000000, '2026-02-09', 'Reksa Dana', 'avrist modal 150, dptnya 8.6%, 13jt. In total, 30/250 = 12% in 11 months'),
        ('reksa_dana', 'Insight Renewable hana', 7200000, '2026-06-10', 'Reksa Dana', 'ini gara2 bunga bank naik drastis, lgs anjlok, harusnya dapet 8jtan'),
        ('reksa_dana', 'Insight Renewable mama', 100000, '2026-06-15', 'Reksa Dana', 'ini gara2 bunga bank naik drastis, lgs anjlok, harusnya dapet 1jtan'),
        ('reksa_dana', 'iHajj Hana', 5300000, '2026-06-18', 'Reksa Dana', 'harusnya ada 5 juta lagi, cuma ilang gara2 bunga BI naik'),
        ('reksa_dana', 'iHajj Mama', -4800000, '2026-06-18', 'Reksa Dana', 'ini minus gara2 bank bunga naik, cut loss')
) AS v (category_key, name, amount_idr, received_date, income_type, note)
JOIN categories c ON c.key = v.category_key
WHERE EXISTS (SELECT 1 FROM users WHERE id = '00000000-0000-0000-0000-000000000001');

-- +goose Down
-- Matched on the exact (name, date, amount) triples seeded above rather
-- than "everything dated 2026": by the time anyone rolls this back the user
-- may well have logged 2026 income of their own, and a date-range delete
-- would take that with it.
DELETE FROM passive_income_entries p
USING (
    VALUES
        ('Dividen BBRI', '2026-01-15'::date, 11000000::bigint),
        ('Dividen BMRI', '2026-01-14', 700000),
        ('Capital BBNI', '2026-01-21', 1800000),
        ('Capital BBCA', '2026-01-30', -2400000),
        ('Capital BBRI', '2026-01-30', -1900000),
        ('Capital Mandiri ajaib', '2026-03-04', 1000000),
        ('Dividen BBCA', '2026-04-08', 5000000),
        ('Dividen BBRI', '2026-05-08', 21500000),
        ('Dividen BTPS', '2026-05-22', 1900000),
        ('Dividen Mandiri', '2026-05-25', 22000000),
        ('BCA', '2026-06-29', 346000),
        ('Dividen BIRD', '2026-07-10', 21000000),
        ('ERAL', '2026-07-17', 2400000),
        ('IPO Hunter', '2026-07-21', 1000000),
        ('Cleo', '2026-07-22', 100000),
        ('I-Hajj', '2026-02-09', 7600000),
        ('KIM Fixed Income', '2026-02-09', 9500000),
        ('Avrist Emerald', '2026-02-09', 13000000),
        ('Insight Renewable hana', '2026-06-10', 7200000),
        ('Insight Renewable mama', '2026-06-15', 100000),
        ('iHajj Hana', '2026-06-18', 5300000),
        ('iHajj Mama', '2026-06-18', -4800000)
) AS v (name, received_date, amount_idr)
WHERE p.user_id = '00000000-0000-0000-0000-000000000001'
  AND p.name = v.name
  AND p.received_date = v.received_date
  AND p.amount_idr = v.amount_idr;

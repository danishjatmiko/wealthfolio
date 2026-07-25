-- +goose Up
-- Corrects the bca catalog entry seeded by 00012 with real data captured
-- from an actual BCA Mobile notification ("Financial Diary: Pengeluaran
-- sebesar IDR 1.00 di kategori Makanan.") — the app is "BCA Mobile"
-- (package com.bca), not myBCA, and its amounts are Western-decimal
-- formatted ("IDR 1.00"), not Rupiah-grouped ("Rp1.234") like GoPay's.
UPDATE notification_apps
SET display_name = 'BCA Mobile',
    amount_thousand_sep = ',',
    amount_decimal_sep = '.',
    updated_at = now()
WHERE source = 'bca';

DELETE FROM notification_app_packages
WHERE package_name = 'com.bca.mybca.omni.android';

INSERT INTO notification_app_packages (app_id, package_name)
SELECT id, 'com.bca' FROM notification_apps WHERE source = 'bca'
ON CONFLICT (package_name) DO NOTHING;

-- GoPay: two notification templates seen so far — an outgoing transfer to
-- a person, and a merchant payment. Both captured from real GoPay
-- notifications.
INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 10, 'text',
       'successfully transferred Rp(?P<amount>[\d.]+) to (?P<merchant>.+?)\.?$',
       'Outgoing transfer, e.g. "You''ve successfully transferred Rp1.234 to Mba Jen."',
       true
FROM notification_apps WHERE source = 'gopay'
ON CONFLICT (app_id, priority) DO NOTHING;

INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 20, 'text',
       'you just paid Rp(?P<amount>[\d.]+) to (?P<merchant>.+?)\.?$',
       'Merchant payment, e.g. "Awesome, you just paid Rp1 to Kantin Sedap Rasa, Karet Semanggi."',
       true
FROM notification_apps WHERE source = 'gopay'
ON CONFLICT (app_id, priority) DO NOTHING;

-- BCA Mobile: no merchant name in this notification format, so its
-- spending category (e.g. "Makanan") is captured into the merchant slot
-- instead.
INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 10, 'text',
       'Pengeluaran sebesar IDR (?P<amount>[\d.,]+) di kategori (?P<merchant>.+?)\.?$',
       'BCA mobile (ID), e.g. "Financial Diary: Pengeluaran sebesar IDR 1.00 di kategori Makanan."',
       true
FROM notification_apps WHERE source = 'bca'
ON CONFLICT (app_id, priority) DO NOTHING;

-- +goose Down
DELETE FROM notification_patterns
WHERE app_id = (SELECT id FROM notification_apps WHERE source = 'bca')
  AND regex = 'Pengeluaran sebesar IDR (?P<amount>[\d.,]+) di kategori (?P<merchant>.+?)\.?$';

DELETE FROM notification_patterns
WHERE app_id = (SELECT id FROM notification_apps WHERE source = 'gopay')
  AND regex IN (
    'successfully transferred Rp(?P<amount>[\d.]+) to (?P<merchant>.+?)\.?$',
    'you just paid Rp(?P<amount>[\d.]+) to (?P<merchant>.+?)\.?$'
  );

DELETE FROM notification_app_packages WHERE package_name = 'com.bca';

INSERT INTO notification_app_packages (app_id, package_name)
SELECT id, 'com.bca.mybca.omni.android' FROM notification_apps WHERE source = 'bca'
ON CONFLICT (package_name) DO NOTHING;

UPDATE notification_apps
SET display_name = 'BCA / m-BCA',
    amount_thousand_sep = '.',
    amount_decimal_sep = ',',
    updated_at = now()
WHERE source = 'bca';

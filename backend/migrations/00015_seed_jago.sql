-- +goose Up
-- Bank Jago: transfer-to-person and merchant-payment templates, both
-- captured from real Jago notifications. Idempotent so it's a no-op on
-- databases (like local dev) that already have this from manual testing.
INSERT INTO notification_apps (source, display_name, amount_thousand_sep, amount_decimal_sep, enabled)
VALUES ('jago', 'Bank Jago', '.', ',', true)
ON CONFLICT (source) DO NOTHING;

INSERT INTO notification_app_packages (app_id, package_name)
SELECT id, 'com.jago.digitalBanking' FROM notification_apps WHERE source = 'jago'
ON CONFLICT (package_name) DO NOTHING;

INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 10, 'text',
       'melakukan transfer Rp\s*(?P<amount>[\d.,]+) ke (?P<merchant>.+?)\. Butuh bantuan',
       'Outgoing transfer, e.g. "Kamu telah melakukan transfer Rp1 ke JEAN GEISENA LANGUYU. Butuh bantuan? ..."',
       true
FROM notification_apps WHERE source = 'jago'
ON CONFLICT (app_id, priority) DO NOTHING;

INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 20, 'text',
       'membayar Rp\s*(?P<amount>[\d.,]+) ke (?P<merchant>.+?)\. Butuh bantuan',
       'Merchant payment, e.g. "Kamu telah membayar Rp 1 ke NONIK FOOD. Butuh bantuan? ..."',
       true
FROM notification_apps WHERE source = 'jago'
ON CONFLICT (app_id, priority) DO NOTHING;

-- +goose Down
DELETE FROM notification_apps WHERE source = 'jago';

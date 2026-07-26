-- +goose Up
-- GoPay: a third notification template — payment via a linked payment
-- method (e.g. GoPay Tabungan by Jago), which names both the merchant and
-- the funding source: "You just paid Rp139.296 to TOKOPEDIA using GoPay
-- Tabungan by Jago." Anchored on lowercase "paid Rp" (not "You"/"you just
-- paid") so it matches regardless of how the sentence starts, and merchant
-- is bounded by " using" rather than end-of-string since this template has
-- trailing content after the merchant name.
INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
SELECT id, 30, 'text',
       'paid Rp(?P<amount>[\d.]+) to (?P<merchant>.+?) using',
       'Payment via linked payment method, e.g. "You just paid Rp139.296 to TOKOPEDIA using GoPay Tabungan by Jago."',
       true
FROM notification_apps WHERE source = 'gopay'
ON CONFLICT (app_id, priority) DO NOTHING;

-- +goose Down
DELETE FROM notification_patterns
WHERE app_id = (SELECT id FROM notification_apps WHERE source = 'gopay')
  AND regex = 'paid Rp(?P<amount>[\d.]+) to (?P<merchant>.+?) using';

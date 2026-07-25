-- +goose Up
-- Seeds the 3 known sources with corrected package names. No patterns yet
-- — real GoPay/DANA/BCA notification formats haven't been captured, so
-- there's nothing to write a regex against until the Android app starts
-- forwarding real notifications (see notification_expense_events' raw_*
-- columns, kept for exactly this purpose).
INSERT INTO notification_apps (source, display_name, amount_thousand_sep, amount_decimal_sep, enabled) VALUES
    ('gopay', 'GoPay', '.', ',', true),
    ('dana', 'DANA', '.', ',', true),
    ('bca', 'BCA / m-BCA', '.', ',', true);

INSERT INTO notification_app_packages (app_id, package_name)
SELECT id, 'com.gojek.gopay' FROM notification_apps WHERE source = 'gopay'
UNION ALL
SELECT id, 'id.dana' FROM notification_apps WHERE source = 'dana'
UNION ALL
SELECT id, 'com.bca.mybca.omni.android' FROM notification_apps WHERE source = 'bca';

-- +goose Down
DELETE FROM notification_apps WHERE source IN ('gopay', 'dana', 'bca');

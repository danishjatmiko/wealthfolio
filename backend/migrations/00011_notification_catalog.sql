-- +goose Up
-- notification_apps: catalog of e-wallet/bank apps Android is allowed to
-- read notifications from. Replaces the hardcoded NotificationSource enum
-- — Android fetches this via GET /notification-apps. `source` matches the
-- existing free-text `source` column on expense_source_mappings /
-- notification_expense_events (not FK'd, so catalog edits never touch
-- historical audit rows).
CREATE TABLE notification_apps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source text NOT NULL UNIQUE,
    display_name text NOT NULL,
    -- Amount formatting ("Rp50.000" vs "Rp 50,000.00") is a property of
    -- the app/bank, not of one notification template, so it lives here.
    amount_thousand_sep char(1) NOT NULL DEFAULT '.',
    amount_decimal_sep char(1) NOT NULL DEFAULT ',',
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- notification_app_packages: 1:N Android package names per app, in its own
-- table (not a text[] column) so a package name can carry its own UNIQUE
-- constraint — two catalog entries can never claim the same installed app.
CREATE TABLE notification_app_packages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id uuid NOT NULL REFERENCES notification_apps (id) ON DELETE CASCADE,
    package_name text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_notification_app_packages_app ON notification_app_packages (app_id);

-- notification_patterns: ordered regex templates per app (lowest priority
-- tried first, first match wins). One RE2 regex per row with named groups
-- (?P<amount>...) and (?P<merchant>...), tested against one raw field —
-- covers the common case where both appear in the same notification field.
CREATE TABLE notification_patterns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id uuid NOT NULL REFERENCES notification_apps (id) ON DELETE CASCADE,
    priority integer NOT NULL,
    field text NOT NULL CHECK (field IN ('title', 'text', 'big_text')),
    regex text NOT NULL,
    description text,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (app_id, priority)
);
CREATE INDEX idx_notification_patterns_app ON notification_patterns (app_id);

-- +goose Down
DROP TABLE notification_patterns;
DROP TABLE notification_app_packages;
DROP TABLE notification_apps;

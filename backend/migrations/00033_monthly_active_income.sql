-- +goose Up
CREATE TABLE monthly_active_income (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    month date NOT NULL CHECK (month >= DATE '2026-08-01' AND EXTRACT(DAY FROM month) = 1),
    amount_idr bigint NOT NULL CHECK (amount_idr >= 0 AND amount_idr <= 9007199254740991),
    note text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, month)
);

-- +goose Down
DROP TABLE IF EXISTS monthly_active_income;

-- +goose Up
-- Nullable: manually-created expenses (POST/PUT /fixed-expenses) have no
-- source; only ones created via notification ingestion do. Immutable after
-- creation — editing an expense's name/amount never touches this, so the
-- provenance survives a manual correction.
ALTER TABLE fixed_expenses ADD COLUMN source text;

-- +goose Down
ALTER TABLE fixed_expenses DROP COLUMN source;

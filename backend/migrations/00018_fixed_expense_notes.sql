-- +goose Up
-- Free-form notes a user can attach to a fixed expense (e.g. context on
-- what it was for) — nullable, manual-entry only, shown/edited from the
-- expense modal.
ALTER TABLE fixed_expenses ADD COLUMN notes text;

-- +goose Down
ALTER TABLE fixed_expenses DROP COLUMN notes;

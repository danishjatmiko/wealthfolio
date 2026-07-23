-- +goose Up
-- Email/password sign-in as a second login method alongside Google. Stays
-- nullable at the column level since every existing user (and any new
-- Google sign-up) has no password — only accounts that explicitly opt in
-- get a hash. Seeding an actual account's hash is done directly against the
-- database, not via a migration, so it never ends up committed to source
-- control.
ALTER TABLE users ADD COLUMN password_hash text;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

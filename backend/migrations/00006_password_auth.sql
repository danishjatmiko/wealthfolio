-- +goose Up
-- Email/password sign-in as a second login method alongside Google. Stays
-- nullable at the column level since every existing user (and any new
-- Google sign-up) has no password — only accounts that explicitly opt in
-- get a hash. The value below is an Argon2id hash (OWASP's recommended
-- password-hashing algorithm: m=19MiB, t=2, p=1 — see
-- internal/service/password.go) of a single, temporary password for one
-- named account; the plaintext itself is never stored anywhere.
ALTER TABLE users ADD COLUMN password_hash text;

UPDATE users
SET password_hash = '$argon2id$v=19$m=19456,t=2,p=1$lyZqlGMimAyAeu5hQvJEyA$GdtDxiZjs60NBhSn+Mb23T3DThmKlYJLx1iHRO5TxJU'
WHERE email = 'danishkho19@gmail.com';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

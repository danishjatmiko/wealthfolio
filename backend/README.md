# Etherna backend

Go API server for Etherna — see the [root README](../README.md) for what the whole project is, and [DOCUMENTATION.md](../DOCUMENTATION.md) for the full architecture/schema/API reference. This file is just the backend module's local dev quickstart.

## Stack

Go 1.25 · [`chi`](https://github.com/go-chi/chi) router · [`pgx/v5`](https://github.com/jackc/pgx) (no ORM — hand-written SQL) · PostgreSQL · [`goose`](https://github.com/pressly/goose) migrations · Google OAuth 2.0 + Argon2id password auth.

## Layout

```
cmd/
  api/            entrypoint: config, migrations, DB pool, HTTP server
  rates-sync/     standalone cron script (see below) — not part of the API binary
internal/
  domain/         plain structs shared across every layer, no logic
  db/             pgxpool + hand-written repository queries, one file per resource
  service/        business logic (locking, value derivation, notification parsing, aggregation)
  httpapi/        chi router, handlers, middleware (auth, CORS, body-size limit)
migrations/       goose SQL migrations, embedded into the binary via //go:embed
```

Strict dependency order: `domain ← db ← service ← httpapi`. Nothing below `httpapi` knows about HTTP; nothing below `service` knows about business rules.

## Local development

```bash
createdb wealthfolio_dev   # needs a local Postgres; migrations run automatically on API startup

DATABASE_URL="postgres://<user>@localhost:5432/wealthfolio_dev?sslmode=disable" \
PORT=8080 \
CORS_ORIGIN="http://localhost:5173" \
GOOGLE_CLIENT_ID="<from Google Cloud Console>" \
GOOGLE_CLIENT_SECRET="<from Google Cloud Console>" \
GOOGLE_REDIRECT_URL="http://localhost:8080/api/v1/auth/google/callback" \
APP_BASE_URL="http://localhost:5173" \
go run ./cmd/api
```

`DATABASE_URL`, `GOOGLE_CLIENT_ID`, and `GOOGLE_CLIENT_SECRET` are the only required env vars — the binary refuses to start without them. Everything else has a sane local-dev default (see `internal/config/config.go`). Full env var reference is in the root [DOCUMENTATION.md](../DOCUMENTATION.md#42-configuration).

Serves at `http://localhost:8080/api/v1`; health check at `GET /healthz` (no auth, no prefix).

## Tests

```bash
go test ./...
go vet ./...
```

## Migrations

Plain SQL files in `migrations/`, run automatically in order on every API startup — nothing to run by hand. To add one, drop a new `NNNNN_description.sql` file with `-- +goose Up` / `-- +goose Down` blocks; see any existing file for the style.

## Money unit convention

**Every `*_idr`/`value_idr` field, across every table, is an integer in full/raw Rupiah — no scaling factor.** (This changed in migration `00014_exact_rupiah_unit.sql`; older code/comments mentioning "thousands of IDR" predate that and are wrong if you find any lingering.) The only non-IDR numeric fields are physical quantities (`gram`, `qty`) and USD amounts (`usd_value`).

## `cmd/rates-sync`

A separate, standalone binary — not started by `cmd/api`, meant to run once daily via cron on the deploy host. Scrapes Antam/UBS (indogold.id), King Halim (kinghalim.com), and USD/IDR (Google Finance) and `POST`s them to a *running* API's `/api/v1/rates`, authenticating via the `X-Rates-Sync-Token` header rather than a session login (see `RatesSyncOrAuthMiddleware`). Falls back to the previous day's value (via `GET /rates/latest`) if an individual scrape fails, and only hard-fails if a scrape fails with no prior value to fall back to.

```bash
RATES_SYNC_TOKEN="<shared secret, must match the API's RATES_SYNC_TOKEN>" \
ETHERNA_API_BASE_URL="https://etherna.id/api/v1" \
go run ./cmd/rates-sync
```

The API side needs `RATES_SYNC_TOKEN` and `RATES_SYNC_EMAIL` (which account the synced rates get attributed to) set for this to authenticate — both are optional/unset by default, which simply disables the token path entirely (falls through to requiring a normal session).

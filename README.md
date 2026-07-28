# Etherna

A personal net-worth / asset-portfolio and monthly-expense tracker — replaces a monthly spreadsheet workflow with a real app you can update from your phone or laptop. PostgreSQL + Go backend, React web frontend, and an Android app that both hosts the web frontend in a WebView and auto-captures e-wallet/bank payment notifications into expense records.

## Stack

- **Database**: PostgreSQL (schema + seed in [`backend/migrations/`](backend/migrations/), applied automatically on API startup via [goose](https://github.com/pressly/goose))
- **Backend**: Go, `chi` router, `pgx` (no ORM — hand-written SQL in [`backend/internal/db/`](backend/internal/db/))
- **Auth**: Google Sign-In (OAuth 2.0 Authorization Code flow via `golang.org/x/oauth2`) or email/password (Argon2id), server-side sessions in Postgres — see [Authentication](#authentication) below
- **Web frontend**: React + TypeScript + Vite, `react-router`, `@tanstack/react-query`, hand-rolled SVG charts
- **Android app**: Kotlin + Jetpack Compose, hosts the web frontend in a WebView and runs a `NotificationListenerService` that auto-captures GoPay/DANA/BCA/Bank Jago payment notifications into expenses — see [`android/README.md`](android/README.md)
- **Deploy**: Docker Compose (Postgres + API + Caddy serving the built frontend and reverse-proxying `/api/*`)

## What's in here

- **Portfolio tracking**: dated snapshots of asset holdings (gold, stocks, USD bonds/ETF, cash, property, crypto, etc.) and debts, each snapshot immutable once superseded by a newer one — only the latest is ever editable.
- **Monthly expense tracking**: a custom pay-cycle "period" (25th of one month through the 24th of the next), budget envelopes with a committed target compared against realized spend, and individual fixed expenses — either entered manually or auto-created from a captured payment notification (see below).
- **Notification-driven expense capture**: the Android app watches for GoPay/DANA/BCA/Bank Jago payment notifications (only for sources the user explicitly enables) and forwards the raw, unparsed notification text to the backend, which matches it against a per-app catalog of regex patterns stored in the database — so adding support for a new app, or fixing a broken pattern after an app updates its notification format, is a database edit, not an APK release. See [`backend/internal/service/notificationparse/`](backend/internal/service/notificationparse/) and [`android/README.md`](android/README.md#the-notification-capture-pipeline).
- **Passive income, targets, and progress**: annual passive-income sources vs. a target, user-defined financial goals (equity/gold-grams/passive-income/debt-ratio/custom), and net-equity/debt time-series charts.
- **Daily gold/USD rate sync**: [`backend/cmd/rates-sync`](backend/cmd/rates-sync/) is a standalone cron script that scrapes Antam/UBS/King Halim gold prices and the USD/IDR rate and posts them to the API — see [`backend/README.md`](backend/README.md#cmdrates-sync).

## Important: money unit convention

**Every monetary field in the database and API (`value_idr`, `amount_idr`, `target_value`, gold prices, etc.) is an integer in full/raw IDR — whole Rupiah, no scaling factor.** E.g. `900000000` means Rp 900,000,000. (This used to be "thousands of IDR" — migration [`00014_exact_rupiah_unit.sql`](backend/migrations/00014_exact_rupiah_unit.sql) rescaled every affected column and removed that convention entirely; if you see a comment anywhere still claiming "thousands," it predates that migration and is wrong.) The only non-IDR numeric fields are physical quantities (`gram`, `qty`) and raw USD amounts (`usd_value`).

## Authentication

Every account signs in with Google **or** email/password and gets its own private, fully isolated workspace — every table is scoped by `user_id`, and every read/write path enforces it (see `internal/httpapi/middleware.go`'s `AuthMiddleware` and the ownership checks in `internal/db/*.go`). Sign-up is open: any Google account can sign in and gets a fresh, empty workspace. The very first Google login ever claims the original pre-auth seeded user in place, so existing local data isn't lost when auth is turned on.

Sessions are opaque random tokens in an `HttpOnly`/`Secure`/`SameSite=Lax` cookie, backed by a `sessions` table in Postgres (not JWTs — revocable instantly on logout). A session stays valid for 7 days of inactivity, refreshed automatically while in use, capped at 30 days from creation regardless of activity.

**Google OAuth setup**: create an OAuth 2.0 Client ID (type "Web application") at [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials), with an authorized redirect URI matching `GOOGLE_REDIRECT_URL` (e.g. `http://localhost:8080/api/v1/auth/google/callback` for local dev). Set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, and `APP_BASE_URL` (see `.env.example`) — the backend refuses to start without the client id/secret.

## Local development

### 1. Database

```bash
createdb wealthfolio_dev   # needs a local Postgres; migrations run automatically on API startup
```

### 2. Backend

```bash
cd backend
DATABASE_URL="postgres://<user>@localhost:5432/wealthfolio_dev?sslmode=disable" \
PORT=8080 \
CORS_ORIGIN="http://localhost:5173" \
GOOGLE_CLIENT_ID="<from Google Cloud Console>" \
GOOGLE_CLIENT_SECRET="<from Google Cloud Console>" \
GOOGLE_REDIRECT_URL="http://localhost:8080/api/v1/auth/google/callback" \
APP_BASE_URL="http://localhost:5173" \
go run ./cmd/api
```

Runs migrations, then serves the API at `http://localhost:8080/api/v1`. Health check: `GET /healthz`. See [`backend/README.md`](backend/README.md) for the full env var reference and how to run `cmd/rates-sync`.

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Serves at `http://localhost:5173`; `vite.config.ts` proxies `/api` to `http://localhost:8080` in dev. See [`frontend/README.md`](frontend/README.md).

### 4. Android app (optional)

Only needed if you're working on the Android client itself, or want notification-driven expense capture on a phone. See [`android/README.md`](android/README.md) — the short version is: update the hardcoded LAN IP in `app/build.gradle.kts`'s debug config to match your machine, then run from Android Studio or `./gradlew assembleDebug`.

## Production deploy (Docker Compose)

```bash
cp .env.example .env   # edit POSTGRES_PASSWORD, DOMAIN, CORS_ORIGIN, GOOGLE_*, RATES_SYNC_* as needed
docker compose up -d --build
```

This starts Postgres, the Go API, and Caddy (serving the built frontend + reverse-proxying `/api/*` to the API, with automatic HTTPS if `DOMAIN` is a real domain pointed at the server). See [`docker-compose.yml`](docker-compose.yml).

A production deployment typically also wants, outside this repo (not tracked — deploy-host-specific):
- A daily cron'd `pg_dump` of the database to a backup directory, with an age-based cleanup.
- A daily cron'd run of `cmd/rates-sync` (see [`backend/README.md`](backend/README.md#cmdrates-sync)) to keep gold/USD rates current without manual entry.

## Project layout

```
backend/
  cmd/api/                  entrypoint: config, migrations, DB pool, HTTP server
  cmd/rates-sync/           standalone daily cron script (gold/USD rate scraper)
  internal/domain/          plain structs shared across layers
  internal/db/              pgxpool + hand-written repository queries
  internal/service/         business logic (value derivation, dashboard/progress aggregation, notification parsing, targets)
  internal/service/notificationparse/  pure regex-matching engine for turning captured notification text into an amount + merchant
  internal/httpapi/         chi router, handlers, middleware
  migrations/               goose SQL migrations (embedded into the binary)

frontend/
  src/pages/                one folder per screen (Dashboard, Assets, Debts, Expenses, PassiveIncome, Targets, Progress, Rates)
  src/components/           AppShell (sidebar/header/bottom-nav + pull-to-refresh), Modal, DonutChart, LineChart
  src/hooks/                one react-query hook per API resource
  src/lib/                  format.ts (money formatting), holdingCalc.ts (client-side value preview), api.ts

android/
  app/src/main/kotlin/com/wealthfolio/mobile/
    web/                     WebView host + JS bridge to native screens
    notifications/           NotificationListenerService + idempotency-key builder
    sync/                    WorkManager scheduling, outbox sync/cleanup workers, Sync Status screen
    settings/                Settings screen (source toggles, envelope mapping)
    data/outbox/             durable local queue of captured notifications
    data/notificationcatalog/ local mirror of the backend's supported-apps catalog

docker-compose.yml          postgres + api + web (Caddy)
```

## What's not built yet

- No CSV/spreadsheet import — the user chose to re-enter historical snapshots manually rather than build one-time import tooling.

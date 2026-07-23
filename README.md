# Etherna

A personal net-worth / asset-portfolio tracker — replaces a monthly spreadsheet workflow with a real app you can update from your phone or laptop. PostgreSQL + Go backend + React frontend, built from the design in [`design_handoff_portfolio_app/`](design_handoff_portfolio_app/).

## Stack

- **Database**: PostgreSQL (schema + seed in [`backend/migrations/`](backend/migrations/), applied automatically on API startup via [goose](https://github.com/pressly/goose))
- **Backend**: Go, `chi` router, `pgx` (no ORM — hand-written SQL in [`backend/internal/db/`](backend/internal/db/))
- **Auth**: Google Sign-In (OAuth 2.0 Authorization Code flow via `golang.org/x/oauth2`), server-side sessions in Postgres — see [Authentication](#authentication-google-sign-in) below
- **Frontend**: React + TypeScript + Vite, `react-router`, `@tanstack/react-query`, hand-rolled SVG charts
- **Deploy**: Docker Compose (Postgres + API + Caddy serving the built frontend and reverse-proxying `/api/*`)

## Important: money unit convention

Every monetary field in the database and API (`value_idr`, `target_value`, `per_year_idr`, gold prices, etc.) is an integer in **thousands of IDR** — e.g. `3750000` means Rp 3,750,000,000. This matches the design prototype's internal unit exactly, so the `money()`/`goldPrice()`/derivation formulas could be ported verbatim. The one exception is `rate_entries.usd_idr`, which is full IDR per 1 USD. See the comment at the top of [`backend/migrations/00001_init.sql`](backend/migrations/00001_init.sql).

## Authentication: Google Sign-In

Every account signs in with Google and gets its own private, fully isolated workspace — every table is scoped by `user_id`, and every read/write path enforces it (see `internal/httpapi/middleware.go`'s `AuthMiddleware` and the ownership checks in `internal/db/*.go`). Sign-up is open: any Google account can sign in and gets a fresh, empty workspace. The very first Google login ever claims the original pre-auth seeded user in place, so existing local data isn't lost when auth is turned on.

Sessions are opaque random tokens in an `HttpOnly`/`Secure`/`SameSite=Lax` cookie, backed by a `sessions` table in Postgres (not JWTs — revocable instantly on logout). A session stays valid for 7 days of inactivity, refreshed automatically while in use, capped at 30 days from creation regardless of activity.

**Setup**: create an OAuth 2.0 Client ID (type "Web application") at [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials), with an authorized redirect URI matching `GOOGLE_REDIRECT_URL` (e.g. `http://localhost:8080/api/v1/auth/google/callback` for local dev). Set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, and `APP_BASE_URL` (see `.env.example`) — the backend refuses to start without the client id/secret.

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

Runs migrations, then serves the API at `http://localhost:8080/api/v1`. Health check: `GET /healthz`.

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Serves at `http://localhost:5173`; `vite.config.ts` proxies `/api` to `http://localhost:8080` in dev.

## Production deploy (Docker Compose)

```bash
cp .env.example .env   # edit POSTGRES_PASSWORD, DOMAIN, CORS_ORIGIN
docker compose up -d --build
```

This starts Postgres, the Go API, and Caddy (serving the built frontend + reverse-proxying `/api/*` to the API, with automatic HTTPS if `DOMAIN` is a real domain pointed at the server). See [`docker-compose.yml`](docker-compose.yml).

## Project layout

```
backend/
  cmd/api/main.go          entrypoint: config, migrations, DB pool, HTTP server
  internal/domain/         plain structs shared across layers
  internal/db/             pgxpool + hand-written repository queries
  internal/service/        business logic (value derivation, dashboard/progress aggregation, targets)
  internal/httpapi/        chi router, handlers, middleware
  migrations/               goose SQL migrations (embedded into the binary)

frontend/
  src/pages/                one folder per screen (Dashboard, Assets, Debts, PassiveIncome, Targets, Progress, Rates)
  src/components/           Sidebar/BottomTabBar shell, Modal, DonutChart, LineChart
  src/hooks/                one react-query hook per API resource
  src/lib/                  format.ts (money formatting), holdingCalc.ts (client-side value preview), api.ts

docker-compose.yml          postgres + api + web (Caddy)
```

## What's not built yet

- No CSV/spreadsheet import — the user chose to re-enter historical snapshots manually rather than build one-time import tooling.

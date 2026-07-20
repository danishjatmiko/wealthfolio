# Wealthfolio

A personal net-worth / asset-portfolio tracker — replaces a monthly spreadsheet workflow with a real app you can update from your phone or laptop. PostgreSQL + Go backend + React frontend, built from the design in [`design_handoff_portfolio_app/`](design_handoff_portfolio_app/).

## Stack

- **Database**: PostgreSQL (schema + seed in [`backend/migrations/`](backend/migrations/), applied automatically on API startup via [goose](https://github.com/pressly/goose))
- **Backend**: Go, `chi` router, `pgx` (no ORM — hand-written SQL in [`backend/internal/db/`](backend/internal/db/))
- **Frontend**: React + TypeScript + Vite, `react-router`, `@tanstack/react-query`, hand-rolled SVG charts
- **Deploy**: Docker Compose (Postgres + API + Caddy serving the built frontend and reverse-proxying `/api/*`)

## Important: money unit convention

Every monetary field in the database and API (`value_idr`, `target_value`, `per_year_idr`, gold prices, etc.) is an integer in **thousands of IDR** — e.g. `3750000` means Rp 3,750,000,000. This matches the design prototype's internal unit exactly, so the `money()`/`goldPrice()`/derivation formulas could be ported verbatim. The one exception is `rate_entries.usd_idr`, which is full IDR per 1 USD. See the comment at the top of [`backend/migrations/00001_init.sql`](backend/migrations/00001_init.sql).

## Single-user now, multi-user-ready

v1 has no login screen — every table has a `user_id` column, but the backend's `CurrentUserMiddleware` ([`backend/internal/httpapi/middleware.go`](backend/internal/httpapi/middleware.go)) just injects one fixed seeded user (`00000000-0000-0000-0000-000000000001`) into every request. Adding real auth later means replacing that one middleware plus adding login endpoints — no schema or handler changes needed.

**Security note**: because there's no login, anyone with the URL can view and edit your financial data once this is on a public server. At minimum keep the URL unguessable, or put Caddy `basicauth` in front of it (see `frontend/Caddyfile`) until real per-user auth exists.

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
- No authentication (see above).

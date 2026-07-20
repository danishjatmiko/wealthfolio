# Wealthfolio — Codebase Documentation

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture](#2-architecture)
3. [Directory Structure](#3-directory-structure)
4. [Backend (Go)](#4-backend-go)
   - [Entry Point](#41-entry-point)
   - [Configuration](#42-configuration)
   - [Domain Layer](#43-domain-layer)
   - [Database Layer](#44-database-layer)
   - [Service Layer](#45-service-layer)
   - [HTTP API Layer](#46-http-api-layer)
   - [Migrations](#47-migrations)
5. [Database Schema](#5-database-schema)
6. [API Reference](#6-api-reference)
7. [Business Rules](#7-business-rules)
8. [Frontend (React + TypeScript)](#8-frontend-react--typescript)
   - [Pages](#81-pages)
   - [Hooks](#82-hooks)
   - [Components](#83-components)
   - [Contexts](#84-contexts)
   - [Utility Libraries](#85-utility-libraries)
9. [Money Unit Convention](#9-money-unit-convention)
10. [Deployment](#10-deployment)

---

## 1. Project Overview

**Wealthfolio** is a personal net-worth and wealth-tracking application designed for Indonesian users. It lets a single user record monthly portfolio snapshots, track debts, log passive income sources, and set financial targets — all denominated in Indonesian Rupiah (IDR), with special handling for gold prices and USD-denominated assets.

**Tech stack:**

| Layer | Technology |
|---|---|
| Backend | Go 1.23, chi router, pgx v5 |
| Database | PostgreSQL 16 |
| Migrations | Goose |
| Frontend | React 18, TypeScript, Vite |
| Serving | Caddy (reverse proxy + TLS) |
| Orchestration | Docker Compose |

---

## 2. Architecture

```
┌─────────────────────────────────────────────┐
│  Browser (React + TypeScript, Vite)         │
│  Pages → Hooks → lib/api.ts → /api/v1       │
└───────────────────┬─────────────────────────┘
                    │ HTTP (JSON)
┌───────────────────▼─────────────────────────┐
│  Caddy (reverse proxy / TLS)                │
│  /api/v1/* → api:8080                       │
│  /*        → static frontend assets         │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│  Go API Server (cmd/api/main.go)            │
│                                             │
│  httpapi layer (chi router, handlers)       │
│       ↓                    ↓                │
│  service layer       db (repos) layer       │
│  (business logic)    (SQL via pgx)          │
│                            ↓                │
│            domain layer (shared types)      │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│  PostgreSQL 16                              │
└─────────────────────────────────────────────┘
```

The backend follows a strict four-layer dependency order:

```
domain  ←  db  ←  service  ←  httpapi
```

- `domain` — pure data structs; no dependencies.
- `db` — repository structs that issue SQL and return domain types.
- `service` — business logic that orchestrates repositories; no HTTP awareness.
- `httpapi` — HTTP handlers that parse requests, call services/repos, and write JSON responses.

---

## 3. Directory Structure

```
wealth_management/
├── docker-compose.yml          # Full-stack orchestration
├── .env.example                # Environment variable template
├── backend/
│   ├── cmd/api/main.go         # Server entry point
│   ├── go.mod / go.sum
│   ├── Dockerfile
│   ├── internal/
│   │   ├── config/config.go    # Env var loading
│   │   ├── domain/
│   │   │   ├── domain.go       # Shared data types (Category, Holding, Debt…)
│   │   │   └── date.go         # Custom Date type (YYYY-MM-DD, SQL/JSON aware)
│   │   ├── db/
│   │   │   ├── repos.go        # Repos bundle + ErrNotFound sentinel
│   │   │   ├── pool.go         # pgxpool setup
│   │   │   ├── migrate.go      # Goose migration runner
│   │   │   ├── categories.go
│   │   │   ├── rates.go
│   │   │   ├── snapshots.go
│   │   │   ├── holdings.go
│   │   │   ├── debts.go
│   │   │   ├── passive_income.go
│   │   │   └── targets.go
│   │   ├── service/
│   │   │   ├── services.go     # Services bundle
│   │   │   ├── valuation.go    # ComputeHoldingValue (gold / USD / IDR)
│   │   │   ├── calc.go         # netEquity / investedTotal / percentOf helpers
│   │   │   ├── holdings.go     # HoldingsService (create/update/delete + lock check)
│   │   │   ├── snapshots.go    # SnapshotsService (summaries, detail, create)
│   │   │   ├── dashboard.go    # DashboardService (equity, debt, passive aggregates)
│   │   │   ├── progress.go     # ProgressService (time-series at monthly/Q/Y)
│   │   │   ├── targets.go      # TargetsService (current_value / percent computation)
│   │   │   ├── errors.go       # Domain-level sentinel errors
│   │   │   └── valuation_test.go
│   │   └── httpapi/
│   │       ├── router.go       # Chi router + CORS + middleware wiring
│   │       ├── handler.go      # Handler struct
│   │       ├── middleware.go   # CurrentUserMiddleware (injects fixed user ID)
│   │       ├── respond.go      # writeJSON / writeError helpers
│   │       ├── errors.go       # HTTP error mapping (service/db errors → status codes)
│   │       ├── categories.go
│   │       ├── rates.go
│   │       ├── snapshots.go
│   │       ├── holdings.go
│   │       ├── debts.go
│   │       ├── passive_income.go
│   │       ├── targets.go
│   │       ├── dashboard.go
│   │       └── progress.go
│   └── migrations/
│       ├── embed.go            # //go:embed *.sql
│       ├── 00001_init.sql      # Full schema creation
│       └── 00002_seed.sql      # Default user + categories
└── frontend/
    ├── index.html
    ├── package.json
    ├── vite.config.ts
    ├── Dockerfile
    ├── Caddyfile
    └── src/
        ├── main.tsx            # React root, providers
        ├── App.tsx             # BrowserRouter + routes
        ├── types.ts            # All TypeScript types mirroring backend contract
        ├── lib/
        │   ├── api.ts          # Typed fetch wrapper + api.* call objects
        │   ├── format.ts       # fmtIdr / money / goldFmt / usdFmt / parseNumeric
        │   ├── holdingCalc.ts  # Client-side value preview (mirrors backend logic)
        │   └── colors.ts       # Design token colours
        ├── context/
        │   ├── MoneyVisibilityContext.tsx  # Global hide/show values toggle
        │   └── ToastContext.tsx            # Toast notification system
        ├── hooks/              # One SWR/fetch hook per API resource
        ├── components/
        │   ├── layout/AppShell.tsx  # Sidebar + header + bottom nav wrapper
        │   ├── charts/DonutChart.tsx
        │   ├── charts/LineChart.tsx
        │   └── Modal.tsx
        ├── pages/
        │   ├── Dashboard.tsx
        │   ├── assets/         # Assets page + AssetModal + NewSnapshotModal
        │   ├── debts/          # Debts page + DebtModal
        │   ├── passive/        # Passive Income page + PassiveIncomeModal
        │   ├── targets/        # Targets page + TargetModal
        │   ├── progress/       # Progress page (line chart)
        │   └── rates/          # Rates page
        └── styles/             # Global CSS, design tokens, component helpers
```

---

## 4. Backend (Go)

### 4.1 Entry Point

**`backend/cmd/api/main.go`**

Startup sequence:
1. Load config from environment (`config.Load()`).
2. Run pending SQL migrations via Goose (`db.RunMigrations`).
3. Open a `pgxpool` connection pool (`db.NewPool`).
4. Build the repository bundle (`db.NewRepos`).
5. Build the service bundle (`service.NewServices`).
6. Build the chi router (`httpapi.NewRouter`).
7. Start `http.ListenAndServe` on `:PORT`.

---

### 4.2 Configuration

**`backend/internal/config/config.go`**

| Env Var | Default | Description |
|---|---|---|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `PORT` | `8080` | HTTP listen port |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |

---

### 4.3 Domain Layer

**`backend/internal/domain/domain.go`**

All plain data structs shared across layers. No logic, no SQL, no HTTP.

| Struct | Mirrors DB Table | Key Fields |
|---|---|---|
| `Category` | `categories` | `key`, `label`, `color_oklch`, `kind` (asset/liability), `price_linked`, `sort_order` |
| `RateEntry` | `rate_entries` | `entry_date`, `antam`, `kinghalim`, `ubs`, `usd_idr` |
| `Snapshot` | `snapshots` | `snapshot_date` |
| `Holding` | `holdings` | `category_key`, `value_idr`, `is_liability`, `gram`, `qty`, `brand`, `usd_value`, `currency` |
| `Debt` | `debts` | `name`, `type`, `value_idr`, `direction` (i_owe / owed_to_me) |
| `PassiveIncomeSource` | `passive_income_sources` | `name`, `per_year_idr` |
| `Target` | `targets` + computed | `metric_type`, `target_value`, `current_value`, `percent`, `lower_is_better` |

**`backend/internal/domain/date.go`**

`domain.Date` wraps `time.Time` so that all date fields serialize/deserialize as bare `"YYYY-MM-DD"` strings in JSON, and scan/value correctly against PostgreSQL `date` columns. It implements `json.Marshaler`, `json.Unmarshaler`, `sql.Scanner`, and `driver.Valuer`.

---

### 4.4 Database Layer

**`backend/internal/db/repos.go`**

`Repos` is a bundle of all seven repository structs plus the raw `pgxpool.Pool` (exposed for transactions if needed). Constructed once in `main.go` and passed through the whole stack.

```
Repos
├── CategoriesRepo   — List, GetByID
├── RatesRepo        — List, GetLatest, Create
├── SnapshotsRepo    — ListWithAgg, ListWithAggAsc, GetLatest, GetByDate, Create
├── HoldingsRepo     — ListBySnapshot, GetByID, Create, Update, Delete, CopyFromSnapshot
├── DebtsRepo        — List, SumByDirection, Create, Update, Delete
├── PassiveIncomeRepo — List, Sum, Create, Update, Delete
└── TargetsRepo      — List, FirstTargetValueByMetricType, Create, Update, Delete
```

`db.ErrNotFound` is a package-level sentinel used by all repos when a query returns no row (`pgx.ErrNoRows` is normalized here so upper layers never import pgx).

---

### 4.5 Service Layer

**`backend/internal/service/services.go`**

`Services` bundles:

| Field | Type | Responsibility |
|---|---|---|
| `Holdings` | `*HoldingsService` | Create/update/delete holdings with lock check + value derivation |
| `Snapshots` | `*SnapshotsService` | Snapshot listing, creation, copy-from-latest |
| `Dashboard` | `*DashboardService` | Aggregate equity/debt/passive payload |
| `Progress` | `*ProgressService` | Net-equity time series at monthly/quarterly/yearly |
| `Targets` | `*TargetsService` | Target CRUD + current_value computation |

---

#### HoldingsService (`service/holdings.go`)

- **Snapshot lock check**: before any write (create/update/delete), it verifies the target holding belongs to the user's *latest* snapshot. If not, it returns `ErrSnapshotLocked`.
- **Value derivation**: calls `ComputeHoldingValue` (see `valuation.go`) which applies gold/USD/IDR logic.
- `CreateUnlocked` bypasses the lock check — used only by `SnapshotsService.Create` when populating a freshly-created (possibly backfilled) snapshot.

---

#### SnapshotsService (`service/snapshots.go`)

- `ListSummaries` — returns all snapshots newest-first, with `is_editable: true` only for index 0 (the latest).
- `Create` — creates a new snapshot; optionally copies all holdings from the previous latest (`copy_from_latest`); optionally populates a set of `initialHoldings` directly (bypasses the lock rule, enabling backfill).
- `is_editable` is always computed dynamically from `MAX(snapshot_date)`, not stored.

---

#### Valuation (`service/valuation.go`)

`ComputeHoldingValue(categoryKey, input, rate)` determines `value_idr` and the display `detail` string:

| Category | Input | Derived Value |
|---|---|---|
| `logam_mulia` (gold) | gram + qty + brand | `gram × qty × goldPricePerGram(brand, rate)` |
| `bonds_usd` | usd_value | `usd_value × (rate.UsdIdr / 1000)` |
| `uang_tunai` (USD) | usd_value, currency=USD | `usd_value × (rate.UsdIdr / 1000)` |
| Everything else | value_idr | used as-is |

Returns `ErrNoRateEntry` if a rate-dependent computation is requested but no rate entry exists yet.

---

#### DashboardService (`service/dashboard.go`)

Assembles the `GET /dashboard` payload:

- **equity**: `total_idr` = assets + liabilities (liabilities stored as positive, represent negative impact), `invested_idr` = assets only, `incl_passive_idr` = equity + annual passive income, `mom_change_idr/pct` = month-on-month delta vs previous snapshot.
- **debt**: totals for `i_owe` and `owed_to_me` debts, `ratio_pct` = debt / total equity.
- **passive**: sum of `per_year_idr` across all sources vs a target, monthly equivalents.
- **allocation**: per-category breakdown sorted by `sort_order`, with percent of invested total.

---

#### ProgressService (`service/progress.go`)

`GET /progress?granularity=monthly|quarterly|yearly`

Fetches all snapshots in ascending order and builds a time series of `net_equity_idr` per period:
- **monthly**: one point per snapshot.
- **quarterly**: collapses to last snapshot in each `Q1/Q2/Q3/Q4` of a year.
- **yearly**: collapses to last snapshot in each year.

Returns `delta_idr` and `delta_pct` comparing the two most recent points in the series.

---

#### TargetsService (`service/targets.go`)

Supported `metric_type` values and how `current_value` is computed:

| metric_type | current_value source |
|---|---|
| `equity` | `netEquity(latestHoldings)` — sum of all holdings (assets + liabilities) |
| `gold_grams` | sum of `gram × qty` for all `logam_mulia` holdings in latest snapshot |
| `passive_income` | `PassiveIncomeRepo.Sum` |
| `debt_ratio` | `debtIowe / netEquity × 100` |
| `custom` | stored `manual_current_value` field |

`lower_is_better` is `true` only for `debt_ratio`.

---

### 4.6 HTTP API Layer

**`backend/internal/httpapi/router.go`**

Chi middleware stack applied globally:
1. `chimiddleware.Logger` — request logging.
2. `chimiddleware.Recoverer` — panic recovery.
3. `cors.Handler` — restricts origins to `cfg.CORSOrigin`.

All `/api/v1` routes additionally use `CurrentUserMiddleware` which injects the fixed user ID `00000000-0000-0000-0000-000000000001` into the request context.

**`backend/internal/httpapi/errors.go`**

Service/db errors are mapped to HTTP status codes:

| Error | HTTP Status |
|---|---|
| `db.ErrNotFound` | 404 Not Found |
| `service.ErrSnapshotLocked` | 409 Conflict |
| `service.ErrSnapshotDateExists` | 409 Conflict |
| `service.ErrInvalidCategory` | 422 Unprocessable Entity |
| `service.ErrNoRateEntry` | 422 Unprocessable Entity |
| `service.ErrInvalidInput` | 422 Unprocessable Entity |
| anything else | 500 Internal Server Error |

---

### 4.7 Migrations

Located in `backend/migrations/`, embedded into the binary via `//go:embed *.sql` and run at startup using [Goose](https://github.com/pressly/goose).

| File | Contents |
|---|---|
| `00001_init.sql` | Full schema: users, categories, rate_entries, snapshots, holdings, debts, passive_income_sources, targets |
| `00002_seed.sql` | Default user (fixed UUID) + 9 seeded categories |

---

## 5. Database Schema

```
users
  id uuid PK
  email text
  display_name text
  created_at timestamptz

categories
  id smallserial PK
  key text UNIQUE        -- e.g. "logam_mulia", "saham"
  label text             -- e.g. "Logam Mulia", "Saham"
  color_oklch text
  kind text              -- 'asset' | 'liability'
  price_linked bool      -- true if value is derived from a rate entry
  sort_order smallint

rate_entries
  id uuid PK
  user_id uuid FK → users
  entry_date date
  antam numeric(14,2)      -- gold price Antam (thousands IDR / gram)
  kinghalim numeric(14,2)  -- gold price King Halim (thousands IDR / gram)
  ubs numeric(14,2)        -- gold price UBS (thousands IDR / gram)
  usd_idr numeric(14,2)   -- full IDR per 1 USD (NOT thousands)
  UNIQUE(user_id, entry_date)

snapshots
  id uuid PK
  user_id uuid FK → users
  snapshot_date date
  UNIQUE(user_id, snapshot_date)

holdings
  id uuid PK
  snapshot_id uuid FK → snapshots (CASCADE DELETE)
  category_id smallint FK → categories
  name text
  detail text
  value_idr bigint        -- THOUSANDS of IDR
  is_liability bool
  gram numeric(14,3)      -- for gold
  qty numeric(14,3)       -- number of gold bars/coins
  brand text              -- 'Antam' | 'King Halim' | 'UBS'
  usd_value numeric(14,2) -- for USD-denominated assets
  currency text           -- 'IDR' | 'USD'

debts
  id uuid PK
  user_id uuid FK → users
  name text
  type text
  value_idr bigint
  direction text          -- 'i_owe' | 'owed_to_me'

passive_income_sources
  id uuid PK
  user_id uuid FK → users
  category_id smallint FK → categories
  name text
  per_year_idr bigint     -- THOUSANDS of IDR per year

targets
  id uuid PK
  user_id uuid FK → users
  name text
  year int
  metric_type text        -- 'equity'|'gold_grams'|'passive_income'|'debt_ratio'|'custom'
  target_value numeric(18,2)
  unit text
  manual_current_value numeric(18,2)  -- used only for metric_type='custom'
```

**Seeded categories (sort_order):**

| # | key | label | kind |
|---|---|---|---|
| 1 | `logam_mulia` | Logam Mulia | asset |
| 2 | `saham` | Saham | asset |
| 3 | `bonds_usd` | Bonds USD | asset |
| 4 | `uang_tunai` | Uang Tunai | asset |
| 5 | `us_etf` | US ETF | asset |
| 6 | `properti` | Properti | asset |
| 7 | `crypto` | Crypto | asset |
| 8 | `reksa_dana` | Reksa Dana | asset |
| 9 | `liabilitas` | Liabilitas | liability |

---

## 6. API Reference

Base URL: `/api/v1`

All endpoints except `/healthz` require the `CurrentUserMiddleware` (user ID is implicit — no auth token needed in the current single-user design).

### Health

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Returns `{"status":"ok"}` |

### Categories

| Method | Path | Description |
|---|---|---|
| GET | `/categories` | List all categories |

### Rate Entries

| Method | Path | Description |
|---|---|---|
| GET | `/rates` | List all rate entries |
| GET | `/rates/latest` | Latest rate entry |
| POST | `/rates` | Create a rate entry |

**POST /rates body:**
```json
{
  "entry_date": "2025-07-01",
  "antam": 1650,
  "kinghalim": 1640,
  "ubs": 1630,
  "usd_idr": 16350
}
```

### Snapshots

| Method | Path | Description |
|---|---|---|
| GET | `/snapshots` | List all snapshot summaries (newest first) |
| GET | `/snapshots/latest` | Latest snapshot with all holdings |
| POST | `/snapshots` | Create a new snapshot |
| GET | `/snapshots/{date}` | Get snapshot for a specific date (YYYY-MM-DD) |
| GET | `/snapshots/{date}/holdings` | List holdings for a date |
| POST | `/snapshots/{date}/holdings` | Add a holding to the snapshot on that date |

**POST /snapshots body:**
```json
{
  "snapshot_date": "2025-07-01",
  "copy_from_latest": true,
  "initial_holdings": []
}
```

### Holdings

| Method | Path | Description |
|---|---|---|
| PUT | `/holdings/{id}` | Update a holding |
| DELETE | `/holdings/{id}` | Delete a holding |

**POST/PUT holding body:**
```json
{
  "category_id": 1,
  "name": "Antam 10g",
  "gram": 10,
  "qty": 3,
  "brand": "Antam"
}
```
Value is derived server-side from the latest rate entry; `value_idr` only needed for manual/fallback.

### Debts

| Method | Path | Description |
|---|---|---|
| GET | `/debts` | List all debts |
| POST | `/debts` | Create a debt |
| PUT | `/debts/{id}` | Update a debt |
| DELETE | `/debts/{id}` | Delete a debt |

**POST/PUT body:**
```json
{
  "name": "KPR BCA",
  "type": "mortgage",
  "value_idr": 750000,
  "direction": "i_owe"
}
```

### Passive Income

| Method | Path | Description |
|---|---|---|
| GET | `/passive-income` | List all passive income sources |
| POST | `/passive-income` | Create a source |
| PUT | `/passive-income/{id}` | Update a source |
| DELETE | `/passive-income/{id}` | Delete a source |

### Targets

| Method | Path | Description |
|---|---|---|
| GET | `/targets` | List all targets with computed `current_value`/`percent` |
| POST | `/targets` | Create a target |
| PUT | `/targets/{id}` | Update a target |
| DELETE | `/targets/{id}` | Delete a target |

**POST/PUT body:**
```json
{
  "name": "Equity 2025",
  "year": 2025,
  "metric_type": "equity",
  "target_value": 5000000,
  "unit": "IDR"
}
```

### Dashboard

| Method | Path | Description |
|---|---|---|
| GET | `/dashboard` | Full dashboard payload |

**Response shape:**
```json
{
  "equity": {
    "total_idr": 4200000,
    "invested_idr": 4000000,
    "incl_passive_idr": 4500000,
    "mom_change_idr": 200000,
    "mom_change_pct": 5.0,
    "by_category": [{ "category_key": "saham", "label": "Saham", "value_idr": 1500000, "percent": 37.5, "color_oklch": "..." }]
  },
  "debt": {
    "total_debt_idr": 750000,
    "total_receivable_idr": 100000,
    "ratio_pct": 17.86
  },
  "passive": {
    "per_year_idr": 300000,
    "target_per_year_idr": 600000,
    "percent": 50.0,
    "per_month_idr": 25000,
    "per_month_target_idr": 50000
  },
  "allocation": [ ... ]
}
```

### Progress

| Method | Path | Description |
|---|---|---|
| GET | `/progress?granularity=monthly` | Net-equity time series |

`granularity`: `monthly` (default) | `quarterly` | `yearly`

---

## 7. Business Rules

### Snapshot Immutability
Only the snapshot with the latest `snapshot_date` is editable. All write operations on holdings (create, update, delete) check this before proceeding. Attempting to mutate a historical snapshot returns `409 Conflict`.

### Value Derivation
Asset values are never stored as a raw number typed by the user for price-linked categories. Instead:
- **Gold** (`logam_mulia`): `gram × qty × goldPricePerGram(brand, latestRateEntry)` (thousands IDR)
- **USD Bonds** (`bonds_usd`) and **USD Cash** (`uang_tunai` + `currency=USD`): `usd_value × (usd_idr / 1000)`
- If no rate entry exists and a derived value is required, the API returns `422 Unprocessable Entity`.

### Single-User Design
The application is currently single-user. The fixed user ID `00000000-0000-0000-0000-000000000001` is seeded in `00002_seed.sql` and injected by `CurrentUserMiddleware`. There is no authentication layer.

### Backfilling Snapshots
When creating a snapshot with a date *before* the current latest, the resulting snapshot is immediately locked (`is_editable: false`). Initial holdings can still be supplied via `initial_holdings[]` in the create request — this is the only path for writing to a non-latest snapshot.

---

## 8. Frontend (React + TypeScript)

### 8.1 Pages

| Path | Component | Description |
|---|---|---|
| `/` | `Dashboard` | Equity/debt/passive summary cards + allocation donut chart |
| `/assets` | `Assets` | Snapshot selector, holdings table with category filter chips, add/edit modal |
| `/debts` | `Debts` | Debt list split into "I Owe" and "Owed to Me" |
| `/passive-income` | `PassiveIncome` | Annual passive income sources, yearly/monthly totals |
| `/targets` | `Targets` | Financial goals with progress bars |
| `/progress` | `Progress` | Line chart of net equity over time (monthly/quarterly/yearly toggle) |
| `/rates` | `Rates` | Gold price and USD/IDR rate entry history + add form |

All pages are rendered inside `AppShell`, which provides the sidebar navigation (desktop) and bottom tab bar (mobile).

---

### 8.2 Hooks

Each resource has a dedicated custom hook in `frontend/src/hooks/`:

| Hook | Fetches |
|---|---|
| `useCategories` | `GET /categories` |
| `useRates` / `useLatestRate` | `GET /rates`, `GET /rates/latest` |
| `useSnapshots` / `useLatestSnapshot` / `useSnapshotByDate` | Snapshot endpoints |
| `useHoldings` | Holdings on a given snapshot date |
| `useDebts` | `GET /debts` |
| `usePassiveIncome` | `GET /passive-income` |
| `useTargets` | `GET /targets` |
| `useDashboard` | `GET /dashboard` |
| `useProgress` | `GET /progress?granularity=…` |

---

### 8.3 Components

| Component | Purpose |
|---|---|
| `AppShell` | Layout wrapper: sidebar, header (date + hide-toggle), bottom nav, `<Outlet />` |
| `Modal` | Generic accessible modal container |
| `AssetModal` | Add/edit holding form (adapts fields based on category) |
| `NewSnapshotModal` | Create new snapshot (with copy-from-latest option) |
| `DebtModal` | Add/edit debt form |
| `PassiveIncomeModal` | Add/edit passive income source |
| `TargetModal` | Add/edit target form |
| `DonutChart` | SVG donut chart for portfolio allocation |
| `LineChart` | SVG line chart for progress trend |

---

### 8.4 Contexts

**`MoneyVisibilityContext`** (`context/MoneyVisibilityContext.tsx`)

Global toggle that hides or reveals all monetary values. Persisted to `localStorage` under key `wealthfolio:hideValues`. Exposes:
- `hidden: boolean`
- `toggle: () => void`
- `fmt: (value: number) => string` — formats or masks a thousands-of-IDR amount

**`ToastContext`** (`context/ToastContext.tsx`)

Provides `showToast(message, type)` for success/error notifications.

---

### 8.5 Utility Libraries

**`lib/format.ts`**

| Function | Description |
|---|---|
| `fmtIdr(value)` | Format thousands-IDR integer → Indonesian shortened form (`Rp3.75 M`, `Rp202 jt`, `Rp16.8 jt`, `Rp800 rb`) |
| `money(value, hidden)` | Hide-aware wrapper around `fmtIdr` — returns `"Rp ••••"` when hidden |
| `goldFmt(value)` | Format thousands-IDR/gram gold price → `"Rp2.65 jt/g"` |
| `usdFmt(value)` | Format full-IDR USD rate → `"Rp18,100"` |
| `parseNumeric(input)` | Parse free-typed numeric string (strips non-numeric chars) |

**`lib/api.ts`**

Typed fetch wrapper. All calls go through `request<T>()` which:
1. Sends `Content-Type: application/json`.
2. On non-2xx, throws `ApiError(status, message)` (extracts `message` or `error` from JSON body).
3. On 204, returns `undefined`.

The `api` object groups every endpoint call by resource, e.g. `api.holdings.create(date, input)`.

**`lib/holdingCalc.ts`**

Client-side mirror of the backend's `ComputeHoldingValue`. Used for live preview inside the AssetModal form — calculates the displayed IDR value as the user types gram/usd fields, before the form is submitted. Also provides form-field visibility helpers (`isGoldCategory`, `showsUsdInput`, etc.) and `prefillFromHolding` for editing an existing holding.

---

## 9. Money Unit Convention

> **All `*_idr` / `value_idr` / `target_value` integers represent THOUSANDS of IDR.**
> Example: `3_750_000` means Rp 3,750,000,000 (3.75 billion rupiah).

The only exception is `rate_entries.usd_idr`, which is **full IDR per 1 USD** (not thousands), matching common exchange-rate quoting convention.

Gold prices (`antam`, `kinghalim`, `ubs`) are **thousands of IDR per gram** — consistent with the general `*_idr` convention.

This convention is enforced in:
- The SQL schema comment in `00001_init.sql`
- The `domain.RateEntry` Go docstring
- The `types.ts` file header comment
- The `lib/format.ts` file header comment

---

## 10. Deployment

### Docker Compose

Three services defined in `docker-compose.yml`:

| Service | Image | Role |
|---|---|---|
| `db` | `postgres:16-alpine` | PostgreSQL database, data persisted in `db_data` volume |
| `api` | Built from `./backend/Dockerfile` | Go API server on port 8080 (internal only) |
| `web` | Built from `./frontend/Dockerfile` | Caddy serving static frontend + reverse-proxying `/api/v1` to `api:8080` |

The `web` service exposes ports `80` and `443`. Caddy handles TLS automatically when `DOMAIN` is set to a real hostname.

### Environment Variables

Copy `.env.example` to `.env` and set:

| Variable | Example | Description |
|---|---|---|
| `POSTGRES_USER` | `wealthfolio` | DB username |
| `POSTGRES_PASSWORD` | *(secret)* | DB password |
| `POSTGRES_DB` | `wealthfolio` | DB name |
| `CORS_ORIGIN` | `https://yourdomain.com` | Allowed frontend origin |
| `DOMAIN` | `yourdomain.com` | Caddy domain for TLS (leave as `:80` for HTTP-only) |

### Local Development

**Backend:**
```bash
cd backend
DATABASE_URL="postgres://wealthfolio:wealthfolio@localhost:5432/wealthfolio?sslmode=disable" go run ./cmd/api
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev   # starts Vite dev server on http://localhost:5173
```

The Vite dev server proxies `/api/v1` to `http://localhost:8080` (configured in `vite.config.ts`).

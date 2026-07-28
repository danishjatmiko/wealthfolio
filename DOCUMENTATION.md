# Etherna — Codebase Documentation

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture](#2-architecture)
3. [Directory Structure](#3-directory-structure)
4. [Backend (Go)](#4-backend-go)
   - [Entry Points](#41-entry-points)
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
9. [Android App (Kotlin + Jetpack Compose)](#9-android-app-kotlin--jetpack-compose)
10. [Money Unit Convention](#10-money-unit-convention)
11. [Deployment](#11-deployment)

---

## 1. Project Overview

**Etherna** is a personal net-worth, portfolio, and monthly-expense tracking application for a single household, denominated in Indonesian Rupiah with special handling for gold prices and USD-denominated assets. It has three parts:

- A **Go + PostgreSQL backend** exposing a JSON API.
- A **React web frontend** — the primary UI for everything: portfolio snapshots, debts, monthly expenses/budget envelopes, passive income, targets, and gold/USD rate history.
- An **Android app** that hosts that same web frontend in a WebView (no separate native UI for the actual data screens) and additionally runs a `NotificationListenerService` that watches for GoPay/DANA/BCA/Bank Jago payment notifications and auto-forwards them to the backend, which turns them into expense records without the user typing anything in.

**Tech stack:**

| Layer | Technology |
|---|---|
| Backend | Go 1.25, chi router, pgx v5 |
| Database | PostgreSQL |
| Migrations | Goose |
| Web frontend | React 18, TypeScript, Vite, TanStack Query |
| Android | Kotlin, Jetpack Compose, Hilt, Room, WorkManager |
| Serving | Caddy (reverse proxy + TLS) |
| Orchestration | Docker Compose |

---

## 2. Architecture

```
┌──────────────────────────┐   ┌───────────────────────────────────────┐
│  Browser (React + TS)    │   │  Android app                          │
│  Pages → Hooks → api.ts  │   │  ┌─────────────────────────────────┐  │
└────────────┬──────────────┘   │  │ WebView hosting the same        │  │
             │                  │  │ React frontend (via WEB_ORIGIN) │  │
             │ HTTP (JSON)      │  └─────────────────────────────────┘  │
             │                  │  NotificationListenerService          │
             │                  │  → Room outbox → WorkManager sync     │
             │                  └──────────────────┬────────────────────┘
             │                                     │ HTTP (JSON, API_BASE_URL)
┌────────────▼─────────────────────────────────────▼────────────────────┐
│  Caddy (reverse proxy / TLS)                                          │
│  /api/v1/* → api:8080         /* → static frontend assets             │
└────────────────────────────────────┬───────────────────────────────────┘
                                     │
┌────────────────────────────────────▼───────────────────────────────────┐
│  Go API Server (cmd/api/main.go)                                       │
│                                                                         │
│  httpapi layer (chi router, handlers, auth middleware)                 │
│       ↓                              ↓                                 │
│  service layer                 db (repos) layer                        │
│  (business logic,              (SQL via pgx)                           │
│   notificationparse)                 ↓                                 │
│                     domain layer (shared types)                        │
└────────────────────────────────────┬───────────────────────────────────┘
                                     │
┌────────────────────────────────────▼───────────────────────────────────┐
│  PostgreSQL                                                             │
└──────────────────────────────────────────────────────────────────────┘

  (separately) cmd/rates-sync — a standalone cron script, no relation to
  cmd/api's process — scrapes gold/USD rates daily and POSTs to /api/v1/rates
  using a shared-secret header instead of a session.
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
├── keystore.properties.example # Android release-signing template
├── backend/
│   ├── cmd/
│   │   ├── api/main.go         # Server entry point
│   │   └── rates-sync/main.go  # Standalone daily cron script (gold/USD scraper)
│   ├── go.mod / go.sum
│   ├── internal/
│   │   ├── config/config.go    # Env var loading
│   │   ├── domain/
│   │   │   ├── domain.go       # Shared data types
│   │   │   └── date.go         # Custom Date type (YYYY-MM-DD, SQL/JSON aware)
│   │   ├── db/                 # One repo file per resource (see §4.4)
│   │   ├── service/             # One service file per resource + notificationparse/ (see §4.5)
│   │   └── httpapi/             # chi router, handlers, middleware (see §4.6)
│   └── migrations/
│       ├── embed.go            # //go:embed *.sql
│       └── 00001..00018_*.sql  # see §4.7
└── frontend/
    ├── index.html / package.json / vite.config.ts / Dockerfile / Caddyfile
    └── src/
        ├── main.tsx / App.tsx / types.ts
        ├── lib/                 # api.ts, format.ts, holdingCalc.ts, colors.ts
        ├── context/              # AuthContext, MoneyVisibilityContext, ToastContext
        ├── hooks/                # one hook per API resource + usePullToRefresh
        ├── components/           # layout/AppShell, charts/, Modal
        └── pages/                # Dashboard, assets/, debts/, expenses/, passive/, targets/, progress/, rates/, auth/

android/
├── keystore.properties          # gitignored — real release signing config
├── release.keystore.jks         # gitignored — real signing key
└── app/src/main/kotlin/com/wealthfolio/mobile/
    ├── AppConfig.kt, MainActivity.kt, WealthfolioApp.kt
    ├── auth/                     # session storage, Google OAuth, password login
    ├── data/outbox/              # durable local queue of captured notifications
    ├── data/notificationcatalog/ # local mirror of the backend's supported-apps catalog
    ├── notifications/            # NotificationListenerService + idempotency-key builder
    ├── sync/                     # WorkManager scheduling + Sync Status screen
    ├── settings/                 # Settings screen
    ├── network/                  # Retrofit ApiService + DTOs
    ├── di/                       # Hilt modules
    ├── ui/                       # App shell composables, theme
    └── web/                      # WebView host + JS bridge
```

---

## 4. Backend (Go)

Module path `wealthfolio/backend`, Go 1.25. Key deps: `chi/v5` + `go-chi/cors`, `pgx/v5`, `pressly/goose/v3`, `google/uuid`, `golang.org/x/crypto` (Argon2id), `golang.org/x/oauth2`.

### 4.1 Entry Points

**`backend/cmd/api/main.go`** — the API server. Startup sequence:
1. Load config from environment (`config.Load()`); `log.Fatal` if `DATABASE_URL`, `GOOGLE_CLIENT_ID`, or `GOOGLE_CLIENT_SECRET` are empty.
2. Run pending SQL migrations via Goose, embedded in the binary (`db.RunMigrations`).
3. Open a `pgxpool` connection pool (`db.NewPool`, 15s connect timeout).
4. Build the repository bundle (`db.NewRepos`) → service bundle (`service.NewServices`) → chi router (`httpapi.NewRouter`).
5. Start `http.ListenAndServe` on `:PORT`.

**`backend/cmd/rates-sync/main.go`** — a *separate*, standalone binary, not started by `cmd/api` and not part of the API's process. Meant to run once daily via cron on the deploy host. Scrapes:
- Antam & UBS gold prices from indogold.id (session-cookie + CSRF-token-gated AJAX endpoint)
- King Halim gold price from kinghalim.com (scraped from rendered HTML)
- USD/IDR from Google Finance's quote page (plain HTML regex)

(`logammulia.com` was tried as a source but is blocked by Akamai's WAF for VPS IPs, hence the indogold.id fallback source for Antam/UBS.) If an individual scrape fails, it falls back to the previous day's value via `GET /rates/latest`, and only hard-fails if a scrape fails with no prior value to fall back to. Authenticates to the API via the `X-Rates-Sync-Token` header (see `RatesSyncOrAuthMiddleware` in §4.6) rather than a session login, then `POST`s the resolved values to `/api/v1/rates`. Env vars: `RATES_SYNC_TOKEN` (required), `ETHERNA_API_BASE_URL` (default `https://etherna.id/api/v1`).

---

### 4.2 Configuration

**`backend/internal/config/config.go`**

| Env Var | Default | Required? |
|---|---|---|
| `DATABASE_URL` | *(none)* | **Yes** — `cmd/api` fatals if empty |
| `GOOGLE_CLIENT_ID` | *(none)* | **Yes** — fatals if empty |
| `GOOGLE_CLIENT_SECRET` | *(none)* | **Yes** — fatals if empty |
| `PORT` | `8080` | No |
| `CORS_ORIGIN` | `http://localhost:5173` | No |
| `GOOGLE_REDIRECT_URL` | `http://localhost:8080/api/v1/auth/google/callback` | No |
| `APP_BASE_URL` | `http://localhost:5173` | No |
| `COOKIE_SECURE` | `true` iff `APP_BASE_URL` starts with `https://` | No — auto-derived, overridable |
| `RATES_SYNC_TOKEN` | *(empty)* | No — unset disables `RatesSyncOrAuthMiddleware`'s token path entirely |
| `RATES_SYNC_EMAIL` | *(empty)* | No — which account synced rates get attributed to |

---

### 4.3 Domain Layer

**`backend/internal/domain/domain.go`** — plain structs shared across every layer, no logic/SQL/HTTP.

| Struct | Mirrors Table | Key Fields |
|---|---|---|
| `User` | `users` | `Email *string`, `DisplayName`, `AvatarURL *string`, `GoogleSub *string`, `PasswordHash *string`, `CreatedAt` |
| `Category` | `categories` | `Key`, `Label`, `ColorOKLCH`, `Kind` (asset/liability), `PriceLinked`, `SortOrder` |
| `RateEntry` | `rate_entries` | `EntryDate`, `Antam`, `Kinghalim`, `Ubs`, `UsdIdr` — all full/raw IDR |
| `Snapshot` | `snapshots` | `SnapshotDate` |
| `Holding` | `holdings` (+ `categories` join) | `CategoryKey`, `CategoryLabel`, `Name`, `Detail`, `ValueIdr`, `IsLiability`, `Gram *float64`, `Qty *float64`, `Brand *string`, `UsdValue *float64`, `Currency *string` |
| `DebtSnapshot` | `debt_snapshots` | `SnapshotDate` |
| `DebtEntry` | `debt_entries` | `Name`, `Type`, `ValueIdr`, `Direction` (i_owe / owed_to_me) |
| `PassiveIncomeSource` | `passive_income_sources` (+ `categories` join) | `CategoryKey`, `CategoryLabel`, `Name`, `PerYearIdr` |
| `ExpensePeriod` | `expense_periods` | `StartDate`, `EndDate` |
| `BudgetEnvelope` | `budget_envelopes` | `PeriodID`, `Name`, `CommittedAmountIdr` |
| `FixedExpense` | `fixed_expenses` | `PeriodID`, `EnvelopeID`, `Name`, `AmountIdr`, `Source *string` (nil unless notification-created), `Notes *string` |
| `ExpenseSourceMapping` | `expense_source_mappings` | `Source`, `EnvelopeName` |
| `NotificationExpenseEvent` | `notification_expense_events` | `IdempotencyKey`, `Source`, `RawTitle/RawText/RawBigText *string`, `OccurredAt`, `ParseStatus`, `AmountIdr *int64`, `MerchantName *string`, `EnvelopeID/FixedExpenseID *uuid.UUID` |
| `NotificationApp` | `notification_apps` (+ `notification_app_packages` join) | `Source`, `DisplayName`, `PackageNames []string`, `AmountThousandSep`/`AmountDecimalSep` (backend-only, `json:"-"`), `Enabled` |
| `NotificationPattern` | `notification_patterns` | `AppID`, `Priority`, `Field`, `Regex`, `Description`, `Enabled` |
| `Target` | `targets` (+ computed) | `Name`, `Year`, `MetricType`, `TargetValue`, `Unit`, `ManualCurrentValue *float64` (json:"-"), plus server-computed `CurrentValue`/`Percent`/`LowerIsBetter` |

**`backend/internal/domain/date.go`** — `domain.Date` wraps `time.Time` so date fields serialize/deserialize as bare `"YYYY-MM-DD"` strings in JSON and scan/value correctly against Postgres `date` columns (`json.Marshaler`/`Unmarshaler`, `sql.Scanner`, `driver.Valuer`).

---

### 4.4 Database Layer

**`backend/internal/db/repos.go`** — `Repos` bundles every repository struct + the raw `pgxpool.Pool`. `db.ErrNotFound` is the package-level sentinel every repo normalizes `pgx.ErrNoRows` into, so upper layers never import pgx directly.

| Repo | Notable methods |
|---|---|
| `UsersRepo` (auth.go) | `GetByGoogleSub`, `GetByEmail`, `HasAnyGoogleUser`, `ClaimSeedUser`, `CreateUser` |
| `SessionsRepo` (auth.go) | `Create`, `GetByToken` (joined w/ user), `Refresh` (extend sliding expiry), `Delete` |
| `CategoriesRepo` | `List`, `GetByID` |
| `RatesRepo` | `List`, `GetLatest`, `Upsert` (on `(user_id, entry_date)` conflict) |
| `SnapshotsRepo` | `ListWithAgg`/`ListWithAggAsc`, `GetByID`/`GetByDate`/`GetLatest`, `Create`, `Delete` (soft) |
| `HoldingsRepo` | `ListBySnapshot`, `GetByID`, `Create`, `Update`, `Delete` |
| `DebtSnapshotsRepo` | `ListWithAgg`/`ListWithAggAsc` (+ `IOweIdr`/`OwedToMeIdr`), `GetByID`/`GetByDate`/`GetLatest`, `Create`, `Delete` (soft) |
| `DebtEntriesRepo` | `ListByDebtSnapshot`, `GetByID`, `Create`, `Update`, `Delete`, `CopyFromSnapshot`, `MaxUpdatedAt` |
| `ExpensePeriodsRepo` | `ListWithAgg`, `GetByID`/`GetByStartDate`/`GetLatest`, `Create`, `Delete` (hard — periods never lock) |
| `BudgetEnvelopesRepo` | `ListByPeriod`, `GetByID`, `Create`, `Update`, `Delete` (cascades fixed_expenses), `CopyFromPeriod` |
| `FixedExpensesRepo` | `ListByPeriod`, `GetByID`, `Create`, `Update`, `Delete` |
| `PassiveIncomeRepo` | `List`, `GetByID`, `Create`, `Update`, `Delete`, `Sum`, `MaxUpdatedAt` |
| `TargetsRepo` | `List`, `GetByID`, `Create`, `Update`, `Delete`, `FirstTargetValueByMetricType` |
| `ExpenseSourceMappingsRepo` | `ListByUser`, `GetBySource`, `Upsert` |
| `NotificationExpenseEventsRepo` | `GetByIdempotencyKey`, `CreateIgnored`, `CreateExpense` (fixed_expense + audit event atomically, idempotent, race-safe) |
| `NotificationAppsRepo` | `ListAll`, `GetBySource`, `ExistsSource`, `Create` (+ packages, transactional), `UpdateBySource` (full replace, transactional) |
| `NotificationPatternsRepo` | `ListByAppSource`, `ListAllEnabled` (warms the whole parse cache in one round trip), `GetByID`, `MaxPriorityForSource`, `Create`, `Update` |

---

### 4.5 Service Layer

**`backend/internal/service/services.go`** — `Services` bundles every service. **`errors.go`** defines every sentinel error and its intended HTTP status (see §4.6 for the actual mapping).

| Service | Responsibility & notable rules |
|---|---|
| `AuthService` (auth.go) | Google OAuth code exchange → upsert user → issue session; `PasswordLogin` (constant-time-cost dummy-hash comparison even for unknown emails, to avoid a timing side-channel); `Authenticate` validates + auto-refreshes sessions (sliding 7-day window, reissued when <3 days remain; 30-day absolute cap regardless of activity) |
| `password.go` | `HashPassword`/`VerifyPassword` — Argon2id, OWASP params, constant-time comparison |
| `HoldingsService` | Snapshot-lock check before every write (only the latest snapshot is editable); delegates value derivation to `ComputeHoldingValue`; `CreateUnlocked` is internal-only, used by `SnapshotsService` to seed a new snapshot |
| `SnapshotsService` | `ListSummaries` (`is_editable` true only for the single latest); `Create` (date must be today-or-later; optional copy-forward of the latest snapshot's holdings, **repriced** against current rates, not raw-copied); `Delete` is soft and allowed on any snapshot, not just the latest (deleting the current latest promotes the next-most-recent) |
| `ComputeHoldingValue` (valuation.go) | Gold (`gram × qty × GoldPricePerGram(brand, rate)`, unrecognized brand falls back to Antam) and USD-denominated categories (`usd_value × rate.usd_idr`) require a `RateEntry` — returns `ErrNoRateEntry` if none exists and no manual fallback given; everything else passes the user-entered value through unchanged |
| `DebtSnapshotsService` / `DebtEntriesService` | Same snapshot-locking pattern as assets, on its own independent timeline |
| `ExpensePeriodsService` | Implements the 25th-of-month pay-cycle rule (`boundsForPeriodMonth`); periods **never lock** — no latest-only write restriction, `Delete` hard-deletes any period; `Create` can `copyEnvelopes` (name + committed target only) forward from the current latest period |
| `BudgetEnvelopesService` / `FixedExpensesService` | Plain CRUD, no locking; `FixedExpensesService` validates an envelope belongs to the target period |
| `ExpenseSourceMappingsService` | `Upsert` validates the source exists in the `notification_apps` catalog (even if currently disabled) |
| `ExpenseIngestionService` | The notification pipeline entry point — see §7 |
| `NotificationCatalogService` | Bridges the DB-backed catalog to the pure `notificationparse` engine; owns an explicit-invalidation in-memory parse cache (no TTL — invalidated on every admin write, so an edited pattern is live immediately with no redeploy); admin CRUD for apps/patterns, validating regexes compile and contain a named `amount` capture group, auto-assigning `priority = max+10` when omitted |
| `notificationparse` package | Pure, DB-free: `Match(patterns, format, title, text, bigText)` tries patterns in ascending-priority order, first full match wins; `NormalizeAmount` strips everything but digits before the decimal separator → exact int64 full-Rupiah amount, no rounding |
| `PassiveIncomeService` | Plain CRUD + `Sum` |
| `TargetsService` | `enrich()` computes `CurrentValue`/`Percent`/`LowerIsBetter` per `metric_type`: `equity` = net equity of latest snapshot; `gold_grams` = Σ gram×qty of `logam_mulia` holdings; `passive_income` = Σ passive income sources; `debt_ratio` = i_owe/net_equity as a percent (the only `lower_is_better`); `custom` = stored manual value as-is |
| `DashboardService` | Assembles the full `GET /dashboard` payload (equity, debt, passive, expense, allocation) from the latest snapshot + independently-timelined debt/expense-period data; zeroed gracefully if the user has no snapshots yet |
| `ProgressService` | Monthly/quarterly/yearly net-equity and debt time series; each debt-progress point pairs that debt snapshot's totals with the most recent asset snapshot on or before its date |

---

### 4.6 HTTP API Layer

**`backend/internal/httpapi/router.go`** — global middleware: chi `Logger`, `Recoverer`, `BodyLimit` (1 MiB request body cap, applied even to unauthenticated routes), CORS restricted to `cfg.CORSOrigin`.

Three route groups:
1. **Public** — `GET /healthz` (no `/api/v1` prefix), `GET /auth/google/login`, `GET /auth/google/callback`, `POST /auth/logout`, `POST /auth/login`.
2. **`RatesSyncOrAuthMiddleware`** — `GET /rates/latest`, `POST /rates` only. Accepts *either* the normal session cookie *or* a matching `X-Rates-Sync-Token` header (falls through to plain session auth if the token is unset/doesn't match) — scoped to just these two routes so a leaked token can only touch gold rates, nothing else.
3. **`AuthMiddleware`** (session cookie required) — everything else. Full route list, all under `/api/v1`:

```
GET    /auth/me
GET    /categories
GET    /rates

GET    /snapshots                              GET  /snapshots/latest
POST   /snapshots                              GET  /snapshots/{date}
GET    /snapshots/{date}/holdings              POST /snapshots/{date}/holdings
DELETE /snapshots/{id}
PUT    /holdings/{id}                          DELETE /holdings/{id}

GET    /debt-snapshots                         GET  /debt-snapshots/latest
POST   /debt-snapshots                         GET  /debt-snapshots/{date}
POST   /debt-snapshots/{date}/entries          DELETE /debt-snapshots/{id}
PUT    /debt-entries/{id}                      DELETE /debt-entries/{id}

GET    /expense-periods                        GET  /expense-periods/latest
POST   /expense-periods                        GET  /expense-periods/{id}
DELETE /expense-periods/{id}
POST   /expense-periods/{periodId}/envelopes
POST   /expense-periods/{periodId}/fixed-expenses
PUT    /budget-envelopes/{id}                  DELETE /budget-envelopes/{id}
PUT    /fixed-expenses/{id}                    DELETE /fixed-expenses/{id}

GET    /expense-source-mappings
PUT    /expense-source-mappings/{source}
POST   /expense-ingestions

GET    /notification-apps                      POST /notification-apps
PUT    /notification-apps/{source}
GET    /notification-apps/{source}/patterns     POST /notification-apps/{source}/patterns
PUT    /notification-patterns/{id}

GET    /passive-income                         POST /passive-income
PUT    /passive-income/{id}                    DELETE /passive-income/{id}

GET    /targets                                POST /targets
PUT    /targets/{id}                           DELETE /targets/{id}

GET    /dashboard
GET    /progress
GET    /debt-progress
```

The `notification-apps`/`notification-patterns` `POST`/`PUT` routes are admin-only in intent but reuse this same session auth rather than a separate admin surface — see §9's notification-catalog section for why (there's no separate admin UI; these are called directly, e.g. via curl, by whoever manages the catalog).

**`middleware.go`**:
- `AuthMiddleware` — reads the `wf_session` cookie, validates via `AuthService.Authenticate` (transparently reissuing the cookie if the session is close to its sliding-window expiry), injects the `domain.User` into request context. 401 JSON on missing/invalid/expired.
- `RatesSyncOrAuthMiddleware` — see above; inert (always falls through) when `RATES_SYNC_TOKEN` is unset.

**`errors.go`** — service/db errors mapped to HTTP status codes:

| Error | HTTP Status |
|---|---|
| `db.ErrNotFound` | 404 |
| `service.ErrSnapshotLocked` | 409 |
| `service.ErrSnapshotDateExists` | 409 |
| `service.ErrPeriodMonthExists` | 409 |
| `service.ErrSnapshotDateInPast` | 400 |
| `service.ErrInvalidCategory` | 400 |
| `service.ErrInvalidInput` | 400 |
| `service.ErrNoRateEntry` | 422 |
| `service.ErrNoActivePeriod` | 422 |
| `service.ErrNoSourceMapping` | 422 |
| `service.ErrEnvelopeNotFound` | 422 |
| anything else | 500 |

---

### 4.7 Migrations

Located in `backend/migrations/`, embedded via `//go:embed *.sql`, run automatically at API startup via Goose.

| # | File | Summary |
|---|---|---|
| 00001 | `init.sql` | Initial schema: users, categories, rate_entries, snapshots, holdings, debts (flat table, later replaced), passive_income_sources, targets |
| 00002 | `seed.sql` | Seeds the pre-auth default user (fixed UUID) + 9 static categories |
| 00003 | `debt_snapshots.sql` | Replaces the flat `debts` table with a snapshot/entry model (`debt_snapshots`/`debt_entries`) mirroring `snapshots`/`holdings` |
| 00004 | `snapshot_soft_delete.sql` | Adds `deleted_at` soft-delete to `snapshots` and `debt_snapshots`; replaces the unique-date constraint with a partial index scoped to non-deleted rows |
| 00005 | `auth.sql` | Google Sign-In: adds `google_sub`/`email_verified`/`avatar_url` to `users`, creates `sessions` |
| 00006 | `password_auth.sql` | Adds nullable `password_hash` to `users` for email/password sign-in |
| 00007 | `monthly_expenses.sql` | `expense_periods` (25th–24th pay-cycle windows, never lock), `budget_envelopes`, `fixed_expenses` |
| 00008 | `expense_categories.sql` | Adds (later-removed) `expense_categories` grouping layer above envelopes |
| 00009 | `notification_expense_ingestion.sql` | `expense_source_mappings` + `notification_expense_events` (idempotent ingestion audit trail) |
| 00010 | `remove_expense_categories.sql` | Drops `expense_categories` — turned out not to earn its keep above envelopes |
| 00011 | `notification_catalog.sql` | `notification_apps` + `notification_app_packages` — replaces a hardcoded Android enum with a backend-driven catalog |
| 00012 | `seed_notification_apps.sql` | Seeds gopay/dana/bca with package names; no patterns yet |
| 00013 | `gopay_bca_patterns.sql` | Corrects BCA's catalog entry against real captured data ("BCA Mobile", `com.bca`, Western-decimal format) |
| 00014 | `exact_rupiah_unit.sql` | **Removes the "thousands of IDR" convention** — rescales every affected `*_idr` column ×1000 so all money is full/raw Rupiah going forward |
| 00015 | `seed_jago.sql` | Seeds Bank Jago as a notification source with real captured regex patterns |
| 00016 | `seed_gopay_using_pattern.sql` | Adds a third GoPay template (payment via a linked funding source) |
| 00017 | `fixed_expense_source.sql` | Adds nullable, write-once `source` column to `fixed_expenses` |
| 00018 | `fixed_expense_notes.sql` | Adds nullable, manual-entry-only `notes` column to `fixed_expenses` |

---

## 5. Database Schema

```
users
  id uuid PK
  email text
  display_name text
  avatar_url text
  google_sub text
  email_verified bool
  password_hash text          -- nullable; only set for the one account using password login
  created_at timestamptz

sessions
  token text PK                -- the opaque session-cookie value itself, not a JWT
  user_id uuid FK → users
  created_at timestamptz
  expires_at timestamptz       -- sliding 7-day window, refreshed on use; 30-day absolute cap from created_at

categories
  id smallserial PK
  key text UNIQUE               -- e.g. "logam_mulia", "saham"
  label text
  color_oklch text
  kind text                     -- 'asset' | 'liability'
  price_linked bool             -- true if value is derived from a rate entry
  sort_order smallint

rate_entries
  id uuid PK
  user_id uuid FK → users
  entry_date date
  antam numeric      -- gold price Antam, full IDR/gram
  kinghalim numeric   -- gold price King Halim, full IDR/gram
  ubs numeric          -- gold price UBS, full IDR/gram
  usd_idr numeric     -- full IDR per 1 USD
  UNIQUE(user_id, entry_date)

snapshots
  id uuid PK
  user_id uuid FK → users
  snapshot_date date
  deleted_at timestamptz        -- soft delete
  UNIQUE(user_id, snapshot_date) WHERE deleted_at IS NULL

holdings
  id uuid PK
  snapshot_id uuid FK → snapshots (CASCADE DELETE)
  category_id smallint FK → categories
  name text
  detail text
  value_idr bigint        -- full/raw IDR
  is_liability bool
  gram numeric             -- for gold
  qty numeric               -- number of gold bars/coins
  brand text                -- 'Antam' | 'King Halim' | 'UBS'
  usd_value numeric        -- for USD-denominated assets (raw USD, not IDR)
  currency text              -- 'IDR' | 'USD'
  created_at / updated_at timestamptz

debt_snapshots
  id uuid PK
  user_id uuid FK → users
  snapshot_date date
  deleted_at timestamptz
  UNIQUE(user_id, snapshot_date) WHERE deleted_at IS NULL

debt_entries
  id uuid PK
  debt_snapshot_id uuid FK → debt_snapshots (CASCADE DELETE)
  name text
  type text
  value_idr bigint        -- full/raw IDR
  direction text            -- 'i_owe' | 'owed_to_me'
  created_at / updated_at timestamptz

passive_income_sources
  id uuid PK
  user_id uuid FK → users
  category_id smallint FK → categories
  name text
  per_year_idr bigint     -- full/raw IDR per year

expense_periods
  id uuid PK
  user_id uuid FK → users
  start_date date          -- the 25th
  end_date date              -- the 24th of the following month
  created_at timestamptz
  UNIQUE(user_id, start_date)

budget_envelopes
  id uuid PK
  period_id uuid FK → expense_periods (CASCADE DELETE)
  name text
  committed_amount_idr bigint   -- full/raw IDR
  created_at / updated_at timestamptz

fixed_expenses
  id uuid PK
  period_id uuid FK → expense_periods (CASCADE DELETE)
  envelope_id uuid FK → budget_envelopes (CASCADE DELETE, nullable)  -- NULL = standalone, not bundled
  name text
  amount_idr bigint             -- full/raw IDR
  source text                     -- nullable; set once by notification ingestion, never touched by manual edits
  notes text                       -- nullable; manual-entry only
  created_at / updated_at timestamptz

expense_source_mappings
  id uuid PK
  user_id uuid FK → users
  source text                      -- e.g. "gopay"
  envelope_name text              -- by NAME, not id — envelopes are period-scoped
  updated_at timestamptz
  UNIQUE(user_id, source)

notification_expense_events
  id uuid PK
  user_id uuid FK → users
  idempotency_key text
  source text
  raw_title / raw_text / raw_big_text text   -- nullable, kept even when ignored
  occurred_at timestamptz
  parse_status text                -- 'created' | 'ignored'
  amount_idr bigint                -- nullable, full/raw IDR
  merchant_name text               -- nullable
  envelope_id uuid FK → budget_envelopes
  fixed_expense_id uuid FK → fixed_expenses
  created_at timestamptz
  UNIQUE(user_id, idempotency_key)

notification_apps
  id uuid PK
  source text UNIQUE                -- e.g. "gopay", "dana", "bca", "jago"
  display_name text
  amount_thousand_sep / amount_decimal_sep char(1)   -- backend-only; how this app formats amounts
  enabled bool
  updated_at timestamptz

notification_app_packages
  id uuid PK
  app_id uuid FK → notification_apps (CASCADE DELETE)
  package_name text UNIQUE          -- Android package name, e.g. "com.gojek.gopay"

notification_patterns
  id uuid PK
  app_id uuid FK → notification_apps (CASCADE DELETE)
  priority int                       -- ascending; first match wins
  field text                          -- 'title' | 'text' | 'big_text'
  regex text                          -- must contain a named "amount" capture group
  description text
  enabled bool
  updated_at timestamptz

targets
  id uuid PK
  user_id uuid FK → users
  name text
  year int
  metric_type text        -- 'equity'|'gold_grams'|'passive_income'|'debt_ratio'|'custom'
  target_value numeric
  unit text
  manual_current_value numeric   -- used only for metric_type='custom'
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

Base URL: `/api/v1`. Every endpoint except `/healthz`, `/auth/google/*`, `/auth/logout`, and `/auth/login` requires a valid `wf_session` cookie (or, for `/rates`/`/rates/latest` only, a matching `X-Rates-Sync-Token` header) — the user id is resolved from that, never passed explicitly. See §4.6 for the full route table.

### Rate Entries

```json
POST /rates
{ "entry_date": "2026-07-01", "antam": 1900000, "kinghalim": 1890000, "ubs": 1880000, "usd_idr": 16350 }
```

### Snapshots / Holdings

```json
POST /snapshots
{ "snapshot_date": "2026-07-01", "copy_from_latest": true, "initial_holdings": [] }

POST/PUT holding
{ "category_id": 1, "name": "Antam 10g", "gram": 10, "qty": 3, "brand": "Antam" }
```
Value is derived server-side from the latest rate entry for price-linked categories; `value_idr` is only used directly for everything else / as a manual fallback.

### Debt Snapshots / Entries

Mirrors the Snapshots/Holdings shape one-for-one, on its own independent date timeline:
```json
POST /debt-snapshots
{ "snapshot_date": "2026-07-01", "copy_from_latest": true }

POST /debt-snapshots/{date}/entries
{ "name": "KPR BCA", "type": "mortgage", "value_idr": 750000000, "direction": "i_owe" }
```

### Expense Periods / Budget Envelopes / Fixed Expenses

```json
POST /expense-periods
{ "year": 2026, "month": 8, "copy_envelopes": true }

POST /expense-periods/{periodId}/envelopes
{ "name": "Kebutuhan Rumah", "committed_amount_idr": 8000000 }

POST /expense-periods/{periodId}/fixed-expenses
{ "envelope_id": "...", "name": "Listrik", "amount_idr": 350000, "notes": "optional" }
```
`fixed_expenses.source` is never client-settable — it's only ever populated by the notification-ingestion pipeline (§7), and stays untouched by later manual edits to the same row.

### Expense Source Mappings / Ingestion (used by the Android app)

```json
PUT /expense-source-mappings/{source}
{ "envelope_name": "Kebutuhan Rumah" }

POST /expense-ingestions
{
  "idempotency_key": "…",
  "source": "gopay",
  "raw_title": "Cool, transfer successful",
  "raw_text": "You've successfully transferred Rp1.234 to Mba Jen.",
  "raw_big_text": null,
  "occurred_at": "2026-07-26T09:15:00Z"
}
```
Always responds `200` with `{"status": "created" | "ignored", ...}` — never a 4xx for "didn't match a pattern," since that's an expected, idempotent-on-replay outcome, not a client error. A `422` here means something the user needs to fix (no active period / no envelope mapping / mapped envelope missing), and *is* meant to be retried once they do.

### Notification Catalog (Android sync target + admin)

```json
GET /notification-apps
[{ "source": "gopay", "display_name": "GoPay", "package_names": ["com.gojek.gopay"], "enabled": true }, ...]

POST /notification-apps/{source}/patterns
{ "field": "text", "regex": "transferred Rp(?P<amount>[\\d.]+) to (?P<merchant>.+)\\.", "description": "outgoing transfer" }
```
The regex must contain a named `amount` group; a named `merchant` group is optional, but if present in the regex it must also capture non-empty text for that pattern to count as a match. `priority` auto-assigns to `max(existing)+10` when omitted.

### Passive Income / Targets

```json
POST /passive-income
{ "category_id": 1, "name": "Rental unit", "per_year_idr": 24000000 }

POST /targets
{ "name": "Equity 2026", "year": 2026, "metric_type": "equity", "target_value": 5000000000, "unit": "IDR" }
```

### Dashboard / Progress

`GET /dashboard` returns nested `equity`/`debt`/`passive`/`expense`/`allocation` sections aggregating the latest snapshot and current expense period. `GET /progress?granularity=monthly|quarterly|yearly` and `GET /debt-progress?granularity=…` return net-equity/debt time series, each point carrying `delta_idr`/`delta_pct` versus the previous point.

---

## 7. Business Rules

### Snapshot Immutability (assets and debts alike)
Only the snapshot with the latest `snapshot_date` is editable, on each of the two independent timelines (asset snapshots and debt snapshots don't have to be dated in lockstep). All writes to holdings/debt entries check this first; mutating a historical snapshot returns `409 Conflict`. Deleting a snapshot is allowed even if it isn't the latest — deleting the current latest promotes the next-most-recent to editable.

### Expense Periods Never Lock
Unlike snapshots, an expense period (and everything inside it — envelopes, fixed expenses) stays editable indefinitely; it's an ongoing log, not a point-in-time record. A period runs from the 25th of one month through the 24th of the next, named after the month it ends in.

### The Notification-Driven Expense Pipeline
`ExpenseIngestionService.Ingest` (called by `POST /expense-ingestions`, the Android app's sync target):
1. **Parse** the raw captured text via `NotificationCatalogService.Parse`, which matches it against that source's cached regex patterns in priority order. No match → `"ignored"` — a terminal, non-error outcome; the client should not retry it, since the text will never match retroactively. This also writes a `notification_expense_events` audit row even on failure, since raw captures double as the sample data new patterns get written against.
2. **Resolve the period** covering the notification's `occurred_at`, using the same 25th-of-month boundary rule as everything else. Missing period, missing source→envelope mapping, or a mapped envelope that no longer exists are all *retryable* errors (`422`, not `"ignored"`) — they resolve themselves once the user fixes their setup, so the client's outbox keeps retrying rather than discarding the capture.
3. **Idempotently create** the `fixed_expenses` row + its `notification_expense_events` audit row in one transaction, keyed on `(user_id, idempotency_key)` — a client retry (or a concurrent double-send) of the same key returns the original result rather than creating a duplicate expense.

### Value Derivation
Price-linked categories never store a value the user typed directly:
- **Gold** (`logam_mulia`): `gram × qty × goldPricePerGram(brand, latestRateEntry)`.
- **USD Bonds/ETF/Cash** (`bonds_usd`, `us_etf`, `uang_tunai` with `currency=USD`): `usd_value × usd_idr`.
- If no rate entry exists yet and a derived value is required, the API returns `422`.

### Multi-User Design
Every account signs in with Google or email/password and gets its own isolated workspace; every table is scoped by `user_id` and every read/write path enforces it. The seeded pre-auth user is claimed in place by whichever account logs in first, so local data predating auth isn't lost.

### Backfilling Snapshots
Creating a snapshot dated *before* the current latest immediately locks it (`is_editable: false`). `initial_holdings[]` on the create request is the only path for writing to a non-latest snapshot.

---

## 8. Frontend (React + TypeScript)

`package.json`: React 18, TypeScript, Vite, `react-router-dom`, `@tanstack/react-query`. No CSS framework — hand-written CSS under `src/styles/` + one stylesheet per page/component. Linted with `oxlint`, not ESLint.

`main.tsx` provider tree: `QueryClientProvider` (`staleTime: 30_000`, `refetchOnWindowFocus: true`, `retry: 1`, `networkMode: 'always'`) → `AuthProvider` → `ToastProvider` → `MoneyVisibilityProvider` → `App`.

`App.tsx`: `useAuth()` branches into a loading spinner, a standalone `<Login/>` (no router, no chrome) if signed out, or the full `<BrowserRouter>` tree if signed in — every route below is nested under `<AppShell/>`.

### 8.1 Pages

| Path | Component | Description |
|---|---|---|
| `/` | `Dashboard` | Equity/debt/passive summary cards, allocation donut, plus expense donuts (spent-by-envelope, committed-by-envelope) |
| `/assets` | `Assets` (+ `AssetModal`, `NewSnapshotModal`) | Snapshot picker, category-filtered holdings table, add/edit/copy-forward |
| `/debts` | `Debts` (+ `DebtModal`, `DebtSnapshotModal`) | Mirrors Assets: "I Owe"/"Owed to Me" columns, debt-to-equity ratio banner |
| `/expenses` | `MonthlyExpenses` (+ `FixedExpenseModal`, `NewPeriodModal`, `BudgetEnvelopeModal`) | Period picker, envelope cards with actual-vs-committed + progress bar, fixed-expense rows (with a source badge for notification-created ones), new period/envelope/expense flows |
| `/passive-income` | `PassiveIncome` (+ `PassiveIncomeModal`) | Source list with per-source bars, total/year vs target, monthly equivalents — *hidden from nav, but fully functional* |
| `/targets` | `Targets` (+ `TargetModal`) | Goal cards with progress bars — *hidden from nav, but fully functional* |
| `/progress` | `Progress` | Line charts (monthly/quarterly/yearly): net-equity trend and debt trend (with a secondary debt-ratio series) |
| `/rates` | `Rates` | Latest gold/USD summary cards, new-entry form, full price history table |
| *(standalone)* | `Login` (`pages/auth/`) | Google OAuth button + email/password form; not wrapped in `AppShell` |

### 8.2 Hooks

One React Query hook per API resource in `src/hooks/`, each mutation invalidating its own query key plus any derived ones (`dashboard`, `targets`, `progress`, `debtProgress` as relevant): `useCategories`, `useRates`, `useSnapshots`, `useHoldings`, `useDebtSnapshots`, `useDebtEntries` (bundled in `useDebtSnapshots.ts`), `useDebtProgress`, `useExpensePeriods`, `useBudgetEnvelopes`, `useFixedExpenses`, `usePassiveIncome`, `useTargets`, `useDashboard`, `useProgress`. `usePullToRefresh` is the odd one out — a reusable touch-gesture hook, not API-backed (see below).

### 8.3 Components

- **`layout/AppShell.tsx`** — sidebar (nav, snapshot net-worth footer, user card, sign-out), header (title, date, hide/show-values toggle, and — only inside the Android WebView, detected via `window.WealthfolioNative` — sync/settings icon links), a mobile user bar, the routed `<Outlet/>`, and a bottom nav for mobile. Implements pull-to-refresh: wraps `<main>` (keyed on route, so it fully remounts on navigation — which is why `usePullToRefresh` returns a *callback* ref, not a plain `RefObject`, to correctly re-attach after each remount) and calls `queryClient.refetchQueries({ type: 'active' })` on a completed pull gesture.
- **`layout/nav.ts`** — `NAV_ITEMS` (the 6 linked routes) and `PAGE_TITLES`.
- **`Modal.tsx`** — the generic modal shell every add/edit form uses.
- **`charts/DonutChart.tsx`** / **`charts/LineChart.tsx`** — hand-rolled SVG charts with hover interaction; `LineChart` supports an optional secondary dashed series on its own right-hand axis (used for debt-ratio alongside debt totals).

### 8.4 Contexts

- **`AuthContext`** — `{ user, isLoading, logout }`; `logout()` resets the auth query (flips back to `<Login/>` immediately) then clears the entire React Query cache, so a different account never sees a stale previous session's data.
- **`MoneyVisibilityContext`** — `{ hidden, toggle, fmt, fmtExact }`; `hidden` persists to `localStorage`. `fmt` is the abbreviated formatter (intended for Dashboard/sidebar only), `fmtExact` the exact one (intended everywhere else) — see §10 for a note on where this split isn't perfectly followed in practice.
- **`ToastContext`** — `{ showError, showSuccess }` plus a standalone `errorMessage(err)` helper used throughout the modals.

### 8.5 Utility Libraries

- **`lib/api.ts`** — the single typed `fetch` client (`credentials: 'include'`, JSON, throws `ApiError` on non-2xx, 401 triggers an auth-query invalidation). The `api` object namespaces every call by resource.
- **`lib/format.ts`** — `fmtIdr`/`money` (abbreviated), `fmtIdrExact`/`moneyExact` (exact), `goldFmt`/`goldFmtShort` (per-gram price), `usdFmt`, `formatShortDate`/`formatTimestamp`, `sourceLabel` (notification source id → display name, e.g. `gopay`→`GoPay`, with a capitalize-fallback for unknown sources), `parseNumeric`.
- **`lib/holdingCalc.ts`** — client-side mirror of the backend's `ComputeHoldingValue`, for `AssetModal`'s live value preview as the user types, plus the category-driven field-visibility predicates that decide which inputs `AssetModal` shows.
- **`lib/colors.ts`** — a handful of literal OKLCH color constants mirroring `styles/tokens.css`, needed because raw SVG presentation attributes don't reliably resolve `var()` cross-browser.

---

## 9. Android App (Kotlin + Jetpack Compose)

Full detail lives in [`android/README.md`](android/README.md) — this section is a summary for cross-reference from the backend/frontend docs above.

**What it is**: package `com.wealthfolio.mobile`, `minSdk 26`. Two jobs: (1) host the web frontend in a WebView as the entire visible UI, and (2) run a `NotificationListenerService` that watches GoPay/DANA/BCA/Bank Jago payment notifications and auto-forwards them to `POST /expense-ingestions`.

**Architecture**: `MainShell.kt` composes the WebView (`web/WebTabScreen.kt`) exactly once, never tearing it down; two native Compose screens (Settings, Sync Status) render as full-screen overlays on top of it, toggled by a JS bridge call from the web page (`window.WealthfolioNative.openNative(...)`) rather than Compose Navigation — which would otherwise dispose and reload the WebView on every screen open/close. System back dismisses the open overlay via a `BackHandler` rather than falling through to closing the app.

**Notification pipeline** (see §7 for the backend side): `TransactionNotificationListener` filters by package name against an in-memory cache *before* even launching a coroutine → re-checks the per-source DataStore toggle → captures raw title/text/bigText untouched (parsing is backend-only) → computes a deterministic idempotency key once at capture time → writes to a local Room "outbox" table (durable across app close) → an expedited WorkManager job plus a ~15-minute belt-and-suspenders periodic sweep drain it against `/expense-ingestions`, applying retry semantics based on the response (created/ignored are terminal; 422 and network/5xx errors retry; 401 stops the sweep and routes back to login).

**Notification catalog**: which apps/packages the listener watches for is fetched from `GET /notification-apps` (synced on app-open and Settings-open) and mirrored into local Room — not a hardcoded enum — so the backend can add/rename/retire a supported source without an APK release.

**Native screens**: Settings (permission status, per-source enable toggles, envelope mapping, bundled app icons with a color-initial fallback) and Sync Status (per-status outbox listing, tap-through detail, retry/clear controls, daily auto-clear, pull-to-refresh that triggers an immediate sync attempt).

**Build config**: debug points at a hardcoded LAN IP (`API_BASE_URL`/`WEB_ORIGIN` in `app/build.gradle.kts`, must be edited to match the developer's machine — a physical phone can't resolve `localhost` as itself); release points at `https://etherna.id` and requires a local, gitignored `keystore.properties` + `.jks` to produce a signed build.

---

## 10. Money Unit Convention

> **Every `*_idr` / `value_idr` / `amount_idr` / `target_value` integer represents full/raw IDR — whole Rupiah, no scaling factor.**
> Example: `900000000` means Rp 900,000,000.

This was **not always true**: migration [`00014_exact_rupiah_unit.sql`](backend/migrations/00014_exact_rupiah_unit.sql) rescaled every affected column (×1000) and removed an earlier "thousands of IDR" convention that `00001_init.sql` originally documented. If you find a comment anywhere still claiming "thousands," it predates 00014 and is wrong — trust the code (`domain.RateEntry`'s doc comment, `notificationparse/parse.go`, `service/valuation.go`, `frontend/src/lib/format.ts`'s header comment) over any stale prose.

The only non-IDR numeric fields, never subject to this convention at all, are physical quantities (`gram`, `qty`) and raw USD amounts (`Holding.UsdValue`), plus `Target.TargetValue`/`CurrentValue` for `metric_type` values that aren't currency (`gold_grams` is grams, `debt_ratio` is a percent, `custom` is whatever unit the user picked).

**One known frontend inconsistency worth knowing about, not silently "fixed" by this doc**: `format.ts`'s header comment states the abbreviated formatter (`fmt`) is for "Dashboard + the persistent sidebar total" only, with `fmtExact` used "everywhere else" — but `MonthlyExpenses.tsx` and `Debts.tsx` both currently use the abbreviated formatter throughout their tables, while `PassiveIncome.tsx`/`Targets.tsx`/`Progress.tsx` use the exact one as documented. If you're touching money display on the Expenses or Debts pages, be aware the current behavior doesn't match the stated intent.

---

## 11. Deployment

### Docker Compose

Three services in `docker-compose.yml`:

| Service | Image | Role |
|---|---|---|
| `db` | `postgres:16-alpine` | Database, data in the `db_data` volume |
| `api` | Built from `./backend/Dockerfile` | Go API server on port 8080 (internal only) |
| `web` | Built from `./frontend/Dockerfile` | Caddy serving the static frontend build + reverse-proxying `/api/v1` to `api:8080` |

`web` exposes ports 80/443; Caddy handles TLS automatically when `DOMAIN` is a real hostname.

### Environment Variables

Copy `.env.example` to `.env` and set at minimum `POSTGRES_PASSWORD`, `DOMAIN`, `CORS_ORIGIN`, `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET`/`GOOGLE_REDIRECT_URL`, `APP_BASE_URL`; optionally `RATES_SYNC_TOKEN`/`RATES_SYNC_EMAIL` if running the `cmd/rates-sync` cron job against this deployment.

### Operational extras (deploy-host-specific, not tracked in this repo)

A production deployment typically layers on, outside version control:
- A daily cron'd `pg_dump -Fc` of the database to a backup directory, with age-based cleanup of old dumps.
- A daily cron'd run of `cmd/rates-sync` (see §4.1 and [`backend/README.md`](backend/README.md#cmdrates-sync)) so gold/USD rates stay current without manual entry.

### Local Development

See the root [README.md](README.md#local-development) for the full walkthrough; short version:

```bash
# backend
cd backend && DATABASE_URL="postgres://wealthfolio:wealthfolio@localhost:5432/wealthfolio?sslmode=disable" go run ./cmd/api

# frontend
cd frontend && npm install && npm run dev   # http://localhost:5173, proxies /api/v1 to :8080

# android (optional) — see android/README.md
```

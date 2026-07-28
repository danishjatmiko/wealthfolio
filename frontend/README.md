# Etherna frontend

React + TypeScript web frontend for Etherna. See the [root README](../README.md) and [DOCUMENTATION.md](../DOCUMENTATION.md) for what the whole project is and how the backend it talks to works. This file is just the frontend module's local dev quickstart.

## Stack

React 18 + TypeScript + Vite · [`react-router`](https://reactrouter.com) · [`@tanstack/react-query`](https://tanstack.com/query) for all server state · hand-rolled SVG charts (no charting library) · plain CSS (no framework) under `src/styles/` + one stylesheet per page/component. Linted with [`oxlint`](https://oxc.rs), not ESLint.

## Local development

```bash
npm install
npm run dev
```

Serves at `http://localhost:5173`. `vite.config.ts` proxies `/api` to `http://localhost:8080` (the backend's default port) and binds the dev server to all interfaces (`host: true`, not just `localhost`) — deliberately, so the [Android app's](../android/) embedded WebView, running on a phone on the same LAN, can reach it via your machine's LAN IP.

## Scripts

| Command | Does |
|---|---|
| `npm run dev` | Start the Vite dev server with HMR |
| `npm run build` | Typecheck (`tsc -b`) then production build (`vite build`) |
| `npm run preview` | Serve the production build locally |
| `npm run lint` | Run `oxlint` |

## Structure

```
src/
  main.tsx        React root: QueryClientProvider, AuthProvider, ToastProvider, MoneyVisibilityProvider
  App.tsx         Auth gate (Login vs the routed app) + route table
  types.ts        TypeScript types mirroring the backend's JSON contract
  lib/            api.ts (typed fetch client), format.ts (money/date formatters), holdingCalc.ts, colors.ts
  context/        AuthContext, MoneyVisibilityContext (hide/show values), ToastContext
  hooks/          one React Query hook per API resource, plus usePullToRefresh (a gesture hook, not API-backed)
  components/     layout/AppShell (sidebar/header/bottom-nav shell), charts/, Modal
  pages/          one folder per screen — Dashboard, Assets, Debts, Expenses, PassiveIncome, Targets, Progress, Rates, auth/Login
```

## Money unit convention

**Every monetary field from the API is an integer in full/raw IDR (whole Rupiah) — no scaling factor.** `fmtIdr`/`money` (in `lib/format.ts`) are the *abbreviated* formatters (Dashboard + sidebar total only, e.g. `"Rp3.75 B"`); everywhere else should use `fmtIdrExact`/`moneyExact`. Both are hide-aware — pass through `useMoney()`'s `fmt`/`fmtExact` so the global "hide values" toggle (persisted to `localStorage`) is respected automatically.

## Native app integration

`window.WealthfolioNative` (declared in `src/wealthfolio-native.d.ts`) is injected only when this page is loaded inside the Android app's WebView — `AppShell.tsx` checks for it to conditionally show native-only Settings/Sync-status header icons and to defer sign-out to the native side there. `public/app-vh.js` sets a `--app-vh` CSS variable from `window.innerHeight` because some embedded WebViews don't size `100vh`/`100dvh` reliably. Neither has any effect when the site is opened in a normal browser.

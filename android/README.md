# Etherna Android app

The Android client for Etherna (package `com.wealthfolio.mobile`; the module is also referred to as "Wealthfolio" in a few places — same app). It does two distinct things:

1. **Hosts the project's [web frontend](../frontend/) in a WebView** as the app's entire visible UI — there's no separate native UI for the actual portfolio/expense screens.
2. **Runs a `NotificationListenerService`** that watches for GoPay/DANA/BCA/Bank Jago payment notifications and auto-forwards them to the backend, which turns them into expense records — this is the app's actual reason to exist beyond "the website in a wrapper."

See the [root README](../README.md) and [DOCUMENTATION.md](../DOCUMENTATION.md) for the backend/frontend this app talks to.

## Stack

Kotlin · Jetpack Compose (Material3) · Hilt (DI) · Room (local persistence) · WorkManager (background sync) · Retrofit + OkHttp · DataStore + EncryptedSharedPreferences.

`minSdk 26` (Android 8.0+ — required by `NotificationListenerService`), `targetSdk`/`compileSdk 36`.

## Architecture: WebView + native overlays, not native screens

`MainShell.kt` composes `WebTabScreen` (the WebView) exactly once and never tears it down. Two native Compose screens — Settings and Sync Status — render as full-screen overlays *on top of* the WebView, toggled by a JS bridge call from the web page itself (`window.WealthfolioNative.openNative("sync" | "settings")`, see `web/WebTabScreen.kt`). This is deliberate: Compose Navigation would dispose the WebView composable on route change, reloading the page and losing scroll/state every time you opened Settings. System back (button or gesture) is wired via a `BackHandler` in `MainShell.kt` that dismisses whichever overlay is open, rather than falling through to closing the app.

Package layout under `app/src/main/kotlin/com/wealthfolio/mobile/`:

| Package | Responsibility |
|---|---|
| `auth/` | Session token storage (`TokenStore`, EncryptedSharedPreferences), Google OAuth via Chrome Custom Tab + deep-link callback, email/password login, an OkHttp interceptor that attaches the session cookie and clears it on 401 |
| `data/outbox/` | Room-backed durable queue of captured notifications awaiting sync (see pipeline below) |
| `data/notificationcatalog/` | Local Room mirror of the backend's supported-apps catalog (see below) |
| `notifications/` | The `NotificationListenerService` itself + idempotency-key builder |
| `sync/` | WorkManager scheduling, the two sync/cleanup workers, and the Sync Status screen |
| `settings/` | Settings screen: notification-access permission, per-source toggles, envelope mapping |
| `network/` | Retrofit `ApiService` + DTOs |
| `di/` | Hilt modules (network client, both Room DBs, WorkManager) |
| `ui/` | App shell composables, shared top bar, theme |
| `web/` | The WebView host + its JS bridge |

## The notification-capture pipeline

This is the app's most distinctive feature, so it's worth understanding end to end:

1. **`TransactionNotificationListener.onNotificationPosted`** fires for *every* notification on the device (that's how the OS's notification-access grant works — permission is all-or-nothing at the system level). The very first line is a synchronous, non-suspend lookup — `NotificationCatalogCache.sourceForPackage(sbn.packageName) ?: return` — that discards anything from an irrelevant app before a coroutine is even launched or any content is read.
2. If the package matches a known source, it re-checks `SourcePreferences.isEnabled(source)` fresh (a DataStore-backed per-source toggle, default off) — so toggling a source off in Settings takes effect on the very next notification, no service restart needed.
3. If enabled, it captures the raw `title`/`text`/`bigText` untouched — **no parsing happens on-device**, that's the backend's job (formats drift, and a DB row edit is cheaper to ship than an APK release). It computes a deterministic idempotency key (SHA-256 of source + package + title + text + minute-rounded timestamp) once, at capture time — every retry reuses the same key so the backend can dedupe.
4. The captured notification is written to a local Room table immediately (`OutboxRepository.enqueue`) — this is what makes capture durable across app close/process death.
5. An expedited one-shot `OutboxSyncWorker` fires right away (`SyncScheduler.syncNow()`), backed by a belt-and-suspenders ~15-minute periodic sweep registered on every app open, so a captured notification keeps retrying even if the app is closed and never reopened.
6. `OutboxSyncWorker` POSTs each pending/failed row to `/expense-ingestions` and applies these transitions: **200 "created"** → done (`SENT`); **200 "ignored"** → done, not a recognized transaction (`IGNORED`); **422** → retryable, resolves once the user fixes their period/envelope-mapping setup (`FAILED`, retried next sweep); **network error / 5xx** → stays `PENDING`, retried next sweep; **401** → stop the whole sweep, the session's gone and the UI routes back to login on its own.

You can watch this whole pipeline live from the **Sync Status** screen (reachable from the web app's own header icon) — every captured notification and its current status, a tap-through detail view with timestamps, per-row retry for failed items, manual/auto history clearing, and pull-to-refresh (which just means "attempt to sync right now," since the list itself is already a live Room query).

## The notification catalog

`NotificationSource` used to be a hardcoded Kotlin enum. It's now a table fetched from the backend (`GET /notification-apps`, synced on app-open and on Settings-open) and mirrored into its own local Room database — so the backend can add, rename, or retire a supported app at any time and every install picks it up on its next sync, with no Play Store release. The Settings screen renders one toggle card per catalog entry, with a bundled square icon for the sources that have one (GoPay/BCA/Jago/DANA) and a hash-based color-initial fallback for anything else.

The catalog cache is deliberately split from the Room table it mirrors — `NotificationCatalogCache` is a plain in-memory map, since the notification listener's fast-path filter (step 1 above) can't afford a suspend/Room query.

## Building & running

### Debug (Android Studio)

1. Open `android/` as the project root in Android Studio.
2. **Update the LAN IP** — `app/build.gradle.kts`'s `debug` build type hardcodes `API_BASE_URL`/`WEB_ORIGIN` to a specific `192.168.1.155` address (a physical phone can't resolve `localhost` as itself the way an emulator sometimes can). Change this to your own machine's current LAN IP.
3. Have the backend (`cd ../backend && go run ./cmd/api`) and frontend dev server (`cd ../frontend && npm run dev`) both running and reachable at that IP.
4. Run (▶) with a device/emulator selected. WebView remote debugging is enabled automatically in debug builds — inspect the embedded page via `chrome://inspect` over USB.

### Debug (CLI)

```bash
cd android
./gradlew assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

### Release

Requires a local `keystore.properties` (copy `keystore.properties.example` at the repo root and fill in real values) plus the `.jks` file it points at — both gitignored, never committed. Without them present, the release build config silently skips signing rather than failing at configure time, so double-check the output is actually signed before distributing it.

```bash
cd android
./gradlew assembleRelease
adb install -r app/build/outputs/apk/release/app-release.apk
```

Release points at `https://etherna.id` (no LAN IP to edit) and keeps cleartext traffic disabled — only the debug build's manifest override relaxes that, for the plain-HTTP local dev backend. Minification is currently off (`isMinifyEnabled = false`); `proguard-rules.pro` exists but is a placeholder for whenever that changes.

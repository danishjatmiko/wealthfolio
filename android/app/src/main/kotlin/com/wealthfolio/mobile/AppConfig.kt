package com.wealthfolio.mobile

/**
 * Where the backend lives — values come from BuildConfig, set per build
 * type in app/build.gradle.kts: debug points at the local dev backend
 * (your Mac's LAN IP; a physical phone can't resolve `localhost` as
 * itself, and emulators reach a real LAN IP fine too), release points at
 * https://etherna.id. Change both constants together there; WEB_ORIGIN
 * backs the Web tab's WebView and must NOT have the `/api/v1/` suffix
 * API_BASE_URL needs.
 */
object AppConfig {
    const val API_BASE_URL = BuildConfig.API_BASE_URL
    const val WEB_ORIGIN = BuildConfig.WEB_ORIGIN
}

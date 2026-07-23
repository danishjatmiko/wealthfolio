package com.wealthfolio.mobile

/**
 * Where the backend lives. `10.0.2.2` is the Android emulator's alias for
 * the host machine's `localhost` — it works out of the box against
 * `go run ./cmd/api` on your Mac when running in the emulator. On a
 * physical phone this must instead be your Mac's LAN IP (e.g.
 * `http://192.168.1.23:8080/`), and neither works once the backend is
 * somewhere the phone can reach over the internet — see the plan's
 * "Prerequisite this plan doesn't solve" note (Part C). Change both
 * constants together; WEB_ORIGIN backs the Web tab's WebView and must NOT
 * have the `/api/v1/` suffix API_BASE_URL needs.
 */
object AppConfig {
    const val API_BASE_URL = "http://192.168.1.155:8080/api/v1/"
    const val WEB_ORIGIN = "http://192.168.1.155:5173"
}

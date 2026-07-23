package com.wealthfolio.mobile.notifications

/**
 * The three notification sources this app knows about. [id] is what's
 * sent to the backend as `source` (must match the Go constants in
 * notificationparse/parse.go) and used as the Settings/DataStore key.
 *
 * ⚠️ [packageNames] are best-guesses, NOT yet verified against the real
 * installed apps — confirming and correcting these against a real device
 * is the first thing to do once this project is running (see the plan's
 * "Known gap going in" callout and Build order step 4). A wrong package
 * name here just means that source silently never matches any
 * notification; nothing breaks, it just does nothing.
 */
enum class NotificationSource(val id: String, val packageNames: Set<String>) {
    GOPAY("gopay", setOf("com.gojek.gopay", "com.gojek.app")),
    DANA("dana", setOf("id.dana")),
    BCA("bca", setOf("com.bca", "com.bca.mybca")),
    ;

    companion object {
        fun forPackageName(packageName: String): NotificationSource? =
            entries.firstOrNull { packageName in it.packageNames }
    }
}

package com.wealthfolio.mobile.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.wealthfolio.mobile.auth.TokenStore
import com.wealthfolio.mobile.settings.SettingsScreen
import com.wealthfolio.mobile.sync.SyncStatusScreen
import com.wealthfolio.mobile.web.WebTabScreen

private enum class NativeOverlay { NONE, SYNC, SETTINGS }

/**
 * The Web tab (the whole existing site) is the app's permanent home,
 * composed exactly once and never torn down. Sync Status / Settings —
 * reached only via the site's own header buttons calling
 * window.WealthfolioNative.openNative(...) (see WebTabScreen's JS
 * bridge) — render as a full-screen overlay *on top* of it instead of a
 * separate nav destination. That distinction matters: Compose Navigation
 * disposes a composable when you navigate away from its route, which
 * would tear down and reload the WebView (losing scroll position, page,
 * and any client-side state like the hide-values toggle) every single
 * time you opened Settings and came back.
 */
@Composable
fun MainShell(tokenStore: TokenStore) {
    var overlay by remember { mutableStateOf(NativeOverlay.NONE) }

    Box(modifier = Modifier.fillMaxSize()) {
        WebTabScreen(
            tokenStore = tokenStore,
            onNavigateNative = { route ->
                overlay = when (route) {
                    "sync" -> NativeOverlay.SYNC
                    "settings" -> NativeOverlay.SETTINGS
                    else -> NativeOverlay.NONE
                }
            },
        )

        when (overlay) {
            NativeOverlay.SYNC -> SyncStatusScreen(onBack = { overlay = NativeOverlay.NONE })
            NativeOverlay.SETTINGS -> SettingsScreen(onBack = { overlay = NativeOverlay.NONE })
            NativeOverlay.NONE -> Unit
        }
    }
}

package com.wealthfolio.mobile.web

import android.annotation.SuppressLint
import android.os.Handler
import android.os.Looper
import android.webkit.CookieManager
import android.webkit.JavascriptInterface
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import com.wealthfolio.mobile.AppConfig
import com.wealthfolio.mobile.auth.TokenStore

/**
 * Exposed to the page as `window.WealthfolioNative` (see AppShell.tsx) —
 * its mere presence is also what the site uses to decide whether to show
 * the Sync/Settings links at all, so a plain desktop/mobile browser
 * (where this object is never injected) never sees them.
 *
 * @JavascriptInterface methods run on a background thread, not the UI
 * thread, per the Android WebView docs — posting to the main looper
 * before touching onNavigateNative (which drives Compose state) is
 * required, not just cautious.
 */
private class WealthfolioJsBridge(private val onNavigateNative: (String) -> Unit) {
    private val mainHandler = Handler(Looper.getMainLooper())

    @JavascriptInterface
    fun openNative(route: String) {
        mainHandler.post { onNavigateNative(route) }
    }
}

/**
 * The entire existing web app (Dashboard, Assets, Debt & Loans, Expenses,
 * etc.), unmodified, running in a WebView — this is the app's home
 * screen. Seeded with the same session token TokenStore already holds
 * for API calls, injected as a real cookie before the first load — from
 * then on it's a normal browser cookie jar (see the plan's Part C for
 * why this needed no backend changes at all).
 *
 * Composed exactly once for the app's lifetime (see MainShell — Settings/
 * Sync are overlays on top of this, not separate nav destinations that
 * would tear this WebView down and reload it every time you left and
 * came back).
 */
@SuppressLint("SetJavaScriptEnabled")
@Composable
fun WebTabScreen(tokenStore: TokenStore, onNavigateNative: (String) -> Unit) {
    AndroidView(
        modifier = Modifier.fillMaxSize(),
        factory = { context ->
            val token = tokenStore.token
            if (token != null) {
                val cookieManager = CookieManager.getInstance()
                cookieManager.setAcceptCookie(true)
                cookieManager.setCookie(AppConfig.WEB_ORIGIN, "wf_session=$token")
                cookieManager.flush()
            }

            WebView(context).apply {
                settings.javaScriptEnabled = true
                settings.domStorageEnabled = true
                // Without these two, WebView ignores the page's
                // <meta name="viewport"> tag and lays it out as a ~980px
                // desktop page before scaling it down, breaking the
                // site's responsive CSS.
                settings.useWideViewPort = true
                settings.loadWithOverviewMode = true
                addJavascriptInterface(WealthfolioJsBridge(onNavigateNative), "WealthfolioNative")
                webViewClient = object : WebViewClient() {
                    override fun shouldOverrideUrlLoading(
                        view: WebView,
                        request: android.webkit.WebResourceRequest,
                    ): Boolean {
                        // Keep navigation inside this WebView — the site
                        // is single-origin, so anything else (e.g. the
                        // Google OAuth link) opens here too rather than
                        // escaping to an external browser.
                        return false
                    }

                    override fun onPageFinished(view: WebView, url: String) {
                        super.onPageFinished(view, url)
                        // The site's own logout (now native-only, but kept
                        // as a defensive fallback — e.g. the session
                        // simply expiring) clears the wf_session cookie in
                        // *this WebView's* cookie jar — a completely
                        // separate store from TokenStore, which is what
                        // native API calls and WealthfolioRoot's logged-
                        // in/out switch actually key off.
                        val cookies = CookieManager.getInstance().getCookie(AppConfig.WEB_ORIGIN)
                        val hasSession = cookies?.contains("wf_session=") == true
                        if (!hasSession && tokenStore.token != null) {
                            tokenStore.clear()
                        }
                    }
                }
                loadUrl(AppConfig.WEB_ORIGIN)
            }
        },
    )
}

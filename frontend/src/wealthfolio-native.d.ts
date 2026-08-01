export {}

declare global {
  interface Window {
    /** Injected only by the Android app's Web tab (see WebTabScreen.kt's
     * addJavascriptInterface) — never present in a normal browser. Its
     * presence is what AppShell.tsx uses to decide whether to show the
     * Sync/Settings header links at all. */
    WealthfolioNative?: {
      openNative: (route: 'sync' | 'settings') => void
      /** Reloads the native WebView itself — a plain in-page data refetch
       *  can't pick up a new JS/CSS bundle after a deploy, since the
       *  WebView is only ever loaded once per app launch (see
       *  WebTabScreen.kt). Optional: absent on an app build older than
       *  this bridge method, so callers must check before calling it. */
      reload?: () => void
    }
  }
}

export {}

declare global {
  interface Window {
    /** Injected only by the Android app's Web tab (see WebTabScreen.kt's
     * addJavascriptInterface) — never present in a normal browser. Its
     * presence is what AppShell.tsx uses to decide whether to show the
     * Sync/Settings header links at all. */
    WealthfolioNative?: {
      openNative: (route: 'sync' | 'settings') => void
    }
  }
}

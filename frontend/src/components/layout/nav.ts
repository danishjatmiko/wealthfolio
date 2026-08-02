export interface NavItem {
  to: string
  label: string
  icon: string
  /** Only ever shown in the desktop sidebar — never in the mobile bottom
   * nav, and never at all inside the Android app's WebView regardless of
   * its viewport width (see AppShell.tsx, which filters on both
   * conditions). For a feature that isn't meant to be reachable from a
   * phone at all, not just "not linked from the compact nav." */
  desktopOnly?: boolean
  /** Kept in the mobile bottom nav's primary row (max 4, plus the built-in
   * "More" tab makes 5 — Apple HIG's cap for a tab bar). Every other
   * non-desktopOnly item moves into the "More" sheet instead. Ignored by
   * the desktop sidebar, which always lists everything flat since it
   * isn't cramped. */
  bottomPrimary?: boolean
}

// Targets is hidden from the nav for now (still fully functional at its
// route, just not linked from here) — add an entry back to un-hide.
//
// Labels are kept short because the mobile bottom bar's "More" sheet still
// only has half its normal width per item; PAGE_TITLES below carries the
// full names for page headers. "Passive Income" is the practical ceiling —
// it fills a More-sheet cell at 375px without wrapping, so anything longer
// needs checking there before it goes in.
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Home', icon: '◫', bottomPrimary: true },
  { to: '/assets', label: 'Assets', icon: '▤', bottomPrimary: true },
  { to: '/debts', label: 'Debt', icon: '⇄', bottomPrimary: true },
  { to: '/expenses', label: 'Expenses', icon: '▦', bottomPrimary: true },
  // Not desktopOnly — logging a big purchase is exactly the kind of thing
  // that happens on a phone, in the moment, not at a desk.
  { to: '/big-expenses', label: 'Big Exp', icon: '▣' },
  { to: '/passive-income', label: 'Passive Income', icon: '⊛' },
  { to: '/progress', label: 'Progress', icon: '∿' },
  { to: '/rates', label: 'Rates', icon: '¤' },
  { to: '/bonds', label: 'Bonds', icon: '◈', desktopOnly: true },
  { to: '/simulation', label: 'Simulation', icon: '↗' },
]

export const PAGE_TITLES: Record<string, string> = {
  '/': 'Portfolio Overview',
  '/assets': 'Assets',
  '/debts': 'Debt & Loans',
  '/expenses': 'Monthly Expenses',
  '/big-expenses': 'Big Expense',
  '/passive-income': 'Passive Income',
  '/targets': 'Targets',
  '/progress': 'Progress',
  '/rates': 'Rates & Prices',
  '/simulation': 'Growth Simulation',
  '/bonds': 'Bond Ledger',
}

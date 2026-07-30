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
}

// Targets is hidden from the nav for now (still fully functional at its
// route, just not linked from here) — add an entry back to un-hide.
//
// Labels are kept short because the mobile bottom bar fits seven items at
// ~53px each; PAGE_TITLES below carries the full names for page headers.
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Home', icon: '◫' },
  { to: '/assets', label: 'Assets', icon: '▤' },
  { to: '/debts', label: 'Debt', icon: '⇄' },
  { to: '/expenses', label: 'Expenses', icon: '▦' },
  { to: '/passive-income', label: 'Passive', icon: '⊛' },
  { to: '/progress', label: 'Progress', icon: '∿' },
  { to: '/rates', label: 'Rates', icon: '¤' },
  { to: '/bonds', label: 'Bonds', icon: '◈', desktopOnly: true },
  { to: '/simulation', label: 'Simulation', icon: '↗', desktopOnly: true },
]

export const PAGE_TITLES: Record<string, string> = {
  '/': 'Portfolio Overview',
  '/assets': 'Assets',
  '/debts': 'Debt & Loans',
  '/expenses': 'Monthly Expenses',
  '/passive-income': 'Passive Income',
  '/targets': 'Targets',
  '/progress': 'Progress',
  '/rates': 'Rates & Prices',
  '/simulation': 'Growth Simulation',
  '/bonds': 'Bond Ledger',
}

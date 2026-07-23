export interface NavItem {
  to: string
  label: string
  icon: string
}

// Passive Income and Targets are hidden from the nav for now (still fully
// functional at their routes, just not linked from here) — remove these
// two comments and add their entries back to un-hide.
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: '◫' },
  { to: '/assets', label: 'Assets', icon: '▤' },
  { to: '/debts', label: 'Debt & Loans', icon: '⇄' },
  { to: '/expenses', label: 'Expenses', icon: '▦' },
  { to: '/progress', label: 'Progress', icon: '∿' },
  { to: '/rates', label: 'Rates', icon: '¤' },
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
}

export interface NavItem {
  to: string
  label: string
  icon: string
}

// 8 destinations, in this exact order (per design spec).
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: '◫' },
  { to: '/assets', label: 'Assets', icon: '▤' },
  { to: '/debts', label: 'Debt & Loans', icon: '⇄' },
  { to: '/expenses', label: 'Expenses', icon: '▦' },
  { to: '/passive-income', label: 'Passive Income', icon: '↻' },
  { to: '/targets', label: 'Targets', icon: '◎' },
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

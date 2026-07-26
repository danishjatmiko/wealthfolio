import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { money, moneyExact } from '../lib/format'

const STORAGE_KEY = 'wealthfolio:hideValues'

interface MoneyVisibilityValue {
  hidden: boolean
  toggle: () => void
  /** Abbreviated formatter (Dashboard/sidebar total only), masked when hidden. */
  fmt: (value: number) => string
  /** Exact formatter (every other page), masked when hidden. */
  fmtExact: (value: number) => string
}

const MoneyVisibilityContext = createContext<MoneyVisibilityValue | null>(null)

function readInitial(): boolean {
  if (typeof window === 'undefined') return false
  try {
    return window.localStorage.getItem(STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

export function MoneyVisibilityProvider({ children }: { children: ReactNode }) {
  const [hidden, setHidden] = useState<boolean>(readInitial)

  useEffect(() => {
    try {
      window.localStorage.setItem(STORAGE_KEY, hidden ? '1' : '0')
    } catch {
      // localStorage unavailable (private mode, etc.) — ignore
    }
  }, [hidden])

  const toggle = useCallback(() => setHidden((h) => !h), [])
  const fmt = useCallback((value: number) => money(value, hidden), [hidden])
  const fmtExact = useCallback((value: number) => moneyExact(value, hidden), [hidden])

  const value = useMemo(() => ({ hidden, toggle, fmt, fmtExact }), [hidden, toggle, fmt, fmtExact])

  return <MoneyVisibilityContext.Provider value={value}>{children}</MoneyVisibilityContext.Provider>
}

/** Hook exposing the global "hide values" state and a hide-aware money formatter. */
export function useMoney(): MoneyVisibilityValue {
  const ctx = useContext(MoneyVisibilityContext)
  if (!ctx) throw new Error('useMoney must be used within MoneyVisibilityProvider')
  return ctx
}

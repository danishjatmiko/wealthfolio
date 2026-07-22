import { createContext, useCallback, useContext, useMemo } from 'react'
import type { ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, authQueryKey } from '../lib/api'
import type { User } from '../types'

interface AuthValue {
  user: User | undefined
  isLoading: boolean
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const qc = useQueryClient()

  const { data: user, isLoading } = useQuery({
    queryKey: authQueryKey,
    queryFn: api.auth.me,
    retry: false,
    staleTime: 5 * 60_000,
  })

  const logout = useCallback(async () => {
    try {
      await api.auth.logout()
    } finally {
      // resetQueries (unlike clear()) optimistically flips the still-
      // mounted `me` observer's data back to undefined and refetches
      // immediately, so App() re-renders into <Login/> right away instead
      // of needing a manual page refresh to notice the session is gone.
      await qc.resetQueries({ queryKey: authQueryKey })
      // Every other query is scoped to whichever account was signed in;
      // dropping the whole cache (not just the auth query) keeps the next
      // account from ever seeing a stale query still holding the previous
      // one's data.
      qc.clear()
    }
  }, [qc])

  const value = useMemo(() => ({ user, isLoading, logout }), [user, isLoading, logout])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

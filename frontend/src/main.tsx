import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './styles/global.css'
import './styles/components.css'
import App from './App.tsx'
import { setQueryClient } from './lib/api'
import { AuthProvider } from './context/AuthContext'
import { MoneyVisibilityProvider } from './context/MoneyVisibilityContext'
import { ToastProvider } from './context/ToastContext'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      // Refetch whenever the tab/window regains focus — e.g. after
      // switching away and back — not just on route navigation. Still
      // respects staleTime above, so rapid alt-tabbing within 30s won't
      // refire every query.
      refetchOnWindowFocus: true,
      retry: 1,
      // This app only ever talks to a same-machine backend, so there's no
      // meaningful "offline" state to design around — always attempt
      // fetches rather than letting the browser's online/offline signal
      // (which can get stuck stale in some environments) pause them.
      networkMode: 'always',
    },
    mutations: {
      networkMode: 'always',
    },
  },
})
setQueryClient(queryClient)

const root = createRoot(document.getElementById('root')!)

// The kids' game at /interactive-game is public: it skips AuthProvider
// (so no /auth/me call and no login screen) and is code-split so the
// finance app's bundle doesn't carry it.
if (window.location.pathname.startsWith('/interactive-game')) {
  import('./pages/game/GameApp').then(({ GameApp }) =>
    root.render(
      <StrictMode>
        <GameApp />
      </StrictMode>,
    ),
  )
} else {
  renderApp()
}

function renderApp() {
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <ToastProvider>
            <MoneyVisibilityProvider>
              <App />
            </MoneyVisibilityProvider>
          </ToastProvider>
        </AuthProvider>
      </QueryClientProvider>
    </StrictMode>,
  )
}

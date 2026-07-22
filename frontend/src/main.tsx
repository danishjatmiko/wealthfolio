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
      refetchOnWindowFocus: false,
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

createRoot(document.getElementById('root')!).render(
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

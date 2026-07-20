import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './styles/global.css'
import './styles/components.css'
import App from './App.tsx'
import { MoneyVisibilityProvider } from './context/MoneyVisibilityContext'
import { ToastProvider } from './context/ToastContext'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <MoneyVisibilityProvider>
          <App />
        </MoneyVisibilityProvider>
      </ToastProvider>
    </QueryClientProvider>
  </StrictMode>,
)

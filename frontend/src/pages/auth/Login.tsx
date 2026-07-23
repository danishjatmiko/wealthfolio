import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { api, authQueryKey } from '../../lib/api'
import { errorMessage, useToast } from '../../context/ToastContext'
import './Login.css'

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.9c1.7-1.57 2.7-3.87 2.7-6.62z"
      />
      <path
        fill="#34A853"
        d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.9-2.26c-.8.54-1.83.86-3.06.86-2.36 0-4.35-1.6-5.07-3.74H.9v2.33A9 9 0 0 0 9 18z"
      />
      <path
        fill="#FBBC05"
        d="M3.93 10.68A5.4 5.4 0 0 1 3.65 9c0-.58.1-1.15.28-1.68V4.99H.9A9 9 0 0 0 0 9c0 1.45.35 2.83.9 4.01l3.03-2.33z"
      />
      <path
        fill="#EA4335"
        d="M9 3.58c1.32 0 2.51.46 3.44 1.35l2.58-2.58C13.46.89 11.43 0 9 0A9 9 0 0 0 .9 4.99l3.03 2.33C4.65 5.18 6.64 3.58 9 3.58z"
      />
    </svg>
  )
}

export function Login() {
  const qc = useQueryClient()
  const { showError } = useToast()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handlePasswordLogin(e: React.FormEvent) {
    e.preventDefault()
    if (!email || !password) return
    setSubmitting(true)
    try {
      await api.auth.login(email, password)
      // Same mechanism as AuthContext's logout: resetQueries flips the
      // still-mounted `me` observer to refetch immediately, picking up the
      // session cookie the login response just set, so App() swaps from
      // <Login/> into the real app right away.
      await qc.resetQueries({ queryKey: authQueryKey })
    } catch (err) {
      showError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="login-screen">
      <div className="login-card card">
        <div className="login-brand">
          <div className="brand-mark">₩</div>
          <div className="brand-name">Etherna</div>
        </div>
        <p className="login-subtitle">
          Sign in to see your own net worth. Each account has its own private data — no one else
          can see it.
        </p>
        <a className="btn login-google-btn" href={api.auth.googleLoginUrl}>
          <GoogleIcon />
          Sign in with Google
        </a>

        <div className="login-divider">
          <span>or</span>
        </div>

        <form className="login-password-form" onSubmit={handlePasswordLogin}>
          <input
            className="field-input"
            type="email"
            autoComplete="username"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="field-input"
            type="password"
            autoComplete="current-password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            Sign in
          </button>
        </form>
      </div>
    </div>
  )
}

import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AppShell } from './components/layout/AppShell'
import { useAuth } from './context/AuthContext'
import { Dashboard } from './pages/Dashboard'
import { Login } from './pages/auth/Login'
import { Assets } from './pages/assets/Assets'
import { Debts } from './pages/debts/Debts'
import { PassiveIncome } from './pages/passive/PassiveIncome'
import { Targets } from './pages/targets/Targets'
import { Progress } from './pages/progress/Progress'
import { Rates } from './pages/rates/Rates'

function App() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="auth-loading-screen">
        <div className="brand-mark">₩</div>
      </div>
    )
  }

  if (!user) {
    return <Login />
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/assets" element={<Assets />} />
          <Route path="/debts" element={<Debts />} />
          <Route path="/passive-income" element={<PassiveIncome />} />
          <Route path="/targets" element={<Targets />} />
          <Route path="/progress" element={<Progress />} />
          <Route path="/rates" element={<Rates />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App

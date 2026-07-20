import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useMemo } from 'react'
import { NAV_ITEMS, PAGE_TITLES } from './nav'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useDashboard } from '../../hooks/useDashboard'
import './AppShell.css'

function useTodayLabel() {
  return useMemo(
    () =>
      new Intl.DateTimeFormat('en-GB', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).format(new Date()),
    [],
  )
}

export function AppShell() {
  const location = useLocation()
  const { hidden, toggle, fmt } = useMoney()
  const { data: dashboard } = useDashboard()
  const todayLabel = useTodayLabel()
  const title = PAGE_TITLES[location.pathname] ?? 'Wealthfolio'

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-mark">₩</div>
          <div className="brand-name">Wealthfolio</div>
        </div>
        <nav className="sidebar-nav">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) => 'navi' + (isActive ? ' navi-active' : '')}
            >
              <span className="navi-icon">{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className="sidebar-footer-label">Snapshot net worth</div>
          <div className="sidebar-footer-value mono">
            {dashboard ? fmt(dashboard.equity.total_idr) : '—'}
          </div>
        </div>
      </aside>

      <div className="content-col">
        <header className="content-header">
          <div>
            <div className="content-title">{title}</div>
            <div className="content-subtitle">{todayLabel}</div>
          </div>
          <button type="button" className="hide-toggle" onClick={toggle}>
            <span className="hide-toggle-icon">◉</span>
            {hidden ? 'Show' : 'Hide'} values
          </button>
        </header>

        <main className="content-area pg" key={location.pathname}>
          <Outlet />
        </main>

        <nav className="bottom-nav">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) => 'bottom-navi' + (isActive ? ' bottom-navi-active' : '')}
            >
              <span className="bottom-navi-icon">{item.icon}</span>
              <span className="bottom-navi-label">{item.label}</span>
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  )
}

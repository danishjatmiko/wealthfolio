import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useMemo } from 'react'
import { NAV_ITEMS, PAGE_TITLES } from './nav'
import { useAuth } from '../../context/AuthContext'
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
  const { user, logout } = useAuth()
  const todayLabel = useTodayLabel()
  const title = PAGE_TITLES[location.pathname] ?? 'Etherna'

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <img className="brand-mark" src="/brand-mark.svg" alt="" />
          <div className="brand-name">Etherna</div>
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

        {user && (
          <div className="sidebar-user">
            {user.avatar_url ? (
              <img className="sidebar-user-avatar" src={user.avatar_url} alt="" />
            ) : (
              <div className="sidebar-user-avatar sidebar-user-avatar-fallback">
                {user.display_name.charAt(0).toUpperCase()}
              </div>
            )}
            <div className="sidebar-user-info">
              <div className="sidebar-user-name">{user.display_name}</div>
              <div className="sidebar-user-email">{user.email}</div>
            </div>
            <button
              type="button"
              className="sidebar-user-signout"
              onClick={() => void logout()}
              title="Sign out"
              aria-label="Sign out"
            >
              ⏻
            </button>
          </div>
        )}
      </aside>

      <div className="content-col">
        <header className="content-header">
          <div>
            <div className="content-title">{title}</div>
            <div className="content-subtitle">{todayLabel}</div>
          </div>
          <div className="content-header-actions">
            <button type="button" className="hide-toggle" onClick={toggle}>
              <span className="hide-toggle-icon">◉</span>
              {hidden ? 'Show' : 'Hide'} values
            </button>
            {/* window.WealthfolioNative only exists inside the Android
                app's Web tab (injected by WebTabScreen.kt), so these are
                invisible on desktop/mobile web — sign-in there has no
                native counterpart to redirect to, and signing out is
                deliberately native-app-only now (see SettingsScreen.kt
                on the Android side). */}
            {window.WealthfolioNative && (
              <>
                <button
                  type="button"
                  className="header-native-link"
                  onClick={() => window.WealthfolioNative?.openNative('sync')}
                  title="Sync status"
                >
                  ⟳
                </button>
                <button
                  type="button"
                  className="header-native-link"
                  onClick={() => window.WealthfolioNative?.openNative('settings')}
                  title="Settings"
                >
                  ⚙
                </button>
              </>
            )}
          </div>
        </header>

        {user && (
          <div className="mobile-user-bar">
            {user.avatar_url ? (
              <img className="sidebar-user-avatar" src={user.avatar_url} alt="" />
            ) : (
              <div className="sidebar-user-avatar sidebar-user-avatar-fallback">
                {user.display_name.charAt(0).toUpperCase()}
              </div>
            )}
            <div className="sidebar-user-email">{user.email}</div>
            {/* Same native-vs-web split as the header sync/settings buttons
                above: inside the Android WebView, sign-out is native-app-only
                (SettingsScreen.kt), so this button would have nothing to do. */}
            {!window.WealthfolioNative && (
              <button
                type="button"
                className="sidebar-user-signout"
                onClick={() => void logout()}
                title="Sign out"
                aria-label="Sign out"
              >
                ⏻
              </button>
            )}
          </div>
        )}

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

import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { NAV_ITEMS, PAGE_TITLES } from './nav'
import { useAuth } from '../../context/AuthContext'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useDashboard } from '../../hooks/useDashboard'
import { usePullToRefresh } from '../../hooks/usePullToRefresh'
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
  const queryClient = useQueryClient()
  const { ref: pullRef, pullDistance, refreshing } = usePullToRefresh<HTMLElement>(() => {
    // Inside the Android app, a swipe-down should behave like a real
    // reload — the WebView is only ever loaded once per app launch (see
    // WebTabScreen.kt), so a plain data refetch can leave it running a
    // stale JS bundle indefinitely even past a fresh deploy. Falls back to
    // a data-only refetch on an app build predating this bridge method,
    // and always on plain mobile/desktop web, where there's no WebView to
    // reload in the first place.
    if (window.WealthfolioNative?.reload) {
      window.WealthfolioNative.reload()
      return
    }
    return queryClient.refetchQueries({ type: 'active' })
  })

  // desktopOnly nav items (e.g. Simulation) never appear in the bottom nav
  // at all — that alone is enough on the web, since the sidebar that DOES
  // list them is already display:none below the mobile breakpoint. Inside
  // the Android app's WebView, though, a wide-enough viewport (a tablet)
  // could still render the sidebar, so those items are additionally
  // stripped from the sidebar itself whenever window.WealthfolioNative is
  // present — "desktop only" and "never inside the app" are two separate
  // conditions, not the same one.
  const isNativeApp = !!window.WealthfolioNative
  const sidebarNavItems = NAV_ITEMS.filter((item) => !(item.desktopOnly && isNativeApp))
  const bottomNavItems = NAV_ITEMS.filter((item) => !item.desktopOnly)
  const primaryBottomItems = bottomNavItems.filter((item) => item.bottomPrimary)
  const moreBottomItems = bottomNavItems.filter((item) => !item.bottomPrimary)

  // The "More" sheet is its own affordance, not a route — closing it when
  // the route changes (including navigating to one of its own items) keeps
  // it from staying open over the next page.
  const [moreOpen, setMoreOpen] = useState(false)
  useEffect(() => {
    setMoreOpen(false)
  }, [location.pathname])
  const isMoreActive = moreBottomItems.some((item) => location.pathname === item.to)

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <img className="brand-mark" src="/brand-mark.svg" alt="" />
          <div className="brand-name">Etherna</div>
        </div>
        <nav className="sidebar-nav">
          {sidebarNavItems.map((item) => (
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

        <main className="content-area pg" key={location.pathname} ref={pullRef}>
          <div className="pull-refresh-indicator" style={{ height: refreshing ? 36 : pullDistance }}>
            <span
              className={'pull-refresh-spinner' + (refreshing ? ' pull-refresh-spinning' : '')}
              style={{
                opacity: Math.min(1, (refreshing ? 36 : pullDistance) / 50),
                transform: refreshing ? undefined : `rotate(${pullDistance * 3}deg)`,
              }}
            >
              ⟳
            </span>
          </div>
          <Outlet />
        </main>

        {moreOpen && (
          <div className="bottom-nav-sheet">
            {moreBottomItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) => 'bottom-nav-sheet-item' + (isActive ? ' active' : '')}
              >
                <span className="bottom-navi-icon">{item.icon}</span>
                {item.label}
              </NavLink>
            ))}
          </div>
        )}

        <nav className="bottom-nav">
          {primaryBottomItems.map((item) => (
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
          {moreBottomItems.length > 0 && (
            <button
              type="button"
              className={'bottom-navi bottom-navi-more' + (isMoreActive ? ' bottom-navi-active' : '')}
              onClick={() => setMoreOpen((open) => !open)}
              aria-expanded={moreOpen}
            >
              <span className="bottom-navi-icon">⋯</span>
              <span className="bottom-navi-label">More</span>
            </button>
          )}
        </nav>
      </div>
    </div>
  )
}

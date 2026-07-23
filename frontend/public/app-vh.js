// Some embedded WebViews (notably the Android app's Web tab) don't
// reliably size 100vh/100dvh to the actual visible area — the app
// shell ends up with an unpredictable height, which cascades into
// the bottom nav rendering mid-page instead of pinned to the
// bottom. window.innerHeight is reliable everywhere, embedded or
// not, so --app-vh (used by .app-shell in AppShell.css) is driven
// by that instead of trusting vh/dvh units to behave.
function setAppVh() {
  document.documentElement.style.setProperty('--app-vh', window.innerHeight + 'px')
}
setAppVh()
window.addEventListener('resize', setAppVh)

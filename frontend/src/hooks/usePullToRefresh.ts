import { useCallback, useRef, useState } from 'react'

const PULL_THRESHOLD = 70
const MAX_PULL = 100
const RESISTANCE = 0.5

/**
 * Drag-down-to-refresh gesture for a scrollable container. Only arms when
 * the gesture starts at scrollTop 0 (so it never fights normal scrolling),
 * tracks the pull with resistance while dragging, and calls onRefresh once
 * released past PULL_THRESHOLD. Below that, it just snaps back.
 *
 * Returns a callback ref rather than a plain RefObject — AppShell's
 * `<main>` is keyed on the route (`key={location.pathname}`), so it fully
 * unmounts/remounts on every navigation. A callback ref is what correctly
 * notices that and re-attaches listeners to the new node; a RefObject +
 * mount-time-only effect would silently go stale after the first route
 * change.
 */
export function usePullToRefresh<T extends HTMLElement>(onRefresh: () => Promise<unknown> | void) {
  const [pullDistance, setPullDistance] = useState(0)
  const [refreshing, setRefreshing] = useState(false)

  // Ref trick so listeners always call the latest onRefresh without
  // needing to be torn down and rebuilt whenever it changes identity.
  const onRefreshRef = useRef(onRefresh)
  onRefreshRef.current = onRefresh

  const cleanupRef = useRef<(() => void) | null>(null)

  const setRef = useCallback((el: T | null) => {
    cleanupRef.current?.()
    cleanupRef.current = null
    if (!el) return

    let startY: number | null = null
    let dragging = false
    let currentPull = 0

    function onTouchStart(e: TouchEvent) {
      if (el!.scrollTop > 0) return
      startY = e.touches[0].clientY
      dragging = true
    }

    function onTouchMove(e: TouchEvent) {
      if (!dragging || startY === null) return
      if (el!.scrollTop > 0) {
        dragging = false
        currentPull = 0
        setPullDistance(0)
        return
      }
      const dy = e.touches[0].clientY - startY
      if (dy <= 0) {
        currentPull = 0
        setPullDistance(0)
        return
      }
      e.preventDefault()
      currentPull = Math.min(MAX_PULL, dy * RESISTANCE)
      setPullDistance(currentPull)
    }

    async function onTouchEnd() {
      if (!dragging) return
      dragging = false
      startY = null
      const shouldRefresh = currentPull >= PULL_THRESHOLD
      currentPull = 0
      setPullDistance(0)
      if (shouldRefresh) {
        setRefreshing(true)
        try {
          await onRefreshRef.current()
        } finally {
          setRefreshing(false)
        }
      }
    }

    el.addEventListener('touchstart', onTouchStart, { passive: true })
    el.addEventListener('touchmove', onTouchMove, { passive: false })
    el.addEventListener('touchend', onTouchEnd)
    el.addEventListener('touchcancel', onTouchEnd)

    cleanupRef.current = () => {
      el.removeEventListener('touchstart', onTouchStart)
      el.removeEventListener('touchmove', onTouchMove)
      el.removeEventListener('touchend', onTouchEnd)
      el.removeEventListener('touchcancel', onTouchEnd)
    }
  }, [])

  return { ref: setRef, pullDistance, refreshing }
}

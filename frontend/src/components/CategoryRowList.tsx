import { useMemo, useState } from 'react'
import type { DashboardCategoryRow } from '../types'

type SortKey = 'name' | 'value' | 'percent'
type SortDir = 'asc' | 'desc'

interface CategoryRowListProps {
  rows: DashboardCategoryRow[]
  fmt: (value: number) => string
  emptyMessage: string
}

// Shared by every "category -> value -> %" list on the Dashboard (allocation,
// spent/committed by envelope). Click a header to sort by it; click again to
// flip direction. Numeric columns default to descending (biggest first) on
// first click, the name column to ascending (A-Z).
export function CategoryRowList({ rows, fmt, emptyMessage }: CategoryRowListProps) {
  const [sortKey, setSortKey] = useState<SortKey | null>(null)
  const [sortDir, setSortDir] = useState<SortDir>('desc')

  const sorted = useMemo(() => {
    if (!sortKey) return rows
    const dir = sortDir === 'asc' ? 1 : -1
    return [...rows].sort((a, b) => {
      if (sortKey === 'name') return a.label.localeCompare(b.label) * dir
      if (sortKey === 'value') return (a.value_idr - b.value_idr) * dir
      return (a.percent - b.percent) * dir
    })
  }, [rows, sortKey, sortDir])

  function toggle(key: SortKey) {
    if (sortKey !== key) {
      setSortKey(key)
      setSortDir(key === 'name' ? 'asc' : 'desc')
    } else {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
    }
  }

  function arrow(key: SortKey) {
    return sortKey === key ? <span className="dash-cat-sort-arrow">{sortDir === 'asc' ? '↑' : '↓'}</span> : null
  }

  if (rows.length === 0) return <div className="empty-state">{emptyMessage}</div>

  return (
    <>
      <div className="dash-cat-head">
        <span className="dash-cat-head-spacer" />
        <span className="dash-cat-head-name" onClick={() => toggle('name')}>
          Category{arrow('name')}
        </span>
        <span className="dash-cat-head-val" onClick={() => toggle('value')}>
          Value{arrow('value')}
        </span>
        <span className="dash-cat-head-pct" onClick={() => toggle('percent')}>
          %{arrow('percent')}
        </span>
      </div>
      {sorted.map((c) => (
        <div className="dash-cat-row" key={c.category_key}>
          <span className="dash-cat-swatch" style={{ background: c.color_oklch }} />
          <span className="dash-cat-name">{c.label}</span>
          <span className="mono dash-cat-val">{fmt(c.value_idr)}</span>
          <span className="mono dash-cat-pct">{c.percent.toFixed(2)}%</span>
        </div>
      ))}
    </>
  )
}

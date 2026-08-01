import { useMemo, useState } from 'react'
import { Modal } from '../../components/Modal'
import { DonutChart, type DonutDatum } from '../../components/charts/DonutChart'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useBigExpenses, useBigExpenseSummary } from '../../hooks/useBigExpenses'
import { formatShortDate } from '../../lib/format'
import { BigExpenseModal } from './BigExpenseModal'
import type { BigExpense } from '../../types'
import './BigExpenses.css'

type SortKey = 'expense_date' | 'name' | 'category' | 'amount_idr'

function entryWord(count: number): string {
  return count === 1 ? 'entry' : 'entries'
}

const COLUMNS: { key: SortKey; label: string; numeric?: boolean }[] = [
  { key: 'expense_date', label: 'Date' },
  { key: 'name', label: 'Name' },
  { key: 'category', label: 'Category' },
  { key: 'amount_idr', label: 'Amount', numeric: true },
]

function yearOptions(expenses: BigExpense[]): number[] {
  const current = new Date().getFullYear()
  const years = new Set(expenses.map((e) => new Date(e.expense_date).getFullYear()))
  years.add(current)
  return Array.from(years).sort((a, b) => b - a)
}

export function BigExpenses() {
  const { fmtExact } = useMoney()
  const { data: expenses = [] } = useBigExpenses()

  const [year, setYear] = useState(() => new Date().getFullYear())
  const { data: summary, isLoading } = useBigExpenseSummary(year)

  const [selectedMonth, setSelectedMonth] = useState<number | null>(null)
  const [hoverCategory, setHoverCategory] = useState<DonutDatum | null>(null)
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [sortKey, setSortKey] = useState<SortKey>('expense_date')
  const [sortAsc, setSortAsc] = useState(false)
  const [monthSortKey, setMonthSortKey] = useState<SortKey>('expense_date')
  const [monthSortAsc, setMonthSortAsc] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingExpense, setEditingExpense] = useState<BigExpense | null>(null)

  const years = useMemo(() => yearOptions(expenses), [expenses])
  const months = summary?.months ?? []
  const categories = summary?.categories ?? []
  const maxMonth = useMemo(() => Math.max(1, ...months.map((m) => m.amount_idr)), [months])
  const selected = months.find((m) => m.month === selectedMonth) ?? null

  const biggestExpense = useMemo(() => {
    const inYear = expenses.filter((e) => new Date(e.expense_date).getFullYear() === year)
    return inYear.reduce<BigExpense | null>(
      (max, e) => (!max || e.amount_idr > max.amount_idr ? e : max),
      null,
    )
  }, [expenses, year])

  const filterCategories = useMemo(
    () => Array.from(new Set(expenses.map((e) => e.category))).sort(),
    [expenses],
  )
  const rows = useMemo(() => {
    const filtered =
      categoryFilter === 'all' ? expenses : expenses.filter((e) => e.category === categoryFilter)
    const sorted = [...filtered]
    sorted.sort((a, b) => {
      const av = a[sortKey]
      const bv = b[sortKey]
      const cmp =
        typeof av === 'string' && typeof bv === 'string' ? av.localeCompare(bv) : Number(av) - Number(bv)
      return sortAsc ? cmp : -cmp
    })
    return sorted
  }, [expenses, categoryFilter, sortKey, sortAsc])
  const rowsTotal = useMemo(() => rows.reduce((s, e) => s + e.amount_idr, 0), [rows])

  const monthEntries = useMemo(() => {
    const entries = selected?.entries ?? []
    const sorted = [...entries]
    sorted.sort((a, b) => {
      const av = a[monthSortKey]
      const bv = b[monthSortKey]
      const cmp =
        typeof av === 'string' && typeof bv === 'string' ? av.localeCompare(bv) : Number(av) - Number(bv)
      return monthSortAsc ? cmp : -cmp
    })
    return sorted
  }, [selected, monthSortKey, monthSortAsc])

  function toggleSort(key: SortKey) {
    if (key === sortKey) {
      setSortAsc((v) => !v)
    } else {
      setSortKey(key)
      setSortAsc(key === 'name' || key === 'category')
    }
  }

  function toggleMonthSort(key: SortKey) {
    if (key === monthSortKey) {
      setMonthSortAsc((v) => !v)
    } else {
      setMonthSortKey(key)
      setMonthSortAsc(key === 'name' || key === 'category')
    }
  }

  function openAdd() {
    setEditingExpense(null)
    setModalOpen(true)
  }
  function openEdit(e: BigExpense) {
    setEditingExpense(e)
    setModalOpen(true)
  }

  function openEditFromMonth(e: BigExpense) {
    setSelectedMonth(null)
    openEdit(e)
  }

  return (
    <div>
      <div className="row-wrap bigexp-header">
        <div className="bigexp-header-copy">
          Every large one-off purchase, logged by hand — separate from the monthly budget.
        </div>
        <button type="button" className="btn btn-primary" onClick={openAdd}>
          + Add expense
        </button>
      </div>

      <div className="bigexp-hero-grid">
        <div className="card bigexp-hero-card">
          <div className="dash-mini-label">This year</div>
          <div className="mono bigexp-hero-value">{fmtExact(summary?.total_idr ?? 0)}</div>
          <div className="bigexp-hero-sub">
            {summary?.entries_count ?? 0} {entryWord(summary?.entries_count ?? 0)}
          </div>
        </div>
        <div className="card bigexp-hero-card">
          <div className="dash-mini-label">Biggest expense · {year}</div>
          <div className="mono bigexp-hero-value">
            {biggestExpense ? fmtExact(biggestExpense.amount_idr) : '—'}
          </div>
          <div className="bigexp-hero-sub">{biggestExpense?.name ?? 'No entries yet'}</div>
        </div>
        <div className="card bigexp-hero-card">
          <div className="dash-mini-label">Biggest category · {year}</div>
          <div className="mono bigexp-hero-value">
            {categories[0] ? fmtExact(categories[0].amount_idr) : '—'}
          </div>
          <div className="bigexp-hero-sub">{categories[0]?.category ?? 'No entries yet'}</div>
        </div>
      </div>

      <div className="row-wrap bigexp-year-row">
        <div className="card-title">Through the year</div>
        <div className="segmented">
          {years.map((y) => (
            <button
              key={y}
              type="button"
              className={'segmented-btn' + (year === y ? ' segmented-btn-active' : '')}
              onClick={() => {
                setYear(y)
                setSelectedMonth(null)
              }}
            >
              {y}
            </button>
          ))}
        </div>
      </div>

      <div className="bigexp-grid">
        <div className="card">
          <div className="card-title">By month · {year}</div>
          {isLoading && <div className="empty-state">Loading…</div>}
          {!isLoading && (summary?.total_idr ?? 0) === 0 && (
            <div className="empty-state">No big expenses logged in {year} yet.</div>
          )}
          {months.map((m) => (
            <div
              key={m.month}
              className={'bigexp-month-row' + (selectedMonth === m.month ? ' bigexp-month-row-active' : '')}
              onClick={() => setSelectedMonth(m.month)}
            >
              <div className="bigexp-month-head">
                <span className="bigexp-month-name">{m.label}</span>
                <span className="mono bigexp-month-val">{fmtExact(m.amount_idr)}</span>
              </div>
              <div className="progress-track">
                <div
                  className="progress-fill"
                  style={{ width: `${(m.amount_idr / maxMonth) * 100}%`, background: 'var(--blue)' }}
                />
              </div>
            </div>
          ))}
        </div>

        <div className="card">
          <div className="card-title">By category · {year}</div>
          <div className="bigexp-donut-wrap">
            <DonutChart
              data={categories.map((c) => ({ value: c.amount_idr, color: c.color_oklch, label: c.category }))}
              onHover={setHoverCategory}
            />
            <div className="bigexp-donut-center">
              <div className="bigexp-donut-center-label">{hoverCategory ? hoverCategory.label : year}</div>
              <div className="bigexp-donut-center-value mono">
                {fmtExact(hoverCategory ? hoverCategory.value : summary?.total_idr ?? 0)}
              </div>
            </div>
          </div>
          {categories.length === 0 && <div className="empty-state">No entries yet.</div>}
          {categories.map((c) => (
            <div className="bigexp-cat-row" key={c.category}>
              <span className="bigexp-cat-swatch" style={{ background: c.color_oklch }} />
              <span className="bigexp-cat-name">{c.category}</span>
              <span className="mono bigexp-cat-val">{fmtExact(c.amount_idr)}</span>
              <span className="mono bigexp-cat-pct">{c.percent.toFixed(1)}%</span>
            </div>
          ))}
        </div>
      </div>

      <div className="chips-row bigexp-chips">
        <button
          type="button"
          className={'chip' + (categoryFilter === 'all' ? ' chip-active' : '')}
          onClick={() => setCategoryFilter('all')}
        >
          All categories
        </button>
        {filterCategories.map((c) => (
          <button
            key={c}
            type="button"
            className={'chip' + (categoryFilter === c ? ' chip-active' : '')}
            onClick={() => setCategoryFilter(c)}
          >
            {c}
          </button>
        ))}
      </div>

      <div className="card bigexp-table-card">
        <div className="bigexp-row bigexp-row-head">
          {COLUMNS.map((c) => (
            <span
              key={c.key}
              className={'bigexp-sort' + (c.numeric ? ' bigexp-num' : '')}
              onClick={() => toggleSort(c.key)}
            >
              {c.label}
              {sortKey === c.key && <span className="bigexp-sort-arrow">{sortAsc ? '▲' : '▼'}</span>}
            </span>
          ))}
          <span />
        </div>

        {rows.length === 0 && <div className="empty-state">No big expenses logged yet.</div>}

        {rows.map((e) => (
          <div className="bigexp-row" key={e.id} onClick={() => openEdit(e)}>
            <span className="mono">{formatShortDate(e.expense_date)}</span>
            <span className="bigexp-name">{e.name}</span>
            <span>
              <span className="source-badge bigexp-cat-badge">{e.category}</span>
            </span>
            <span className="mono bigexp-num">{fmtExact(e.amount_idr)}</span>
            <button type="button" title="Edit" className="a-edit-btn" onClick={() => openEdit(e)}>
              ✎
            </button>
          </div>
        ))}

        {rows.length > 0 && (
          <div className="bigexp-net-row">
            <span className="bigexp-net-label">
              {categoryFilter === 'all' ? 'All expenses' : categoryFilter}
            </span>
            <span className="mono bigexp-net-sub">{rows.length} {entryWord(rows.length)}</span>
            <span className="mono bigexp-net-total">{fmtExact(rowsTotal)}</span>
          </div>
        )}
      </div>

      <BigExpenseModal
        open={modalOpen}
        onClose={() => {
          setModalOpen(false)
          setEditingExpense(null)
        }}
        editingExpense={editingExpense}
      />

      <Modal
        open={selected !== null}
        onClose={() => setSelectedMonth(null)}
        title={selected ? `${selected.label} ${year}` : ''}
        wide
        footer={
          <button type="button" className="btn btn-secondary" onClick={() => setSelectedMonth(null)}>
            Close
          </button>
        }
      >
        {selected && selected.entries.length === 0 && (
          <div className="empty-state">No entries in {selected.label}.</div>
        )}
        {selected && selected.entries.length > 0 && (
          <div className="bigexp-modal-table">
            <div className="bigexp-row bigexp-row-head">
              {COLUMNS.map((c) => (
                <span
                  key={c.key}
                  className={'bigexp-sort' + (c.numeric ? ' bigexp-num' : '')}
                  onClick={() => toggleMonthSort(c.key)}
                >
                  {c.label}
                  {monthSortKey === c.key && (
                    <span className="bigexp-sort-arrow">{monthSortAsc ? '▲' : '▼'}</span>
                  )}
                </span>
              ))}
              <span />
            </div>
            {monthEntries.map((e) => (
              <div className="bigexp-row" key={e.id} onClick={() => openEditFromMonth(e)}>
                <span className="mono">{formatShortDate(e.expense_date)}</span>
                <span className="bigexp-name">{e.name}</span>
                <span>
                  <span className="source-badge bigexp-cat-badge">{e.category}</span>
                </span>
                <span className="mono bigexp-num">{fmtExact(e.amount_idr)}</span>
                <button
                  type="button"
                  title="Edit"
                  className="a-edit-btn"
                  onClick={() => openEditFromMonth(e)}
                >
                  ✎
                </button>
              </div>
            ))}
          </div>
        )}
      </Modal>
    </div>
  )
}

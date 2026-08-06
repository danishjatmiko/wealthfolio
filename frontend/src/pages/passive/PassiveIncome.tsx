import { useMemo, useState } from 'react'
import { DonutChart, type DonutDatum } from '../../components/charts/DonutChart'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useCategories } from '../../hooks/useCategories'
import { useIncomeCalendar, usePassiveIncome } from '../../hooks/usePassiveIncome'
import { formatShortDate } from '../../lib/format'
import { PassiveIncomeModal } from './PassiveIncomeModal'
import type { IncomeEntry, IncomeMonth, PassiveIncomeEntry } from '../../types'
import './PassiveIncome.css'

// Bond coupons repeat on the same dates every year, so a short forward
// window is all that's useful; hand-logged income only ever exists in the
// past. Together that makes last year through the year after next the range
// worth offering — enough history to check what actually landed, enough
// forward to see the steady coupon state.
function yearOptions(): number[] {
  const now = new Date().getFullYear()
  return [now - 1, now, now + 1, now + 2]
}

interface MonthSegment {
  key: string
  share: number
  color: string
  label: string
  amountIdr: number
}

/** What the floating tooltip over a stacked bar is currently describing.
 *  Positioned in viewport coordinates, so it follows the pointer without
 *  needing any offset parent to measure against. */
interface BarTooltip {
  x: number
  y: number
  month: string
  label: string
  amountIdr: number
  color: string
}

/** How the month-detail list is ordered. Date ascending is the default
 *  because a month reads as a sequence of payments; the others are for
 *  answering "what was the biggest", grouping by where it came from, or
 *  finding a name. */
type DetailSort = 'date' | 'amount' | 'category' | 'name'

const DETAIL_SORTS: { key: DetailSort; label: string }[] = [
  { key: 'date', label: 'Date' },
  { key: 'amount', label: 'Amount' },
  { key: 'category', label: 'Category' },
  { key: 'name', label: 'Name' },
]

/** Every month the given category paid into, for the expanded legend row.
 *  Derived from the month slices already on the calendar — the backend
 *  doesn't need a second breakdown for this. */
function categoryMonths(months: IncomeMonth[], categoryKey: string) {
  const out: { month: number; label: string; amountIdr: number }[] = []
  for (const m of months) {
    const slice = m.slices.find((s) => s.category_key === categoryKey)
    if (slice) out.push({ month: m.month, label: m.label, amountIdr: slice.amount_idr })
  }
  return out
}

/** Orders a month's entries for the detail list. Amount sorts on the Rupiah
 *  figure even for coupons, so dollar and Rupiah rows stay comparable.
 *  Category groups by asset class and keeps each group in date order, so a
 *  group reads the same way the whole list does when sorted by date. */
function sortEntries(entries: IncomeEntry[], key: DetailSort, desc: boolean): IncomeEntry[] {
  const dir = desc ? -1 : 1
  return [...entries].sort((a, b) => {
    switch (key) {
      case 'amount':
        return (a.amount_idr - b.amount_idr) * dir
      case 'name':
        return a.name.localeCompare(b.name) * dir
      case 'category': {
        const cat = a.category_label.localeCompare(b.category_label)
        if (cat !== 0) return cat * dir
        if (a.pay_date !== b.pay_date) return a.pay_date < b.pay_date ? -1 : 1
        return a.name.localeCompare(b.name)
      }
      default: {
        if (a.pay_date !== b.pay_date) return a.pay_date < b.pay_date ? -dir : dir
        return a.name.localeCompare(b.name)
      }
    }
  })
}

/** Splits a month's bar into one coloured segment per category that paid
 *  into it, so the bar shows composition rather than just size.
 *
 *  The bar's overall length already encodes the month's net (see caller);
 *  these segments divide that length by each category's share of the
 *  month's *positive* income. A category that lost money inside an
 *  otherwise-positive month is left out — its loss is already netted into
 *  the length, and a backwards segment can't be drawn. A month that nets
 *  negative overall is a single red bar, matching how its total reads. */
function monthSegments(m: IncomeMonth): MonthSegment[] {
  if (m.amount_idr <= 0) {
    if (m.amount_idr === 0) return []
    return [
      { key: 'loss', share: 1, color: 'var(--red)', label: 'Net loss', amountIdr: m.amount_idr },
    ]
  }
  const positive = m.slices.filter((s) => s.amount_idr > 0)
  const sum = positive.reduce((acc, s) => acc + s.amount_idr, 0)
  if (sum === 0) return []
  return positive.map((s) => ({
    key: s.category_key,
    share: s.amount_idr / sum,
    color: s.color_oklch,
    label: s.label,
    amountIdr: s.amount_idr,
  }))
}

export function PassiveIncome() {
  const { fmtExact, fmtUsd } = useMoney()
  const { data: categories = [] } = useCategories()
  const { data: entries = [] } = usePassiveIncome()

  const [year, setYear] = useState(() => new Date().getFullYear())
  const { data: calendar, isLoading } = useIncomeCalendar(year)

  const [selectedMonth, setSelectedMonth] = useState<number | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingEntry, setEditingEntry] = useState<PassiveIncomeEntry | null>(null)
  const [hoverCategory, setHoverCategory] = useState<DonutDatum | null>(null)
  const [expandedCategory, setExpandedCategory] = useState<string | null>(null)
  const [barTooltip, setBarTooltip] = useState<BarTooltip | null>(null)
  const [detailSort, setDetailSort] = useState<DetailSort>('date')
  const [detailDesc, setDetailDesc] = useState(false)

  const months = calendar?.months ?? []
  // Distinct from `categories` above: those are the asset classes the modal
  // offers, these are the ones that actually paid out this year.
  const incomeCategories = calendar?.categories ?? []
  // Bars are scaled by magnitude so a loss-making month still draws a bar
  // proportional to how big the loss was.
  const maxMonth = useMemo(
    () => Math.max(1, ...months.map((m) => Math.abs(m.amount_idr))),
    [months],
  )
  const selected = months.find((m) => m.month === selectedMonth) ?? null
  // "Logged income" folds in receivable payments alongside hand-logged
  // dividends/capital/redemptions — everything Rupiah-denominated that
  // isn't a bond coupon, which gets its own USD-native hero card.
  const loggedIncomeIdr = (calendar?.manual_total_idr ?? 0) + (calendar?.receivable_total_idr ?? 0)

  function openAdd() {
    setEditingEntry(null)
    setModalOpen(true)
  }

  // Only manual entries are editable here — a coupon is derived from the
  // bond ledger and has no row of its own, so it carries an empty id and
  // has to be changed on the Bonds page instead.
  function openEntry(e: IncomeEntry) {
    if (e.kind !== 'manual') return
    const entry = entries.find((x) => x.id === e.id)
    if (!entry) return
    setEditingEntry(entry)
    setModalOpen(true)
  }

  const hasNothing = !isLoading && months.every((m) => m.entries.length === 0)

  return (
    <div>
      <div className="row-wrap passive-header">
        <div className="passive-header-copy">
          Every source of passive income, month by month — bond coupons, receivable loan payments,
          and the dividends, capital gains and redemptions you log by hand. Click a month to see the
          exact dates.
        </div>
        <div className="btn-group">
          <div className="segmented">
            {yearOptions().map((y) => (
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
          <button type="button" className="btn btn-primary" onClick={openAdd}>
            + Add income
          </button>
        </div>
      </div>

      <div className="passive-hero-grid">
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Bond coupons · {year}</div>
          <div className="mono passive-hero-value">{fmtUsd(calendar?.coupon_total_usd ?? 0)}</div>
          <div className="mono passive-hero-sub">{fmtExact(calendar?.coupon_total_idr ?? 0)}</div>
        </div>
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Logged income · {year}</div>
          <div
            className={
              'mono passive-hero-value' +
              (loggedIncomeIdr < 0 ? ' passive-negative' : '')
            }
          >
            {fmtExact(loggedIncomeIdr)}
          </div>
          <div className="mono passive-hero-sub">Dividends, capital, redemptions, receivable payments</div>
        </div>
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Combined · {year}</div>
          <div className="mono passive-hero-value">{fmtExact(calendar?.total_idr ?? 0)}</div>
          <div className="mono passive-hero-sub">
            {fmtExact(Math.round((calendar?.total_idr ?? 0) / 12))} / month
          </div>
        </div>
      </div>

      <div className="passive-grid">
        <div className="card">
          <div className="card-title">Income by month · {year}</div>
          {isLoading && <div className="empty-state">Loading…</div>}
          {hasNothing && (
            <div className="empty-state">
              Nothing landed in {year}. Add income above, or log bond purchases on the Bonds page to
              see their coupons here.
            </div>
          )}
          {months.map((m) => (
            <div
              key={m.month}
              className={
                'passive-month-row' + (selectedMonth === m.month ? ' passive-month-row-active' : '')
              }
              onClick={() => setSelectedMonth(selectedMonth === m.month ? null : m.month)}
            >
              <div className="passive-month-head">
                <span className="passive-month-name">{m.label}</span>
                <span
                  className={
                    'mono passive-month-total' + (m.amount_idr < 0 ? ' passive-negative' : '')
                  }
                >
                  {fmtExact(m.amount_idr)}
                </span>
                <span className="mono passive-month-split">
                  {m.coupon_idr !== 0 && <span title="Bond coupons">◈ {fmtUsd(m.coupon_usd)}</span>}
                  {m.coupon_idr !== 0 && (m.manual_idr !== 0 || m.receivable_idr !== 0) && ' · '}
                  {m.receivable_idr !== 0 && (
                    <span title="Receivable payments">⇄ {fmtExact(m.receivable_idr)}</span>
                  )}
                  {m.receivable_idr !== 0 && m.manual_idr !== 0 && ' · '}
                  {m.manual_idr !== 0 && <span title="Logged income">✎ {fmtExact(m.manual_idr)}</span>}
                </span>
              </div>
              <div className="progress-track">
                <div
                  className="passive-month-bar"
                  style={{ width: `${(Math.abs(m.amount_idr) / maxMonth) * 100}%` }}
                >
                  {monthSegments(m).map((seg) => (
                    <div
                      key={seg.key}
                      className="passive-month-seg"
                      style={{ width: `${seg.share * 100}%`, background: seg.color }}
                      onMouseEnter={(ev) =>
                        setBarTooltip({
                          x: ev.clientX,
                          y: ev.clientY,
                          month: m.label,
                          label: seg.label,
                          amountIdr: seg.amountIdr,
                          color: seg.color,
                        })
                      }
                      onMouseMove={(ev) =>
                        setBarTooltip((t) => (t ? { ...t, x: ev.clientX, y: ev.clientY } : t))
                      }
                      onMouseLeave={() => setBarTooltip(null)}
                    />
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="passive-side">
          <div className="card">
            <div className="card-title">Where it came from · {year}</div>
            <div className="passive-donut-wrap">
              <DonutChart
                data={incomeCategories.map((c) => ({
                  // A category that nets negative over the year has no share
                  // of what came in; a slice can't be drawn backwards, so it
                  // drops out of the ring but stays in the list below.
                  value: Math.max(0, c.amount_idr),
                  color: c.color_oklch,
                  label: c.label,
                }))}
                onHover={setHoverCategory}
                onSelect={(d) => {
                  const hit = incomeCategories.find((c) => c.label === d.label)
                  if (!hit) return
                  setExpandedCategory((k) => (k === hit.category_key ? null : hit.category_key))
                }}
              />
              <div className="passive-donut-center">
                <div className="passive-donut-center-label">
                  {hoverCategory ? hoverCategory.label : year}
                </div>
                <div className="passive-donut-center-value mono">
                  {fmtExact(hoverCategory ? hoverCategory.value : (calendar?.total_idr ?? 0))}
                </div>
              </div>
            </div>

            {incomeCategories.length === 0 && (
              <div className="empty-state">Nothing paid out in {year} yet.</div>
            )}
            {incomeCategories.map((c) => (
              <div key={c.category_key}>
                <div
                  className={
                    'passive-cat-row' +
                    (hoverCategory?.label === c.label ? ' passive-cat-row-active' : '') +
                    (expandedCategory === c.category_key ? ' passive-cat-row-open' : '')
                  }
                  onMouseEnter={() =>
                    setHoverCategory({
                      value: Math.max(0, c.amount_idr),
                      color: c.color_oklch,
                      label: c.label,
                    })
                  }
                  onMouseLeave={() => setHoverCategory(null)}
                  onClick={() =>
                    setExpandedCategory((k) => (k === c.category_key ? null : c.category_key))
                  }
                >
                  <span className="passive-cat-name">
                    <span className="passive-cat-swatch" style={{ background: c.color_oklch }} />
                    {c.label}
                    <span className="passive-cat-count">
                      {c.count} {c.count === 1 ? 'payment' : 'payments'}
                    </span>
                  </span>
                  <span className="passive-cat-amounts">
                    <span
                      className={
                        'mono passive-cat-val' + (c.amount_idr < 0 ? ' passive-negative' : '')
                      }
                    >
                      {fmtExact(c.amount_idr)}
                    </span>
                    <span className="mono passive-cat-pct">{c.percent.toFixed(1)}%</span>
                  </span>
                </div>

                {expandedCategory === c.category_key && (
                  <div className="passive-cat-months">
                    {categoryMonths(months, c.category_key).map((cm) => (
                      <div
                        className="passive-cat-month"
                        key={cm.month}
                        onClick={() => setSelectedMonth(cm.month)}
                      >
                        <span className="passive-cat-month-name">{cm.label}</span>
                        <span
                          className={
                            'mono passive-cat-month-val' +
                            (cm.amountIdr < 0 ? ' passive-negative' : '')
                          }
                        >
                          {fmtExact(cm.amountIdr)}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}

            <div className="passive-donut-note">
              Click a slice or a row to see which months it paid in. Bond coupons count under Bonds
              USD and receivable payments under Piutang; coupons dated before a bond was bought — and
              payments before a loan started or after its term ends — are excluded, so a start year
              looks lighter than the steady state. A category that nets a loss leaves the ring but
              stays listed.
            </div>
          </div>

          <div className="card">
            <div className="card-title">
              {selected ? `${selected.label} ${year}` : 'Pick a month'}
            </div>
            {!selected && (
              <div className="empty-state">Click a month to see its exact payment dates.</div>
            )}
            {selected && selected.entries.length === 0 && (
              <div className="empty-state">Nothing landed in {selected.label}.</div>
            )}
            {selected && selected.entries.length > 1 && (
              <div className="passive-sort-bar">
                {DETAIL_SORTS.map(({ key, label }) => (
                  <button
                    key={key}
                    type="button"
                    className={
                      'passive-sort-btn' + (detailSort === key ? ' passive-sort-btn-active' : '')
                    }
                    // Clicking the active field flips direction; clicking a
                    // different one switches to it at its natural default —
                    // oldest-first for dates, biggest-first for amounts, A–Z
                    // for categories and names.
                    onClick={() => {
                      if (detailSort === key) {
                        setDetailDesc((d) => !d)
                      } else {
                        setDetailSort(key)
                        setDetailDesc(key === 'amount')
                      }
                    }}
                  >
                    {label}
                    {detailSort === key && (
                      <span className="passive-sort-arrow">{detailDesc ? '↓' : '↑'}</span>
                    )}
                  </button>
                ))}
              </div>
            )}
            {sortEntries(selected?.entries ?? [], detailSort, detailDesc).map((e) => (
              <div
                className={
                  'passive-entry-row' + (e.kind === 'manual' ? ' passive-entry-row-editable' : '')
                }
                key={`${e.kind}-${e.id || e.name}-${e.pay_date}`}
                onClick={() => openEntry(e)}
              >
                <div className="passive-entry-info">
                  <div className="passive-entry-name">
                    {/* Ties the row back to its slice in the bar and the
                        ring — and makes a category sort legible, since the
                        badge next to it is the income type, not the
                        category. */}
                    <span
                      className="passive-entry-swatch"
                      style={{ background: e.color_oklch }}
                      title={e.category_label}
                    />
                    {e.name}
                    {e.source && <span className="source-badge">{e.source}</span>}
                  </div>
                  <div className="passive-entry-date">{formatShortDate(e.pay_date)}</div>
                  {e.note && <div className="passive-entry-note">{e.note}</div>}
                </div>
                <div className="passive-entry-amounts">
                  {e.kind === 'coupon' ? (
                    <>
                      <div className="mono passive-entry-primary">{fmtUsd(e.amount_usd)}</div>
                      <div className="mono passive-entry-secondary">{fmtExact(e.amount_idr)}</div>
                    </>
                  ) : (
                    <div
                      className={
                        'mono passive-entry-primary' +
                        (e.amount_idr < 0 ? ' passive-negative' : '')
                      }
                    >
                      {fmtExact(e.amount_idr)}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Viewport-positioned, so it escapes the month card's overflow and
          needs no offset parent. Nudged up-right of the pointer, and
          flipped left near the right edge so it never runs off-screen. */}
      {barTooltip && (
        <div
          className="passive-bar-tooltip"
          style={{
            left: barTooltip.x + (barTooltip.x > window.innerWidth - 200 ? -180 : 14),
            top: barTooltip.y - 46,
          }}
        >
          <div className="passive-bar-tooltip-label">
            <span className="passive-cat-swatch" style={{ background: barTooltip.color }} />
            {barTooltip.label}
          </div>
          <div className="mono passive-bar-tooltip-value">{fmtExact(barTooltip.amountIdr)}</div>
          <div className="passive-bar-tooltip-sub">{barTooltip.month}</div>
        </div>
      )}

      <PassiveIncomeModal
        open={modalOpen}
        onClose={() => {
          setModalOpen(false)
          setEditingEntry(null)
        }}
        categories={categories}
        editingEntry={editingEntry}
      />
    </div>
  )
}

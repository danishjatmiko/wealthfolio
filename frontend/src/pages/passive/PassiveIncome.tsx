import { useMemo, useState } from 'react'
import { DonutChart } from '../../components/charts/DonutChart'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useCategories } from '../../hooks/useCategories'
import { useCouponCalendar } from '../../hooks/useBondPurchases'
import { usePassiveIncome } from '../../hooks/usePassiveIncome'
import { formatShortDate } from '../../lib/format'
import { PassiveIncomeModal } from './PassiveIncomeModal'
import type { PassiveIncomeSource } from '../../types'
import './PassiveIncome.css'

// Coupons repeat on the same dates every year, so a short forward window is
// all that's useful: this year answers "what am I actually receiving", and
// the next two show the steady state once every bond has cleared its
// purchase year.
function yearOptions(): number[] {
  const now = new Date().getFullYear()
  return [now, now + 1, now + 2]
}

export function PassiveIncome() {
  const { fmtExact, fmtUsd } = useMoney()
  const { data: categories = [] } = useCategories()
  const { data: sources = [] } = usePassiveIncome()

  const [year, setYear] = useState(() => new Date().getFullYear())
  const { data: calendar, isLoading } = useCouponCalendar(year)

  const [selectedMonth, setSelectedMonth] = useState<number | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingSource, setEditingSource] = useState<PassiveIncomeSource | null>(null)

  const months = calendar?.months ?? []
  const maxMonth = useMemo(() => Math.max(1, ...months.map((m) => m.amount_usd)), [months])
  const selected = months.find((m) => m.month === selectedMonth) ?? null
  const manualMax = useMemo(() => Math.max(1, ...sources.map((s) => s.per_year_idr)), [sources])

  function openAdd() {
    setEditingSource(null)
    setModalOpen(true)
  }
  function openEdit(s: PassiveIncomeSource) {
    setEditingSource(s)
    setModalOpen(true)
  }

  return (
    <div>
      <div className="row-wrap passive-header">
        <div className="passive-header-copy">
          What your bonds pay out, month by month. Click a month to see the exact dates.
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
            + Add income source
          </button>
        </div>
      </div>

      <div className="passive-hero-grid">
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Bond coupons / year</div>
          <div className="mono passive-hero-value">{fmtUsd(calendar?.total_usd ?? 0)}</div>
          <div className="mono passive-hero-sub">{fmtExact(calendar?.total_idr ?? 0)}</div>
        </div>
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Other passive / year</div>
          <div className="mono passive-hero-value">
            {fmtExact(calendar?.manual_per_year_idr ?? 0)}
          </div>
          <div className="mono passive-hero-sub">
            {fmtExact(calendar?.manual_per_month_idr ?? 0)} / month
          </div>
        </div>
        <div className="card passive-hero-card">
          <div className="dash-mini-label">Combined / year</div>
          <div className="mono passive-hero-value">
            {fmtExact(calendar?.combined_per_year_idr ?? 0)}
          </div>
          <div className="mono passive-hero-sub">
            {fmtExact(Math.round((calendar?.combined_per_year_idr ?? 0) / 12))} / month
          </div>
        </div>
      </div>

      <div className="passive-grid">
        <div className="card">
          <div className="card-title">Coupons by month · {year}</div>
          {isLoading && <div className="empty-state">Loading…</div>}
          {!isLoading && (calendar?.total_usd ?? 0) === 0 && (
            <div className="empty-state">
              No bond coupons fall in {year}. Add purchases on the Bonds page to see them here.
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
                <span className="passive-month-name">
                  <span className="passive-month-swatch" style={{ background: m.color_oklch }} />
                  {m.label}
                </span>
                <span className="mono passive-month-usd">{fmtUsd(m.amount_usd)}</span>
                <span className="mono passive-month-idr">{fmtExact(m.amount_idr)}</span>
              </div>
              <div className="progress-track">
                <div
                  className="progress-fill"
                  style={{
                    width: `${(m.amount_usd / maxMonth) * 100}%`,
                    background: m.color_oklch,
                  }}
                />
              </div>
            </div>
          ))}
        </div>

        <div className="passive-side">
          <div className="card">
            <div className="card-title">Share of the year</div>
            <div className="passive-donut-wrap">
              <DonutChart
                data={months.map((m) => ({
                  value: m.amount_usd,
                  color: m.color_oklch,
                  label: m.label,
                }))}
                selectedLabel={selected?.label ?? null}
              />
              <div className="passive-donut-center">
                <div className="passive-donut-center-label">{selected ? selected.label : year}</div>
                <div className="passive-donut-center-value mono">
                  {fmtUsd(selected ? selected.amount_usd : (calendar?.total_usd ?? 0))}
                </div>
              </div>
            </div>
            <div className="passive-donut-note">
              Coupons dated before a bond was bought are excluded, so a purchase year looks lighter
              than the steady state — pick a later year for the full pattern.
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
              <div className="empty-state">No coupons in {selected.label}.</div>
            )}
            {selected?.entries.map((e) => (
              <div className="passive-entry-row" key={`${e.bond_name}-${e.pay_date}`}>
                <div className="passive-entry-info">
                  <div className="passive-entry-name">
                    {e.bond_name}
                    {e.platform && <span className="source-badge">{e.platform}</span>}
                  </div>
                  <div className="passive-entry-date">{formatShortDate(e.pay_date)}</div>
                </div>
                <div className="passive-entry-amounts">
                  <div className="mono passive-entry-usd">{fmtUsd(e.amount_usd)}</div>
                  <div className="mono passive-entry-idr">{fmtExact(e.amount_idr)}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="card passive-manual-card">
        <div className="card-title">Other passive income (manual)</div>
        {sources.length === 0 && <div className="empty-state">No manual income sources yet.</div>}
        {sources.map((s) => (
          <div className="passive-row" key={s.id} onClick={() => openEdit(s)}>
            <div className="passive-row-head">
              <span>{s.name || s.category_label}</span>
              <span className="mono passive-row-val">{fmtExact(s.per_year_idr)}</span>
            </div>
            <div className="progress-track">
              <div
                className="progress-fill"
                style={{
                  width: `${(s.per_year_idr / manualMax) * 100}%`,
                  background: 'var(--blue)',
                }}
              />
            </div>
          </div>
        ))}
      </div>

      <PassiveIncomeModal
        open={modalOpen}
        onClose={() => {
          setModalOpen(false)
          setEditingSource(null)
        }}
        categories={categories}
        editingSource={editingSource}
      />
    </div>
  )
}

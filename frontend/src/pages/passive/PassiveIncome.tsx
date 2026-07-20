import { useMemo, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useCategories } from '../../hooks/useCategories'
import { useDashboard } from '../../hooks/useDashboard'
import { usePassiveIncome } from '../../hooks/usePassiveIncome'
import { PassiveIncomeModal } from './PassiveIncomeModal'
import type { PassiveIncomeSource } from '../../types'
import './PassiveIncome.css'

export function PassiveIncome() {
  const { fmt } = useMoney()
  const { data: categories = [] } = useCategories()
  const { data: dashboard } = useDashboard()
  const { data: sources = [] } = usePassiveIncome()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingSource, setEditingSource] = useState<PassiveIncomeSource | null>(null)

  const maxVal = useMemo(() => Math.max(1, ...sources.map((s) => s.per_year_idr)), [sources])

  const passive = dashboard?.passive
  const gap = passive ? Math.max(0, passive.target_per_year_idr - passive.per_year_idr) : 0
  const pct = passive ? Math.round(passive.percent) : 0

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
        <div className="passive-header-copy">Master list of estimated annual passive income by source.</div>
        <button type="button" className="btn btn-primary" onClick={openAdd}>
          + Add income source
        </button>
      </div>

      <div className="passive-grid">
        <div className="card">
          <div className="card-title">Estimated passive income per category / year</div>
          {sources.length === 0 && <div className="empty-state">No income sources yet.</div>}
          {sources.map((s) => (
            <div className="passive-row" key={s.id} onClick={() => openEdit(s)}>
              <div className="passive-row-head">
                <span>{s.name || s.category_label}</span>
                <span className="mono passive-row-val">{fmt(s.per_year_idr)}</span>
              </div>
              <div className="progress-track">
                <div
                  className="progress-fill"
                  style={{ width: `${(s.per_year_idr / maxVal) * 100}%`, background: 'var(--blue)' }}
                />
              </div>
            </div>
          ))}
        </div>

        <div className="passive-side">
          <div className="card">
            <div className="dash-mini-label">Total per year</div>
            <div className="mono passive-total-value">{passive ? fmt(passive.per_year_idr) : '—'}</div>
            <div className="passive-total-note">
              of {passive ? fmt(passive.target_per_year_idr) : '—'} target · {pct}%
            </div>
            <div className="progress-track passive-total-track">
              <div
                className="progress-fill"
                style={{ width: `${Math.min(100, pct)}%`, background: 'var(--blue)' }}
              />
            </div>
            <div className="passive-gap">{fmt(gap)} still needed</div>
          </div>
          <div className="card">
            <div className="dash-mini-label">Per month now</div>
            <div className="mono dash-mini-value">{passive ? fmt(passive.per_month_idr) : '—'}</div>
            <div className="dash-mini-note">
              target {passive ? fmt(passive.per_month_target_idr) : '—'} / month
            </div>
          </div>
        </div>
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

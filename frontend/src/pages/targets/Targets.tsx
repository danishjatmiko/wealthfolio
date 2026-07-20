import { useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useTargets } from '../../hooks/useTargets'
import { TargetModal } from './TargetModal'
import type { Target, TargetMetricType } from '../../types'
import './Targets.css'

const METRIC_COLOR: Record<TargetMetricType, string> = {
  equity: 'var(--accent)',
  gold_grams: 'var(--cat-logam-mulia)',
  passive_income: 'var(--blue)',
  debt_ratio: 'var(--cat-uang-tunai)',
  custom: 'var(--cat-crypto)',
}

function trimNumber(n: number): string {
  const s = n.toFixed(2)
  return s.replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function formatMetricValue(value: number, metricType: TargetMetricType, unit: string, money: (v: number) => string): string {
  if (metricType === 'equity' || metricType === 'passive_income') return money(value)
  return `${trimNumber(value)}${unit}`
}

export function Targets() {
  const { fmt } = useMoney()
  const { data: targets = [] } = useTargets()
  const [modalOpen, setModalOpen] = useState(false)
  const [editingTarget, setEditingTarget] = useState<Target | null>(null)

  function openAdd() {
    setEditingTarget(null)
    setModalOpen(true)
  }
  function openEdit(t: Target) {
    setEditingTarget(t)
    setModalOpen(true)
  }

  return (
    <div>
      <div className="row-wrap targets-header">
        <div className="targets-header-copy">Manage your goals — click a card to edit, or add a new target.</div>
        <button type="button" className="btn btn-primary" onClick={openAdd}>
          + Add target
        </button>
      </div>

      <div className="targets-grid">
        {targets.length === 0 && <div className="empty-state">No targets set yet.</div>}
        {targets.map((t) => {
          const remaining = t.target_value - t.current_value
          const pct = Math.max(0, Math.min(100, t.percent))
          const color = METRIC_COLOR[t.metric_type]
          const remainingAmount = formatMetricValue(Math.abs(remaining), t.metric_type, t.unit, fmt)
          let toGoText: string
          if (t.lower_is_better) {
            toGoText = remaining >= 0 ? `${remainingAmount} headroom` : `${remainingAmount} over target`
          } else {
            toGoText = remaining > 0 ? `${remainingAmount} to go` : 'Target reached'
          }

          return (
            <div className="card target-card" key={t.id} onClick={() => openEdit(t)}>
              <div className="target-card-head">
                <div className="card-title-inline">{t.name}</div>
                <div className="mono target-card-pct">{t.percent.toFixed(1)}%</div>
              </div>
              <div className="progress-track target-track">
                <div className="progress-fill" style={{ width: `${pct}%`, background: color }} />
              </div>
              <div className="target-card-row">
                <span>
                  Now <b className="mono target-card-val">{formatMetricValue(t.current_value, t.metric_type, t.unit, fmt)}</b>
                </span>
                <span>
                  Target <b className="mono target-card-val">{formatMetricValue(t.target_value, t.metric_type, t.unit, fmt)}</b>
                </span>
              </div>
              <div className="target-card-togo">{toGoText}</div>
            </div>
          )
        })}
      </div>

      <TargetModal
        open={modalOpen}
        onClose={() => {
          setModalOpen(false)
          setEditingTarget(null)
        }}
        editingTarget={editingTarget}
      />
    </div>
  )
}

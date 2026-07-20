import { useMemo, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useDebts } from '../../hooks/useDebts'
import { useDashboard } from '../../hooks/useDashboard'
import { useTargets } from '../../hooks/useTargets'
import { DebtModal } from './DebtModal'
import type { Debt, DebtDirection } from '../../types'
import './Debts.css'

export function Debts() {
  const { fmt } = useMoney()
  const { data: debts = [] } = useDebts()
  const { data: dashboard } = useDashboard()
  const { data: targets = [] } = useTargets()

  const [modalDirection, setModalDirection] = useState<DebtDirection | null>(null)
  const [editingDebt, setEditingDebt] = useState<Debt | null>(null)

  const myDebts = useMemo(() => debts.filter((d) => d.direction === 'i_owe'), [debts])
  const receivables = useMemo(() => debts.filter((d) => d.direction === 'owed_to_me'), [debts])

  const ratioTarget = targets.find((t) => t.metric_type === 'debt_ratio')
  const ratioPct = dashboard?.debt.ratio_pct ?? 0

  function openAdd(direction: DebtDirection) {
    setEditingDebt(null)
    setModalDirection(direction)
  }
  function openEdit(d: Debt) {
    setEditingDebt(d)
    setModalDirection(d.direction)
  }
  function closeModal() {
    setModalDirection(null)
    setEditingDebt(null)
  }

  return (
    <div className="debts-grid">
      <div className="card">
        <div className="debts-card-head">
          <div className="card-title-inline">My debts</div>
          <div className="mono debts-card-total">{dashboard ? fmt(dashboard.debt.total_debt_idr) : '—'}</div>
        </div>
        {myDebts.length === 0 && <div className="empty-state">No debts logged.</div>}
        {myDebts.map((d) => (
          <div className="debt-row" key={d.id} onClick={() => openEdit(d)}>
            <div className="debt-row-info">
              <div className="debt-row-name">{d.name}</div>
              <div className="debt-row-type">{d.type}</div>
            </div>
            <span className="mono debt-row-val debt-row-val-red">{fmt(d.value_idr)}</span>
          </div>
        ))}
        <button type="button" className="btn-dashed" onClick={() => openAdd('i_owe')}>
          + Add debt
        </button>
      </div>

      <div className="card">
        <div className="debts-card-head">
          <div className="card-title-inline">Who owes me</div>
          <div className="mono debts-card-total">{dashboard ? fmt(dashboard.debt.total_receivable_idr) : '—'}</div>
        </div>
        {receivables.length === 0 && <div className="empty-state">Nobody owes you (yet).</div>}
        {receivables.map((d) => (
          <div className="debt-row" key={d.id} onClick={() => openEdit(d)}>
            <div className="debt-row-info">
              <div className="debt-row-name">{d.name}</div>
              <div className="debt-row-type">{d.type}</div>
            </div>
            <span className="mono debt-row-val debt-row-val-green">{fmt(d.value_idr)}</span>
          </div>
        ))}
        <button type="button" className="btn-dashed" onClick={() => openAdd('owed_to_me')}>
          + Add receivable
        </button>
      </div>

      <div className="debt-ratio-banner">
        <div>
          <div className="debt-ratio-label">Debt-to-equity ratio</div>
          <div className="mono debt-ratio-value">{ratioPct.toFixed(2)}%</div>
          {ratioTarget && (
            <div className="debt-ratio-note">
              Goal: keep below {ratioTarget.target_value.toFixed(2)}
              {ratioTarget.unit} — you have {Math.max(0, ratioTarget.target_value - ratioPct).toFixed(2)}
              {ratioTarget.unit} of headroom.
            </div>
          )}
        </div>
      </div>

      <DebtModal
        open={modalDirection !== null}
        onClose={closeModal}
        direction={modalDirection ?? 'i_owe'}
        editingDebt={editingDebt}
      />
    </div>
  )
}

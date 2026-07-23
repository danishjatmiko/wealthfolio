import { useEffect, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useExpensePeriods,
  useExpensePeriodById,
  useLatestExpensePeriod,
  useDeleteExpensePeriod,
} from '../../hooks/useExpensePeriods'
import { api } from '../../lib/api'
import { NewPeriodModal } from './NewPeriodModal'
import { BudgetEnvelopeModal } from './BudgetEnvelopeModal'
import { FixedExpenseModal } from './FixedExpenseModal'
import { formatShortDate } from '../../lib/format'
import type { BudgetEnvelopeDetail, FixedExpense } from '../../types'
import './MonthlyExpenses.css'

export function MonthlyExpenses() {
  const { fmt } = useMoney()
  const { showError, showSuccess } = useToast()
  const { data: periods = [] } = useExpensePeriods()
  const { data: latestPeriod } = useLatestExpensePeriod()
  const deletePeriod = useDeleteExpensePeriod()

  const [selectedId, setSelectedId] = useState<string | undefined>(undefined)
  const { data: period } = useExpensePeriodById(selectedId)

  useEffect(() => {
    if (!latestPeriod) return
    if (!selectedId) setSelectedId(latestPeriod.id)
  }, [selectedId, latestPeriod])

  const [showNewPeriod, setShowNewPeriod] = useState(false)
  const [editingEnvelope, setEditingEnvelope] = useState<BudgetEnvelopeDetail | null>(null)
  const [showEnvelopeModal, setShowEnvelopeModal] = useState(false)
  const [editingExpense, setEditingExpense] = useState<FixedExpense | null>(null)
  const [expenseModalEnvelopeId, setExpenseModalEnvelopeId] = useState<string | null>(null)
  const [showExpenseModal, setShowExpenseModal] = useState(false)

  function openAddEnvelope() {
    setEditingEnvelope(null)
    setShowEnvelopeModal(true)
  }
  function openEditEnvelope(env: BudgetEnvelopeDetail) {
    setEditingEnvelope(env)
    setShowEnvelopeModal(true)
  }
  function openAddExpense(envelopeId: string) {
    setEditingExpense(null)
    setExpenseModalEnvelopeId(envelopeId)
    setShowExpenseModal(true)
  }
  function openEditExpense(expense: FixedExpense) {
    setEditingExpense(expense)
    setExpenseModalEnvelopeId(expense.envelope_id)
    setShowExpenseModal(true)
  }

  async function handleDeletePeriod() {
    if (!period) return
    if (!window.confirm(`Delete the ${period.label} period? This removes it and everything inside it.`)) return
    try {
      await deletePeriod.mutateAsync(period.id)
      showSuccess('Period deleted.')
      // Fetch the fresh list directly rather than waiting on background
      // query invalidation to settle — picking the new id from possibly-
      // still-stale cached data would just re-select what was just deleted.
      const fresh = await api.expensePeriods.list()
      setSelectedId(fresh[0]?.id)
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  if (!period) {
    return (
      <div>
        <div className="empty-state">No expense periods yet.</div>
        <button type="button" className="btn btn-primary" onClick={() => setShowNewPeriod(true)}>
          + Start this period
        </button>
        <NewPeriodModal
          open={showNewPeriod}
          onClose={() => setShowNewPeriod(false)}
          hasExistingPeriod={false}
          latestPeriodEndDate={undefined}
          onCreated={(id) => setSelectedId(id)}
        />
      </div>
    )
  }

  return (
    <div>
      <div className="row-wrap assets-toolbar">
        <div>
          <div className="snapshot-pill">
            <span className="snapshot-pill-label">Period</span>
            <select
              className="snapshot-pill-select"
              value={selectedId ?? ''}
              onChange={(e) => setSelectedId(e.target.value)}
            >
              {periods.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>
          <div className="expense-period-range">
            {formatShortDate(period.start_date)} – {formatShortDate(period.end_date)}
          </div>
        </div>
        <div className="btn-group">
          <button type="button" className="btn btn-secondary" onClick={openAddEnvelope}>
            + Add envelope
          </button>
          <button type="button" className="btn btn-secondary" onClick={() => setShowNewPeriod(true)}>
            ⧉ New period
          </button>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleDeletePeriod}
            disabled={deletePeriod.isPending}
          >
            🗑 Delete period
          </button>
        </div>
      </div>

      <div className="expense-grid">
        {period.envelopes.map((env) => {
          const overBudget = env.actual_total_idr > env.committed_amount_idr
          const pct =
            env.committed_amount_idr > 0 ? (env.actual_total_idr / env.committed_amount_idr) * 100 : 0
          const deltaAbs = Math.abs(env.actual_total_idr - env.committed_amount_idr)

          return (
            <div className="card" key={env.id}>
              <div className="expense-card-head" onClick={() => openEditEnvelope(env)}>
                <div>
                  <div className="expense-card-category">{env.category_name}</div>
                  <div className="expense-card-title">{env.name}</div>
                </div>
                <div className="mono expense-card-total">{fmt(env.actual_total_idr)}</div>
              </div>
              <div className="expense-envelope-meta">
                <span>of {fmt(env.committed_amount_idr)} committed</span>
                <span className={overBudget ? 'expense-over' : 'expense-under'}>
                  {overBudget ? '+' : '−'}
                  {fmt(deltaAbs)} {overBudget ? 'over' : 'under'}
                </span>
              </div>
              <div className="progress-track">
                <div
                  className="progress-fill"
                  style={{
                    width: `${Math.min(100, pct)}%`,
                    background: overBudget ? 'var(--red)' : 'var(--green)',
                  }}
                />
              </div>

              {env.fixed_expenses.length === 0 && <div className="empty-state">No expenses logged yet.</div>}
              {env.fixed_expenses.map((fe) => (
                <div className="expense-row" key={fe.id} onClick={() => openEditExpense(fe)}>
                  <div className="expense-row-name">{fe.name}</div>
                  <span className="mono expense-row-val">{fmt(fe.amount_idr)}</span>
                </div>
              ))}
              <button type="button" className="btn-dashed" onClick={() => openAddExpense(env.id)}>
                + Add expense
              </button>
            </div>
          )
        })}

        {period.envelopes.length === 0 && (
          <div className="card">
            <div className="empty-state">No budget envelopes yet — add one to start logging expenses.</div>
            <button type="button" className="btn-dashed" onClick={openAddEnvelope}>
              + Add envelope
            </button>
          </div>
        )}
      </div>

      <div className="expense-total-banner">
        <div>
          <div className="expense-total-label">{period.label} total spend</div>
          <div className="mono expense-total-value">{fmt(period.actual_total_idr)}</div>
          <div className="expense-total-note">
            of {fmt(period.committed_total_idr)} committed across envelopes
          </div>
        </div>
      </div>

      <NewPeriodModal
        open={showNewPeriod}
        onClose={() => setShowNewPeriod(false)}
        hasExistingPeriod={periods.length > 0}
        latestPeriodEndDate={period.end_date}
        onCreated={(id) => setSelectedId(id)}
      />
      <BudgetEnvelopeModal
        open={showEnvelopeModal}
        onClose={() => setShowEnvelopeModal(false)}
        periodId={period.id}
        editingEnvelope={editingEnvelope}
      />
      <FixedExpenseModal
        open={showExpenseModal}
        onClose={() => setShowExpenseModal(false)}
        periodId={period.id}
        envelopes={period.envelopes}
        editingExpense={editingExpense}
        defaultEnvelopeId={expenseModalEnvelopeId}
      />
    </div>
  )
}

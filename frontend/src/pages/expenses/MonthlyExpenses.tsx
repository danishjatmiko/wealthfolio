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
import { formatShortDate, formatTimestamp, parseLocalDate, sourceLabel } from '../../lib/format'
import type { BudgetEnvelopeDetail, FixedExpense } from '../../types'
import './MonthlyExpenses.css'

const MS_PER_DAY = 86_400_000

/** How far today sits through a period: which day of how many, as a
 * percentage. Both ends are inclusive, so a 25 Jul – 24 Aug period is 31
 * days and its final day reads 100%. A period that hasn't started yet
 * reports day 0, and one already finished caps at its last day, so the bar
 * never runs past its track either way. */
function periodProgress(startDate: string, endDate: string) {
  const start = parseLocalDate(startDate)
  const end = parseLocalDate(endDate)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  const totalDays = Math.round((end.getTime() - start.getTime()) / MS_PER_DAY) + 1
  const elapsed = Math.round((today.getTime() - start.getTime()) / MS_PER_DAY) + 1
  const dayOfPeriod = Math.min(Math.max(elapsed, 0), totalDays)

  return { dayOfPeriod, totalDays, pct: totalDays > 0 ? (dayOfPeriod / totalDays) * 100 : 0 }
}

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

  const dayProgress = periodProgress(period.start_date, period.end_date)

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
            <span className="expense-period-day">
              Day {dayProgress.dayOfPeriod} of {dayProgress.totalDays} · {Math.round(dayProgress.pct)}
              % elapsed
            </span>
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
          // Spending faster than the period is running out. Only meaningful
          // once something's actually been spent — an untouched envelope is
          // trivially "on track" however far into the period we are.
          const aheadOfPace = env.actual_total_idr > 0 && pct > dayProgress.pct

          const spendPct = Math.min(100, pct)
          const pacePct = Math.min(100, dayProgress.pct)
          const paceTitle = `Day ${dayProgress.dayOfPeriod} of ${dayProgress.totalDays} — ${Math.round(dayProgress.pct)}% of the period gone`

          return (
            <div className="card" key={env.id}>
              <div className="expense-card-head" onClick={() => openEditEnvelope(env)}>
                <div>
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
              {/* The marker crosses the whole bar and pokes out top and
                  bottom so it reads as a ruler mark against the fill rather
                  than as part of it: fill short of the mark is on track,
                  past it is spending faster than the calendar. */}
              <div className="expense-bar">
                <div className="progress-track">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${spendPct}%`,
                      background: overBudget ? 'var(--red)' : 'var(--green)',
                    }}
                  />
                </div>
                <span className="expense-pace-marker" style={{ left: `${pacePct}%` }} title={paceTitle} />
              </div>
              <div className="expense-pace-meta">
                <span>{Math.round(pct)}% spent</span>
                <span className={aheadOfPace ? 'expense-pace-ahead' : undefined}>
                  {aheadOfPace ? 'ahead of pace' : 'on track'}
                </span>
              </div>

              {env.fixed_expenses.length === 0 && <div className="empty-state">No expenses logged yet.</div>}
              {/* A busy envelope can run to dozens of rows, which would leave
                  every card in the row stretched to match it. Scrolling the
                  list keeps the cards a readable height, with the total and
                  the add button staying put outside it. */}
              {env.fixed_expenses.length > 0 && (
                <div className="expense-rows">
                  {env.fixed_expenses.map((fe) => {
                    const edited = fe.updated_at !== fe.created_at
                    return (
                      <div className="expense-row" key={fe.id} onClick={() => openEditExpense(fe)}>
                        <div className="expense-row-info">
                          <div className="expense-row-name">
                            {fe.name}
                            {fe.source && <span className="source-badge">{sourceLabel(fe.source)}</span>}
                          </div>
                          <div className="expense-row-date">
                            {edited ? 'Updated' : 'Added'}{' '}
                            {formatTimestamp(edited ? fe.updated_at : fe.created_at)}
                          </div>
                          {fe.notes && <div className="expense-row-notes">{fe.notes}</div>}
                        </div>
                        <span className="mono expense-row-val">{fmt(fe.amount_idr)}</span>
                      </div>
                    )
                  })}
                </div>
              )}
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

import { useEffect, useRef, useState } from 'react'
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
import { periodProgress } from '../../lib/budget'
import { formatShortDate, sourceLabel } from '../../lib/format'
import type { BudgetEnvelopeDetail, FixedExpense } from '../../types'
import './MonthlyExpenses.css'

function amountBound(input: string): number | null {
  const value = input.trim()
  if (!value) return null
  if (!/^(?:\d+|\d{1,3}(?:,\d{3})+)$/.test(value)) return NaN
  const amount = Number(value.replace(/,/g, ''))
  return Number.isSafeInteger(amount) ? amount : NaN
}

export function MonthlyExpenses() {
  const { fmt, fmtExact } = useMoney()
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
  const [minAmount, setMinAmount] = useState('')
  const [maxAmount, setMaxAmount] = useState('')
  const [showAmountFilter, setShowAmountFilter] = useState(false)
  const filterToggleRef = useRef<HTMLButtonElement>(null)

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
  const minimum = amountBound(minAmount)
  const maximum = amountBound(maxAmount)
  const filtering = minimum !== null || maximum !== null
  const filterError = Number.isNaN(minimum) || Number.isNaN(maximum)
    ? 'Enter non-negative whole Rupiah amounts.'
    : minimum !== null && maximum !== null && minimum > maximum
      ? 'Minimum amount must be less than or equal to maximum amount.'
      : ''
  const amountRangeLabel = filterError ? 'Invalid amount range'
    : minimum !== null && maximum !== null ? `${fmtExact(minimum)} – ${fmtExact(maximum)}`
      : minimum !== null ? `From ${fmtExact(minimum)}`
        : maximum !== null ? `Up to ${fmtExact(maximum)}` : ''
  const visibleEnvelopes = period.envelopes.map((env) => ({
    env,
    expenses: env.fixed_expenses.filter((expense) => !filterError
      && (minimum === null || expense.amount_idr >= minimum)
      && (maximum === null || expense.amount_idr <= maximum)),
  })).filter(({ expenses }) => !filtering || expenses.length > 0)
  const totalCount = period.envelopes.reduce((sum, env) => sum + env.fixed_expenses.length, 0)
  const matchingCount = visibleEnvelopes.reduce((sum, { expenses }) => sum + expenses.length, 0)
  const matchingTotal = visibleEnvelopes.reduce((sum, { expenses }) =>
    sum + expenses.reduce((subtotal, expense) => subtotal + expense.amount_idr, 0), 0)

  function clearAmountFilter() {
    setMinAmount('')
    setMaxAmount('')
  }

  function closeAmountFilter() {
    setShowAmountFilter(false)
    filterToggleRef.current?.focus()
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
            <span className="expense-period-day">
              Day {dayProgress.dayOfPeriod} of {dayProgress.totalDays} · {Math.round(dayProgress.pct)}
              % elapsed
            </span>
          </div>
        </div>
        <div className="btn-group">
          <button type="button"
            ref={filterToggleRef}
            className={`btn expense-filter-toggle ${filtering ? 'expense-filter-toggle-active' : 'btn-secondary'}`}
            aria-expanded={showAmountFilter} aria-controls="expense-amount-panel"
            onClick={() => setShowAmountFilter((shown) => !shown)}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
              <path d="M3 4h18l-7 8v7l-4 2v-9L3 4Z" strokeLinejoin="round" />
            </svg>
            Filter amount
            {filtering && <span className="expense-filter-count">1</span>}
            <span aria-hidden="true">{showAmountFilter ? '▴' : '▾'}</span>
          </button>
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

      {!showAmountFilter && filtering && (
        <div className="expense-active-filter" aria-live="polite">
          <div className={`expense-filter-chip${filterError ? ' expense-filter-chip-error' : ''}`}>
            <button type="button" className="expense-filter-chip-label" onClick={() => setShowAmountFilter(true)}
              aria-label={`Edit amount filter: ${amountRangeLabel}`}>
              Amount: {amountRangeLabel}
            </button>
            <button type="button" className="expense-filter-chip-clear" onClick={clearAmountFilter} aria-label="Clear amount filter">×</button>
          </div>
          {!filterError && <span className="expense-filter-status">{matchingCount} of {totalCount} expenses</span>}
        </div>
      )}

      {showAmountFilter && <div id="expense-amount-panel" className="card expense-amount-filter" role="search" aria-label="Filter expenses by amount"
        onKeyDown={(e) => { if (e.key === 'Escape') { e.stopPropagation(); closeAmountFilter() } }}>
        <div className="expense-filter-head">
          <div>
            <div className="expense-filter-title">Amount range</div>
            <div className="expense-filter-hint" id="expense-amount-hint">
              Applies to every envelope in this period. Both limits are inclusive; leave either blank for no limit.
            </div>
          </div>
          <div className="expense-filter-panel-actions">
            {filtering && <button type="button" className="btn btn-secondary" onClick={clearAmountFilter}>Clear filter</button>}
            <button type="button" className="expense-filter-close" onClick={closeAmountFilter} aria-label="Close amount filter">×</button>
          </div>
        </div>
        <div className="expense-filter-fields">
          <label className="field">Minimum amount (Rp)
            <input autoFocus className="field-input mono" inputMode="numeric" value={minAmount}
              onChange={(e) => setMinAmount(e.target.value)} placeholder="No minimum"
              aria-invalid={!!filterError} aria-describedby="expense-amount-hint expense-filter-status" />
          </label>
          <label className="field">Maximum amount (Rp)
            <input className="field-input mono" inputMode="numeric" value={maxAmount}
              onChange={(e) => setMaxAmount(e.target.value)} placeholder="No maximum"
              aria-invalid={!!filterError} aria-describedby="expense-amount-hint expense-filter-status" />
          </label>
        </div>
        <div id="expense-filter-status" aria-live="polite" className={filterError ? 'error-text' : 'expense-filter-status'}>
          {filterError || (filtering
            ? `${matchingCount} of ${totalCount} expenses · Matching total ${fmtExact(matchingTotal)}`
            : `Showing all ${totalCount} expenses`)}
        </div>
      </div>}

      <div className="expense-grid">
        {visibleEnvelopes.map(({ env, expenses }) => {
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
                <div className="mono expense-card-total">{filtering
                  ? fmtExact(expenses.reduce((sum, expense) => sum + expense.amount_idr, 0))
                  : fmt(env.actual_total_idr)}</div>
              </div>
              {filtering ? <div className="expense-envelope-meta">
                {expenses.length} of {env.fixed_expenses.length} expenses · Matching total
              </div> : <>
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
              </>}

              {expenses.length === 0 && <div className="empty-state">No expenses logged yet.</div>}
              {/* A busy envelope can run to dozens of rows, which would leave
                  every card in the row stretched to match it. Scrolling the
                  list keeps the cards a readable height, with the total and
                  the add button staying put outside it. */}
              {expenses.length > 0 && (
                <div className="expense-rows">
                  {expenses.map((fe) => {
                    return (
                      <div className="expense-row" key={fe.id} onClick={() => openEditExpense(fe)}>
                        <div className="expense-row-info">
                          <div className="expense-row-name">
                            {fe.name}
                            {fe.source && <span className="source-badge">{sourceLabel(fe.source)}</span>}
                          </div>
                          <div className="expense-row-date">{formatShortDate(fe.expense_date)}</div>
                          {fe.notes && <div className="expense-row-notes">{fe.notes}</div>}
                        </div>
                        <span className="mono expense-row-val">{fmtExact(fe.amount_idr)}</span>
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

        {filtering && visibleEnvelopes.length === 0 && (
          <div className="card">
            <div className="empty-state">{filterError ? 'Adjust the amount range to see expenses.' : 'No expenses match this amount range.'}</div>
            <button type="button" className="btn btn-secondary" onClick={clearAmountFilter}>Clear filter</button>
          </div>
        )}

        {!filtering && period.envelopes.length === 0 && (
          <div className="card">
            <div className="empty-state">No budget envelopes yet — add one to start logging expenses.</div>
            <button type="button" className="btn-dashed" onClick={openAddEnvelope}>
              + Add envelope
            </button>
          </div>
        )}
      </div>

      {!filterError && <div className="expense-total-banner">
        <div>
          <div className="expense-total-label">{period.label} {filtering ? 'matching spend' : 'total spend'}</div>
          <div className="mono expense-total-value">{filtering ? fmtExact(matchingTotal) : fmt(period.actual_total_idr)}</div>
          <div className="expense-total-note">
            {filtering
              ? `${matchingCount} of ${totalCount} expenses · ${fmtExact(period.actual_total_idr)} total period spend`
              : `of ${fmt(period.committed_total_idr)} committed across envelopes`}
          </div>
        </div>
      </div>}

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

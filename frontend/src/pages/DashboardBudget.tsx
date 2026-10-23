import {
  BUDGET_STATUS_LABEL,
  budgetStatus,
  budgetUsedPct,
  periodProgress,
  type BudgetStatus,
} from '../lib/budget'
import type { DashboardExpense, ExpensePeriodSummary } from '../types'

type Fmt = (value: number) => string

function StatusChip({ status, soFar }: { status: BudgetStatus | null; soFar?: boolean }) {
  if (!status) return <span className="budget-chip budget-chip-none">No budget</span>
  return (
    <span className={`budget-chip budget-chip-${status}`}>
      {BUDGET_STATUS_LABEL[status]}
      {soFar && status !== 'over' ? ' so far' : ''}
    </span>
  )
}

/** "Rp1.2 mn under" / "Rp300 rb over" — how far spend sits from its target. */
function DeltaText({ actual, committed, fmt }: { actual: number; committed: number; fmt: Fmt }) {
  if (actual === committed && committed > 0) return <span className="expense-under">right on budget</span>
  const over = actual > committed
  return (
    <span className={over ? 'expense-over' : 'expense-under'}>
      {fmt(Math.abs(actual - committed))} {over ? 'over' : 'under'}
    </span>
  )
}

/** The selected period's spend against its targets: one overall verdict,
 *  how many envelopes landed in each bucket, then every envelope. */
export function BudgetVsActualCard({ expense, fmt }: { expense: DashboardExpense; fmt: Fmt }) {
  const progress =
    expense.start_date && expense.end_date ? periodProgress(expense.start_date, expense.end_date) : null
  const inProgress = !!progress?.isCurrent
  const overall = budgetStatus(expense.actual_total_idr, expense.committed_total_idr)
  const usedPct = budgetUsedPct(expense.actual_total_idr, expense.committed_total_idr)

  const counts: Record<BudgetStatus, number> = { over: 0, on_target: 0, under: 0 }
  for (const env of expense.envelopes) {
    const s = budgetStatus(env.actual_idr, env.committed_idr)
    if (s) counts[s]++
  }

  return (
    <div className="card">
      <div className="budget-card-head">
        <div className="card-title">Budget vs actual · {expense.period_label || '—'}</div>
        {expense.period_id && <StatusChip status={overall} soFar={inProgress} />}
      </div>

      {!expense.period_id && <div className="empty-state">No expense periods yet.</div>}

      {expense.period_id && (
        <>
          <div className="budget-overall">
            <div className="mono budget-overall-spent">{fmt(expense.actual_total_idr)}</div>
            <div className="budget-overall-of">
              of {fmt(expense.committed_total_idr)} committed · {Math.round(usedPct)}% used ·{' '}
              <DeltaText actual={expense.actual_total_idr} committed={expense.committed_total_idr} fmt={fmt} />
            </div>
            {inProgress && progress && (
              <div className="budget-overall-note">
                Day {progress.dayOfPeriod} of {progress.totalDays} — still in progress, so "under" may
                fill up before the period ends.
              </div>
            )}
          </div>

          <div className="budget-counts">
            <span className="budget-count budget-count-over">{counts.over} over</span>
            <span className="budget-count budget-count-on_target">{counts.on_target} on target</span>
            <span className="budget-count budget-count-under">{counts.under} under</span>
          </div>

          {expense.envelopes.length === 0 && <div className="empty-state">No budget envelopes yet.</div>}
          {expense.envelopes.map((env) => {
            const status = budgetStatus(env.actual_idr, env.committed_idr)
            const pct = budgetUsedPct(env.actual_idr, env.committed_idr)
            return (
              <div className="budget-row" key={env.envelope_id}>
                <div className="budget-row-top">
                  <span className="dash-cat-swatch" style={{ background: env.color_oklch }} />
                  <span className="budget-row-name">{env.label}</span>
                  <span className="mono budget-row-figures">
                    {fmt(env.actual_idr)} <span className="budget-row-of">/ {fmt(env.committed_idr)}</span>
                  </span>
                  <StatusChip status={status} soFar={inProgress} />
                </div>
                <div className="progress-track budget-row-track">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${Math.min(100, env.committed_idr > 0 ? pct : env.actual_idr > 0 ? 100 : 0)}%`,
                      background: status === 'over' ? 'var(--red)' : 'var(--green)',
                    }}
                  />
                </div>
                <div className="budget-row-meta">
                  <span>{env.committed_idr > 0 ? `${Math.round(pct)}% used` : 'nothing committed'}</span>
                  <DeltaText actual={env.actual_idr} committed={env.committed_idr} fmt={fmt} />
                </div>
              </div>
            )
          })}
        </>
      )}
    </div>
  )
}

/** Every period's overall verdict at a glance, newest first. Clicking one
 *  selects it for the rest of the expense section. */
export function PeriodHistoryCard({
  periods,
  selectedId,
  onSelect,
  fmt,
}: {
  periods: ExpensePeriodSummary[]
  selectedId: string | undefined
  onSelect: (id: string) => void
  fmt: Fmt
}) {
  const counts: Record<BudgetStatus, number> = { over: 0, on_target: 0, under: 0 }
  for (const p of periods) {
    const s = budgetStatus(p.actual_total_idr, p.committed_total_idr)
    if (s) counts[s]++
  }

  return (
    <div className="card">
      <div className="card-title">All periods</div>
      {periods.length === 0 && <div className="empty-state">No expense periods yet.</div>}
      {periods.length > 0 && (
        <div className="budget-counts">
          <span className="budget-count budget-count-over">{counts.over} over</span>
          <span className="budget-count budget-count-on_target">{counts.on_target} on target</span>
          <span className="budget-count budget-count-under">{counts.under} under</span>
        </div>
      )}
      <div className="budget-history">
        {periods.map((p) => {
          const status = budgetStatus(p.actual_total_idr, p.committed_total_idr)
          const pct = budgetUsedPct(p.actual_total_idr, p.committed_total_idr)
          const inProgress = periodProgress(p.start_date, p.end_date).isCurrent
          return (
            <button
              type="button"
              key={p.id}
              className={'budget-history-row' + (p.id === selectedId ? ' budget-history-row-active' : '')}
              onClick={() => onSelect(p.id)}
            >
              <span className="budget-history-top">
                <span className="budget-history-label">{p.label}</span>
                <StatusChip status={status} soFar={inProgress} />
              </span>
              <span className="budget-history-bottom">
                <span className="mono budget-history-figures">
                  {fmt(p.actual_total_idr)} <span className="budget-row-of">/ {fmt(p.committed_total_idr)}</span>
                </span>
                <span className="mono budget-history-pct">
                  {p.committed_total_idr > 0 ? `${Math.round(pct)}% used` : '—'}
                </span>
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}

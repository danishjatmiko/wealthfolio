// Budget-vs-actual judgements shared by the Home page and the Expenses page,
// so both read an envelope or a whole period the same way.
import { parseLocalDate } from './format'

export type BudgetStatus = 'over' | 'on_target' | 'under'

/** Spending at least this share of a budget, without passing it, counts as
 *  meeting the target. Anything past 100% is over — the same line the
 *  Expenses page draws for "over". */
export const ON_TARGET_FROM_PCT = 90

export const BUDGET_STATUS_LABEL: Record<BudgetStatus, string> = {
  over: 'Over',
  on_target: 'On target',
  under: 'Under',
}

/** Percent of the committed amount actually spent; 0 when nothing was
 *  committed, so an unbudgeted envelope never divides by zero. */
export function budgetUsedPct(actualIdr: number, committedIdr: number): number {
  return committedIdr > 0 ? (actualIdr / committedIdr) * 100 : 0
}

/** Judges spend against its target. null means there's nothing to judge:
 *  no budget and no spend. Spend with no budget at all is over. */
export function budgetStatus(actualIdr: number, committedIdr: number): BudgetStatus | null {
  if (committedIdr <= 0) return actualIdr > 0 ? 'over' : null
  if (actualIdr > committedIdr) return 'over'
  if (budgetUsedPct(actualIdr, committedIdr) >= ON_TARGET_FROM_PCT) return 'on_target'
  return 'under'
}

const MS_PER_DAY = 86_400_000

/** How far today sits through a period: which day of how many, as a
 * percentage. Both ends are inclusive, so a 25 Jul – 24 Aug period is 31
 * days and its final day reads 100%. A period that hasn't started yet
 * reports day 0, and one already finished caps at its last day, so the bar
 * never runs past its track either way. */
export function periodProgress(startDate: string, endDate: string) {
  const start = parseLocalDate(startDate)
  const end = parseLocalDate(endDate)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  const totalDays = Math.round((end.getTime() - start.getTime()) / MS_PER_DAY) + 1
  const elapsed = Math.round((today.getTime() - start.getTime()) / MS_PER_DAY) + 1
  const dayOfPeriod = Math.min(Math.max(elapsed, 0), totalDays)
  // Today falls inside the period, so its figures are still moving.
  const isCurrent = elapsed >= 1 && elapsed <= totalDays

  return { dayOfPeriod, totalDays, isCurrent, pct: totalDays > 0 ? (dayOfPeriod / totalDays) * 100 : 0 }
}

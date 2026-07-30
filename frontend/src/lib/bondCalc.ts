// Client mirror of backend/internal/service/bondcalc.go, used only to
// preview the derived figures live while typing in the purchase modal. The
// backend recomputes everything authoritatively on save — same relationship
// holdingCalc.ts has with valuation.go. Keep the two in step.

export const BOND_PLATFORMS = ['OCBC', 'MyBCA', 'Mandiri'] as const

const COUPONS_PER_YEAR = 2

/** Settlement amount: the clean lot price plus accrued interest owed to the seller. */
export function bondTotalUsd(priceUsd: number, accruedUsd: number): number {
  return priceUsd + accruedUsd
}

/** Rupiah cost at the rate actually paid on the purchase date. */
export function bondTotalIdr(totalUsd: number, usdIdr: number): number {
  return Math.round(totalUsd * usdIdr)
}

/** One payment for the whole lot. */
export function couponPerCycleUsd(
  faceValueUsd: number,
  quantity: number,
  interestRatePct: number,
): number {
  return (faceValueUsd * quantity * (interestRatePct / 100)) / COUPONS_PER_YEAR
}

/**
 * The two months a bond pays in, ascending — the maturity month and the one
 * six months away. Returns [] for an unparseable date so the preview just
 * stays blank while the user is mid-type.
 */
export function couponMonths(maturityDate: string): number[] {
  const [, m] = maturityDate.split('-').map(Number)
  if (!m || m < 1 || m > 12) return []
  const other = (((m - 1 + 6) % 12) + 1)
  return [m, other].sort((a, b) => a - b)
}

const MONTH_NAMES = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
]

/** Human-readable pay schedule, e.g. "2 Jul, 2 Jan" — the day is the
 *  maturity day, shown unclamped since this is just a hint. */
export function couponScheduleLabel(maturityDate: string): string {
  const [, , d] = maturityDate.split('-').map(Number)
  const months = couponMonths(maturityDate)
  if (!d || months.length === 0) return ''
  return months.map((m) => `${d} ${MONTH_NAMES[m - 1]}`).join(', ')
}

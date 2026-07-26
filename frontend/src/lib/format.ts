// Money-unit convention: every monetary field from the API is an integer
// representing full/raw IDR (whole Rupiah) — no rounding, no scaling
// factor. `fmtIdr`/`money`/`goldFmt` are abbreviated, summary-only
// formatters (Dashboard + the persistent sidebar total); every other
// display should use `fmtIdrExact`/`moneyExact` instead.

/**
 * Format an IDR amount into the shortened form used on the Dashboard and
 * the persistent sidebar total (e.g. "Rp3.75 B", "Rp202.00 mn", "Rp16.80
 * mn", "Rp800 rb"). Lossy by design — do not use for detail views.
 */
export function fmtIdr(value: number): string {
  const neg = value < 0
  const v = Math.abs(value)
  let s: string
  if (v >= 1_000_000_000) s = 'Rp' + (v / 1_000_000_000).toFixed(2) + ' B'
  else if (v >= 1_000_000) s = 'Rp' + (v / 1_000_000).toFixed(2) + ' mn'
  else s = 'Rp' + Math.round(v / 1000) + ' rb'
  return (neg ? '−' : '') + s
}

/**
 * Hide-aware abbreviated money formatter (Dashboard/sidebar only). Pass the
 * current "hide values" state; when hidden, every figure collapses to
 * "Rp ••••" per the design spec.
 */
export function money(value: number, hidden: boolean): string {
  if (hidden) return 'Rp ••••'
  return fmtIdr(value)
}

/**
 * Exact IDR amount, no abbreviation/rounding — e.g. "Rp1,903,111". Use this
 * everywhere except the Dashboard and the persistent sidebar total.
 */
export function fmtIdrExact(value: number): string {
  const neg = value < 0
  const v = Math.round(Math.abs(value))
  return (neg ? '−' : '') + 'Rp' + v.toLocaleString('en-US')
}

/**
 * Hide-aware exact money formatter. Pass the current "hide values" state;
 * when hidden, collapses to "Rp ••••" like `money()`.
 */
export function moneyExact(value: number, hidden: boolean): string {
  if (hidden) return 'Rp ••••'
  return fmtIdrExact(value)
}

/**
 * Gold price per gram, exact — e.g. "Rp1,903,111/g".
 */
export function goldFmt(value: number): string {
  return 'Rp' + Math.round(Math.abs(value)).toLocaleString('en-US') + '/g'
}

/**
 * USD -> IDR rate formatter. Input is full IDR per 1 USD, exactly like the
 * prototype's usdFmt(). Output e.g. "Rp18,100".
 */
export function usdFmt(value: number): string {
  return 'Rp' + value.toLocaleString('en-US')
}

/** Formats a "YYYY-MM-DD" date string as e.g. "20 Jul 2026". */
export function formatShortDate(dateStr: string): string {
  const [y, m, d] = dateStr.split('-').map(Number)
  return new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }).format(
    new Date(y, m - 1, d),
  )
}

/** Formats an ISO 8601 timestamp (e.g. created_at/updated_at) as e.g. "20 Jul 2026". */
export function formatTimestamp(isoStr: string): string {
  return new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }).format(
    new Date(isoStr),
  )
}

/** Parse a free-typed numeric string the same way the prototype does. */
export function parseNumeric(input: string | number | null | undefined): number {
  if (typeof input === 'number') return Number.isFinite(input) ? input : 0
  const cleaned = (input ?? '').toString().replace(/[^0-9.]/g, '')
  const n = parseFloat(cleaned)
  return Number.isFinite(n) ? n : 0
}

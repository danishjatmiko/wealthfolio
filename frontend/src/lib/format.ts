// Ported verbatim (same formulas/thresholds) from the design prototype's
// money()/goldFmt()/usdFmt() (Portfolio App.dc.html, lines ~662-680).
//
// Money-unit convention: every monetary field from the API is an integer
// representing THOUSANDS of IDR, except RateEntry.usd_idr which is full IDR
// per 1 USD (use usdFmt for that one field only).

/**
 * Format an IDR amount given in THOUSANDS of rupiah into the shortened form
 * used throughout the app (e.g. "Rp3.75 B", "Rp202.00 mn", "Rp16.80 mn",
 * "Rp800 rb"). Mirrors the prototype's `money()`/`fmtIdr()`.
 */
export function fmtIdr(value: number): string {
  const neg = value < 0
  const v = Math.abs(value)
  const j = v / 1000
  let s: string
  if (j >= 1000) s = 'Rp' + (j / 1000).toFixed(2) + ' B'
  else if (j >= 1) s = 'Rp' + j.toFixed(2) + ' mn'
  else s = 'Rp' + Math.round(v) + ' rb'
  return (neg ? '−' : '') + s
}

/**
 * Hide-aware money formatter. Pass the current "hide values" state; when
 * hidden, every figure collapses to "Rp ••••" per the design spec.
 */
export function money(value: number, hidden: boolean): string {
  if (hidden) return 'Rp ••••'
  return fmtIdr(value)
}

/**
 * Gold price per gram. Input is in THOUSANDS of IDR (same unit as all other
 * money fields except usd_idr). Output e.g. "Rp2.65 mn/g".
 */
export function goldFmt(value: number): string {
  return 'Rp' + (value / 1000).toFixed(2) + ' mn/g'
}

/**
 * USD -> IDR rate formatter. Input is FULL IDR per 1 USD (not thousands),
 * exactly like the prototype's usdFmt(). Output e.g. "Rp18,100".
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

/** Parse a free-typed numeric string the same way the prototype does. */
export function parseNumeric(input: string | number | null | undefined): number {
  if (typeof input === 'number') return Number.isFinite(input) ? input : 0
  const cleaned = (input ?? '').toString().replace(/[^0-9.]/g, '')
  const n = parseFloat(cleaned)
  return Number.isFinite(n) ? n : 0
}

// Ported from the prototype's goldPrice()/mdComputedVal() (Portfolio App.dc.html
// lines 719-720). Category behavior is keyed off the fixed, seeded category
// labels rather than free-form ids, matching the backend's fixed category set.
import { parseNumeric } from './format'
import type { RateEntry } from '../types'

export const GOLD_CATEGORY_LABEL = 'Logam Mulia'
export const BONDS_GOV_CATEGORY_LABEL = 'Bonds Gov.'
export const BONDS_CORP_CATEGORY_LABEL = 'Bonds Corp.'
export const US_ETF_CATEGORY_LABEL = 'US ETF / Stocks'
export const IDR_CASH_CATEGORY_LABEL = 'IDR Cash'
export const USD_CASH_CATEGORY_LABEL = 'USD Cash'

export const GOLD_TYPES = ['Antam', 'King Halim', 'UBS'] as const
export type GoldType = (typeof GOLD_TYPES)[number]

/** Latest gold price per gram (in full/raw IDR) for the given brand. */
export function goldPrice(rate: RateEntry | undefined | null, brand: string): number {
  if (!rate) return 0
  if (brand === 'King Halim') return rate.kinghalim
  if (brand === 'UBS') return rate.ubs
  return rate.antam
}

export interface AssetFormValues {
  categoryLabel: string
  val: string // direct IDR (full/raw), as typed
  gram: string
  qty: string
  usd: string
  brand: string
}

/**
 * Live client-side preview of the IDR value (full/raw) for the asset
 * add/edit form, mirroring mdComputedVal() exactly.
 */
/** Categories entered in USD and valued at the latest USD/IDR rate. */
function isFixedUsdCategory(categoryLabel: string): boolean {
  return (
    categoryLabel === BONDS_GOV_CATEGORY_LABEL ||
    categoryLabel === BONDS_CORP_CATEGORY_LABEL ||
    categoryLabel === US_ETF_CATEGORY_LABEL ||
    categoryLabel === USD_CASH_CATEGORY_LABEL
  )
}

export function computeHoldingValue(md: AssetFormValues, latestRate: RateEntry | undefined | null): number {
  if (md.categoryLabel === GOLD_CATEGORY_LABEL) {
    const g = parseNumeric(md.gram)
    return g ? g * (parseNumeric(md.qty) || 1) * goldPrice(latestRate, md.brand) : parseNumeric(md.val)
  }
  if (isFixedUsdCategory(md.categoryLabel)) {
    const u = parseNumeric(md.usd)
    return u && latestRate ? u * latestRate.usd_idr : parseNumeric(md.val)
  }
  return parseNumeric(md.val)
}

/** Whether the form should show the Gram/Qty/Type gold fields. */
export function isGoldCategory(categoryLabel: string): boolean {
  return categoryLabel === GOLD_CATEGORY_LABEL
}

/** The currency a cash holding is stored with — each cash category holds
 *  exactly one — or null for anything that isn't cash. */
export function cashCurrency(categoryLabel: string): 'IDR' | 'USD' | null {
  if (categoryLabel === IDR_CASH_CATEGORY_LABEL) return 'IDR'
  if (categoryLabel === USD_CASH_CATEGORY_LABEL) return 'USD'
  return null
}

/** Whether the form should show a USD input (bonds, US ETF / Stocks, USD cash). */
export function showsUsdInput(categoryLabel: string): boolean {
  return isFixedUsdCategory(categoryLabel)
}

/** Whether the form should show a direct IDR input. */
export function showsIdrInput(categoryLabel: string): boolean {
  return !isGoldCategory(categoryLabel) && !showsUsdInput(categoryLabel)
}

/** Whether the form should show the read-only auto-computed IDR box. */
export function showsComputedBox(categoryLabel: string): boolean {
  return isGoldCategory(categoryLabel) || showsUsdInput(categoryLabel)
}

/** Builds the `detail` string exactly like the prototype's saveAsset(). */
export function buildDetail(md: AssetFormValues): string {
  if (md.categoryLabel === GOLD_CATEGORY_LABEL) {
    const g = parseNumeric(md.gram)
    const q = parseNumeric(md.qty) || 1
    return q > 1 ? `${q} × ${g} g` : `${g} g`
  }
  if (isFixedUsdCategory(md.categoryLabel)) {
    return `${parseNumeric(md.usd)} USD`
  }
  return ''
}

/**
 * Prefill helper for editing an existing holding that may not have
 * structured gram/qty/usd stored (parses from `detail`, per the design
 * spec's fallback behavior) but prefers structured fields when present.
 */
export function prefillFromHolding(h: {
  category_label: string
  detail: string | null
  gram: number | null
  qty: number | null
  usd_value: number | null
  brand: string | null
  name: string
  value_idr: number
}): AssetFormValues {
  const d = h.detail || ''
  let gram = h.gram != null ? String(h.gram) : ''
  let qty = h.qty != null ? String(h.qty) : ''
  let usd = h.usd_value != null ? String(h.usd_value) : ''
  const brand =
    h.brand || (/king/i.test(h.name) ? 'King Halim' : /ubs/i.test(h.name) ? 'UBS' : 'Antam')

  if (h.category_label === GOLD_CATEGORY_LABEL && !gram) {
    const mm = d.match(/(?:([\d.]+)\s*[×x]\s*)?([\d.]+)\s*g/i)
    if (mm) {
      qty = mm[1] || '1'
      gram = mm[2] || ''
    }
  }
  if (isFixedUsdCategory(h.category_label) && !usd && /usd/i.test(d)) {
    const um = d.match(/([\d,.]+)\s*usd/i)
    if (um) {
      usd = um[1].replace(/,/g, '')
    }
  }

  return {
    categoryLabel: h.category_label,
    val: String(h.value_idr),
    gram,
    qty,
    usd,
    brand,
  }
}

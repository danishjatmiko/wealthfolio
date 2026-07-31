import { useMemo, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useBondSummary } from '../../hooks/useBondPurchases'
import { fmtUsdIdrRate, formatShortDate } from '../../lib/format'
import { BondPurchaseModal } from './BondPurchaseModal'
import type { BondNameSummary, BondPurchase } from '../../types'
import './Bonds.css'

type SortKey = 'bond_name' | 'interest_rate' | 'ytm_pct' | 'maturity_date' | 'quantity' | 'total_usd'

const COLUMNS: { key: SortKey; label: string; numeric?: boolean; title?: string }[] = [
  { key: 'bond_name', label: 'Bond' },
  { key: 'interest_rate', label: 'Coupon', numeric: true, title: 'Annual coupon rate' },
  {
    key: 'ytm_pct',
    label: 'YTM',
    numeric: true,
    title: 'Yield to maturity — the annual return if held to redemption, accounting for the price paid',
  },
  { key: 'maturity_date', label: 'Matures' },
  { key: 'quantity', label: 'Qty', numeric: true },
  { key: 'total_usd', label: 'Invested', numeric: true, title: 'Clean price, excluding accrued interest' },
]

export function Bonds() {
  const { fmtExact, fmtUsd } = useMoney()
  const { data: summary, isLoading } = useBondSummary()

  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [sortKey, setSortKey] = useState<SortKey>('total_usd')
  const [sortAsc, setSortAsc] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingPurchase, setEditingPurchase] = useState<BondPurchase | null>(null)
  const [defaultBondName, setDefaultBondName] = useState<string | undefined>(undefined)

  const bonds = useMemo(() => {
    const rows = [...(summary?.bonds ?? [])]
    rows.sort((a, b) => {
      const av = a[sortKey]
      const bv = b[sortKey]
      // bond_name and maturity_date are strings; everything else is numeric.
      // localeCompare keeps ISO dates in chronological order too.
      const cmp =
        typeof av === 'string' && typeof bv === 'string'
          ? av.localeCompare(bv)
          : Number(av) - Number(bv)
      return sortAsc ? cmp : -cmp
    })
    return rows
  }, [summary, sortKey, sortAsc])

  function toggleSort(key: SortKey) {
    if (key === sortKey) {
      setSortAsc((v) => !v)
    } else {
      setSortKey(key)
      // Names read best A–Z; figures read best biggest-first.
      setSortAsc(key === 'bond_name' || key === 'maturity_date')
    }
  }

  function toggleExpand(name: string) {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }

  function openAdd(bondName?: string) {
    setEditingPurchase(null)
    setDefaultBondName(bondName)
    setModalOpen(true)
  }

  function openEdit(purchase: BondPurchase) {
    setEditingPurchase(purchase)
    setDefaultBondName(undefined)
    setModalOpen(true)
  }

  function cell(b: BondNameSummary, key: SortKey) {
    switch (key) {
      case 'bond_name':
        return (
          <span className="bond-name" key={key}>
            {b.bond_name}
            {b.platforms.map((p) => (
              <span className="source-badge" key={p}>
                {p}
              </span>
            ))}
            {b.is_matured && <span className="source-badge bond-badge-matured">Matured</span>}
          </span>
        )
      case 'interest_rate':
        return (
          <span className="mono bond-num" key={key}>
            {b.interest_rate.toFixed(2)}%
          </span>
        )
      case 'ytm_pct':
        return (
          <span className="mono bond-num bond-ytm" key={key}>
            {b.ytm_pct > 0 ? `${b.ytm_pct.toFixed(2)}%` : '—'}
          </span>
        )
      case 'maturity_date':
        return (
          <span className="mono" key={key}>
            {formatShortDate(b.maturity_date)}
          </span>
        )
      case 'quantity':
        return (
          <span className="mono bond-num" key={key}>
            {b.quantity}
          </span>
        )
      case 'total_usd':
        return (
          <span className="mono bond-num" key={key}>
            {fmtUsd(b.total_usd)}
          </span>
        )
    }
  }

  // Every Rupiah figure on this page is a conversion at the latest logged
  // rate, so it's worth saying which rate — otherwise the numbers look
  // unexplained when the rate moves. Omitted entirely if none is logged.
  const latestRate = summary?.latest_usd_idr ?? 0
  const rateNote =
    latestRate > 0 ? (
      <span className="bond-summary-rate"> @ {fmtUsdIdrRate(latestRate)}</span>
    ) : null

  return (
    <div>
      <div className="row-wrap bonds-header">
        <div className="bonds-header-copy">
          Every USD bond purchase, rolled up per bond. Click a bond to see the individual buys behind
          it, or a column heading to sort.
        </div>
        <button type="button" className="btn btn-primary" onClick={() => openAdd()}>
          + Add purchase
        </button>
      </div>

      <div className="bonds-summary-grid">
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Invested</div>
          <div className="mono bond-summary-value">{fmtUsd(summary?.total_usd ?? 0)}</div>
          <div className="mono bond-summary-sub">
            {fmtExact(summary?.total_idr ?? 0)}
            {rateNote}
          </div>
        </div>
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Paid incl. accrued</div>
          <div className="mono bond-summary-value">{fmtUsd(summary?.settlement_usd ?? 0)}</div>
          <div className="mono bond-summary-sub">
            {fmtUsd((summary?.settlement_usd ?? 0) - (summary?.total_usd ?? 0))} accrued advanced
          </div>
        </div>
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Average IDR we buy</div>
          <div className="mono bond-summary-value">
            {summary && summary.average_usd_idr > 0
              ? fmtUsdIdrRate(Math.round(summary.average_usd_idr))
              : '—'}
          </div>
          <div className="bond-summary-sub">blended rate actually paid</div>
        </div>
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Coupons per year</div>
          <div className="mono bond-summary-value">
            {fmtUsd(summary?.coupon_per_year_usd ?? 0)}
            {/* YTM sits with the dollar figure because it's a return on the
                USD position; the Rupiah line is just a conversion of it. */}
            <span className="bond-summary-ytm">{(summary?.ytm_pct ?? 0).toFixed(2)}% YTM</span>
          </div>
          <div className="mono bond-summary-sub">
            {fmtExact(summary?.coupon_per_year_idr ?? 0)}
            {rateNote}
          </div>
        </div>
      </div>

      <div className="card bonds-table-card">
        <div className="bond-row bond-row-head">
          {COLUMNS.map((c) => (
            <span
              key={c.key}
              className={'bond-sort' + (c.numeric ? ' bond-num' : '')}
              title={c.title}
              onClick={() => toggleSort(c.key)}
            >
              {c.label}
              {sortKey === c.key && <span className="bond-sort-arrow">{sortAsc ? '▲' : '▼'}</span>}
            </span>
          ))}
          <span />
        </div>

        {isLoading && <div className="empty-state">Loading…</div>}
        {!isLoading && bonds.length === 0 && (
          <div className="empty-state">No bond purchases logged yet.</div>
        )}

        {bonds.map((b) => {
          const isOpen = expanded.has(b.bond_name)
          return (
            <div key={b.bond_name}>
              <div className="bond-row bond-name-row" onClick={() => toggleExpand(b.bond_name)}>
                {COLUMNS.map((c) => cell(b, c.key))}
                <span className="bond-chevron">{isOpen ? '▾' : '▸'}</span>
              </div>

              {b.has_conflicts && (
                <div className="error-text bond-conflict">
                  Purchases under this name disagree on rate or maturity — they may be two different
                  bonds.
                </div>
              )}

              {isOpen && (
                <div className="bond-purchases">
                  <div className="bond-purchase-row bond-purchase-head">
                    <span>Bought</span>
                    <span>Platform</span>
                    <span className="bond-num">Qty</span>
                    <span className="bond-num">Price</span>
                    <span className="bond-num">Accrued</span>
                    <span className="bond-num">Paid</span>
                    <span className="bond-num">YTM</span>
                    <span className="bond-num">Rate paid</span>
                    <span className="bond-num" title="Paid amount converted at your latest logged rate, not the rate paid that day">
                      IDR value
                    </span>
                    <span />
                  </div>
                  {b.purchases.map((p) => (
                    <div className="bond-purchase-row" key={p.id}>
                      <span className="mono">{formatShortDate(p.buy_date)}</span>
                      <span>{p.platform || '—'}</span>
                      <span className="mono bond-num">{p.quantity}</span>
                      <span className="mono bond-num">{fmtUsd(p.price_usd)}</span>
                      <span className="mono bond-num">{fmtUsd(p.accrued_interest_usd)}</span>
                      <span className="mono bond-num">{fmtUsd(p.settlement_usd)}</span>
                      <span className="mono bond-num bond-ytm">{p.ytm_pct.toFixed(2)}%</span>
                      <span className="mono bond-num">{fmtUsdIdrRate(p.usd_idr_at_purchase)}</span>
                      <span className="mono bond-num">{fmtExact(p.settlement_idr)}</span>
                      <button
                        type="button"
                        title="Edit"
                        className="a-edit-btn"
                        onClick={() => openEdit(p)}
                      >
                        ✎
                      </button>
                    </div>
                  ))}
                  <button type="button" className="btn-dashed" onClick={() => openAdd(b.bond_name)}>
                    + Add another {b.bond_name} purchase
                  </button>
                </div>
              )}
            </div>
          )
        })}

        {bonds.length > 0 && (
          <div className="bond-net-row">
            <span className="bond-net-label">Invested</span>
            <span className="mono bond-net-sub">
              paid {fmtUsd(summary?.settlement_usd ?? 0)} incl. accrued
            </span>
            <span className="mono bond-net-total">{fmtUsd(summary?.total_usd ?? 0)}</span>
          </div>
        )}
      </div>

      <p className="assets-footnote">
        Invested is the clean price; accrued interest is money advanced to the seller and returned at
        the first coupon, so it's shown separately. Rupiah figures convert at your latest logged rate,
        the same way the Assets page values every USD holding — the rate each lot was actually bought
        at is kept per purchase.
      </p>

      <BondPurchaseModal
        open={modalOpen}
        onClose={() => {
          setModalOpen(false)
          setEditingPurchase(null)
          setDefaultBondName(undefined)
        }}
        editingPurchase={editingPurchase}
        defaultBondName={defaultBondName}
      />
    </div>
  )
}

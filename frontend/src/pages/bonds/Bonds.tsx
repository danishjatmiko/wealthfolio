import { useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useBondSummary } from '../../hooks/useBondPurchases'
import { fmtUsdIdrRate, formatShortDate } from '../../lib/format'
import { BondPurchaseModal } from './BondPurchaseModal'
import type { BondPurchase } from '../../types'
import './Bonds.css'

export function Bonds() {
  const { fmtExact, fmtUsd } = useMoney()
  const { data: summary, isLoading } = useBondSummary()

  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [modalOpen, setModalOpen] = useState(false)
  const [editingPurchase, setEditingPurchase] = useState<BondPurchase | null>(null)
  const [defaultBondName, setDefaultBondName] = useState<string | undefined>(undefined)

  function toggle(name: string) {
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

  const bonds = summary?.bonds ?? []

  return (
    <div>
      <div className="row-wrap bonds-header">
        <div className="bonds-header-copy">
          Every USD bond purchase, rolled up per bond. Click a bond to see the individual buys behind
          it.
        </div>
        <button type="button" className="btn btn-primary" onClick={() => openAdd()}>
          + Add purchase
        </button>
      </div>

      <div className="bonds-summary-grid">
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Total invested</div>
          <div className="mono bond-summary-value">{fmtUsd(summary?.total_usd ?? 0)}</div>
          <div className="mono bond-summary-sub">{fmtExact(summary?.total_idr ?? 0)}</div>
        </div>
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Average IDR we buy</div>
          <div className="mono bond-summary-value">
            {summary && summary.average_usd_idr > 0 ? fmtUsdIdrRate(Math.round(summary.average_usd_idr)) : '—'}
          </div>
          <div className="bond-summary-sub">blended across every lot</div>
        </div>
        <div className="card bond-summary-card">
          <div className="bond-summary-label">Coupons per year</div>
          <div className="mono bond-summary-value">{fmtUsd(summary?.coupon_per_year_usd ?? 0)}</div>
          <div className="mono bond-summary-sub">{fmtExact(summary?.coupon_per_year_idr ?? 0)}</div>
        </div>
      </div>

      <div className="card bonds-table-card">
        <div className="bond-row bond-row-head">
          <span>Bond</span>
          <span>Rate</span>
          <span>Matures</span>
          <span className="bond-num">Qty</span>
          <span className="bond-num">Total</span>
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
              <div className="bond-row bond-name-row" onClick={() => toggle(b.bond_name)}>
                <span className="bond-name">
                  {b.bond_name}
                  {b.platforms.map((p) => (
                    <span className="source-badge" key={p}>
                      {p}
                    </span>
                  ))}
                  {b.is_matured && <span className="source-badge bond-badge-matured">Matured</span>}
                </span>
                <span className="mono">{b.interest_rate.toFixed(2)}%</span>
                <span className="mono">{formatShortDate(b.maturity_date)}</span>
                <span className="mono bond-num">{b.quantity}</span>
                <span className="mono bond-num">{fmtUsd(b.total_usd)}</span>
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
                    <span className="bond-num">Total</span>
                    <span className="bond-num">Rate</span>
                    <span className="bond-num">Total IDR</span>
                    <span />
                  </div>
                  {b.purchases.map((p) => (
                    <div className="bond-purchase-row" key={p.id}>
                      <span className="mono">{formatShortDate(p.buy_date)}</span>
                      <span>{p.platform || '—'}</span>
                      <span className="mono bond-num">{p.quantity}</span>
                      <span className="mono bond-num">{fmtUsd(p.price_usd)}</span>
                      <span className="mono bond-num">{fmtUsd(p.accrued_interest_usd)}</span>
                      <span className="mono bond-num">{fmtUsd(p.total_usd)}</span>
                      <span className="mono bond-num">{fmtUsdIdrRate(p.usd_idr_at_purchase)}</span>
                      <span className="mono bond-num">{fmtExact(p.total_idr)}</span>
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
                  <button
                    type="button"
                    className="btn-dashed"
                    onClick={() => openAdd(b.bond_name)}
                  >
                    + Add another {b.bond_name} purchase
                  </button>
                </div>
              )}
            </div>
          )
        })}

        {bonds.length > 0 && (
          <div className="bond-net-row">
            <span className="bond-net-label">Total invested</span>
            <span className="mono bond-net-sub">
              coupons {fmtUsd(summary?.coupon_per_year_usd ?? 0)} / yr
            </span>
            <span className="mono bond-net-total">{fmtUsd(summary?.total_usd ?? 0)}</span>
          </div>
        )}
      </div>

      <p className="assets-footnote">
        Bond purchases are permanent — they never copy forward or lock like snapshot holdings do.
        Your Assets page shows one Bonds USD row per bond name, built from this ledger.
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

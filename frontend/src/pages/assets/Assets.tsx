import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCategories } from '../../hooks/useCategories'
import {
  useSnapshots,
  useLatestSnapshot,
  useSnapshotByDate,
  useDeleteSnapshot,
  useSyncSnapshotBonds,
} from '../../hooks/useSnapshots'
import { useLatestRate } from '../../hooks/useRates'
import { api } from '../../lib/api'
import { BOND_CATEGORY_KEYS } from '../../lib/bondCalc'
import { AssetModal } from './AssetModal'
import { NewSnapshotModal } from './NewSnapshotModal'
import type { Holding } from '../../types'
import './Assets.css'

type ModalKind = 'addAsset' | 'newSnapshot' | null

export function Assets() {
  const { fmt } = useMoney()
  const { showError, showSuccess } = useToast()
  const { data: categories = [] } = useCategories()
  const { data: snapshots = [] } = useSnapshots()
  const { data: latestSnapshot } = useLatestSnapshot()
  const { data: latestRate } = useLatestRate()
  const deleteSnapshot = useDeleteSnapshot()
  const syncBonds = useSyncSnapshotBonds()

  const [selectedDate, setSelectedDate] = useState<string | undefined>(undefined)
  const [assetGroup, setAssetGroup] = useState<string>('all')
  const [modal, setModal] = useState<ModalKind>(null)
  const [editingHolding, setEditingHolding] = useState<Holding | null>(null)

  const { data: snapshot, isError: snapshotMissing } = useSnapshotByDate(selectedDate)

  useEffect(() => {
    if (!latestSnapshot) return
    if (!selectedDate || snapshotMissing) setSelectedDate(latestSnapshot.snapshot_date)
  }, [selectedDate, snapshotMissing, latestSnapshot])

  const isViewingLatest = !!snapshot && !!latestSnapshot && snapshot.snapshot_date === latestSnapshot.snapshot_date
  const isEditable = !!snapshot?.is_editable && isViewingLatest

  const holdings = snapshot?.holdings ?? []
  const visibleHoldings = useMemo(
    () => (assetGroup === 'all' ? holdings : holdings.filter((h) => String(h.category_id) === assetGroup)),
    [holdings, assetGroup],
  )

  const grandAssets = useMemo(
    () => holdings.filter((h) => !h.is_liability).reduce((s, h) => s + h.value_idr, 0),
    [holdings],
  )
  const grandLiab = useMemo(
    () => holdings.filter((h) => h.is_liability).reduce((s, h) => s + h.value_idr, 0),
    [holdings],
  )

  const groupTotal = useMemo(() => visibleHoldings.reduce((s, h) => s + h.value_idr, 0), [visibleHoldings])

  const catById = useMemo(() => new Map(categories.map((c) => [c.id, c])), [categories])

  function openAddAsset() {
    setEditingHolding(null)
    setModal('addAsset')
  }
  function openEditAsset(h: Holding) {
    setEditingHolding(h)
    setModal('addAsset')
  }
  function closeModal() {
    setModal(null)
    setEditingHolding(null)
  }

  async function handleDeleteSnapshot() {
    if (!snapshot) return
    const label = new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'long', year: 'numeric' }).format(
      new Date(snapshot.snapshot_date),
    )
    if (!window.confirm(`Delete the ${label} snapshot? This removes it from your history.`)) return
    try {
      await deleteSnapshot.mutateAsync(snapshot.id)
      showSuccess('Snapshot deleted.')
      // Fetch the fresh list directly rather than waiting on background
      // query invalidation to settle — picking the new date from
      // possibly-still-stale cached data would just re-select what was
      // just deleted.
      const fresh = await api.snapshots.list()
      setSelectedDate(fresh[0]?.snapshot_date)
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  // Bonds Gov. and Bonds Corp. are both built from the bond ledger.
  const bondCategoryIds = new Set(
    categories.filter((c) => BOND_CATEGORY_KEYS.includes(c.key)).map((c) => c.id),
  )
  const bondHoldingCount = holdings.filter((h) => bondCategoryIds.has(h.category_id)).length

  async function handleSyncBonds() {
    if (!snapshot) return
    // This replaces every bond row in the snapshot, including any typed by
    // hand, so say so with counts before doing it.
    const existing = bondHoldingCount === 1 ? '1 bond row' : `${bondHoldingCount} bond rows`
    if (
      !window.confirm(
        `Replace ${existing} (Bonds Gov. and Bonds Corp.) in this snapshot with rows built from your bond ledger?\n\nAnything you typed by hand in these categories will be overwritten.`,
      )
    )
      return
    try {
      await syncBonds.mutateAsync(snapshot.snapshot_date)
      showSuccess('Bonds synced from ledger.')
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const defaultCategoryId =
    assetGroup !== 'all'
      ? Number(assetGroup)
      : categories.find((c) => c.label === 'Uang Tunai')?.id ?? categories[0]?.id ?? 0

  return (
    <div>
      <div className="row-wrap assets-toolbar">
        <div className="snapshot-pill">
          <span className="snapshot-pill-label">Snapshot</span>
          <select
            className="snapshot-pill-select"
            value={selectedDate ?? ''}
            onChange={(e) => setSelectedDate(e.target.value)}
          >
            {snapshots.map((s) => (
              <option key={s.id} value={s.snapshot_date}>
                {new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'long', year: 'numeric' }).format(
                  new Date(s.snapshot_date),
                )}
                {s.is_editable ? '' : ' (locked)'}
              </option>
            ))}
          </select>
        </div>
        <div className="btn-group assets-toolbar-actions">
          <button type="button" className="btn btn-secondary" onClick={() => setModal('newSnapshot')}>
            ⧉ Create new snapshot
          </button>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleDeleteSnapshot}
            disabled={!snapshot || deleteSnapshot.isPending}
          >
            🗑 Delete snapshot
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleSyncBonds}
            disabled={!isEditable || syncBonds.isPending}
            title="Rebuild the Bonds Gov. and Bonds Corp. rows from your bond purchase ledger"
          >
            ⟳ Sync bonds
          </button>
          <button type="button" className="btn btn-primary" onClick={openAddAsset} disabled={!isEditable}>
            + Add to this snapshot
          </button>
        </div>
      </div>

      <div className="chips-row assets-chips">
        <button
          type="button"
          className={'chip' + (assetGroup === 'all' ? ' chip-active' : '')}
          onClick={() => setAssetGroup('all')}
        >
          All assets
        </button>
        {categories.map((c) => (
          <button
            key={c.id}
            type="button"
            className={'chip' + (assetGroup === String(c.id) ? ' chip-active' : '')}
            onClick={() => setAssetGroup(String(c.id))}
          >
            {c.label}
          </button>
        ))}
      </div>
      <select
        className="cat-select field-input"
        value={assetGroup}
        onChange={(e) => setAssetGroup(e.target.value)}
      >
        <option value="all">All assets</option>
        {categories.map((c) => (
          <option key={c.id} value={c.id}>
            {c.label}
          </option>
        ))}
      </select>

      <div className="card assets-table-card">
        <div className="asset-head">
          <span className="a-name-head">Asset</span>
          <span className="a-cat-head">Category</span>
          <span className="a-val-head">Value (edit)</span>
        </div>

        {visibleHoldings.length === 0 && <div className="empty-state">No holdings in this snapshot yet.</div>}

        {visibleHoldings.map((h) => {
          const cat = catById.get(h.category_id)
          // Bond rows are derived from the purchase ledger and kept in sync
          // automatically — editing one here would just get overwritten by
          // the next ledger change, so the edit path is the Bonds page.
          const isBondHolding = bondCategoryIds.has(h.category_id)
          return (
            <div className="asset-row" key={h.id}>
              <div className="a-name">
                <div className="a-name-title">{h.name}</div>
                <div className="a-name-detail">{h.detail}</div>
              </div>
              <span className="a-cat">
                <span className="a-cat-swatch" style={{ background: cat?.color_oklch ?? '#ccc' }} />
                {h.category_label}
              </span>
              <span className="a-val">
                <span className="mono a-val-amount">{fmt(h.value_idr)}</span>
                <button
                  type="button"
                  title={isBondHolding ? 'Managed by the bond ledger — edit purchases on the Bonds page' : 'Edit'}
                  className="a-edit-btn"
                  onClick={() => openEditAsset(h)}
                  disabled={!isEditable || isBondHolding}
                >
                  ✎
                </button>
              </span>
            </div>
          )
        })}

        {assetGroup !== 'all' && visibleHoldings.length > 0 && (
          <div className="asset-net-row">
            <span className="asset-net-label">{catById.get(Number(assetGroup))?.label ?? 'Category'} subtotal</span>
            <span className="mono asset-net-breakdown">
              {visibleHoldings.length} {visibleHoldings.length === 1 ? 'holding' : 'holdings'}
            </span>
            <span className="mono asset-net-total">{fmt(groupTotal)}</span>
          </div>
        )}

        {assetGroup === 'all' && holdings.length > 0 && (
          <div className="asset-net-row">
            <span className="asset-net-label">Net equity (assets + liability)</span>
            <span className="mono asset-net-breakdown">
              assets {fmt(grandAssets)} · liab {fmt(grandLiab)}
            </span>
            <span className="mono asset-net-total">{fmt(grandAssets + grandLiab)}</span>
          </div>
        )}
      </div>
      <p className="assets-footnote">
        Values are stored per snapshot date. Editing a value here only changes this month — past months stay
        locked so your history stays accurate.
        {bondCategoryIds.has(Number(assetGroup)) && (
          <>
            {' '}
            Bond rows are built from your purchase ledger —{' '}
            <Link to="/bonds">manage bond ledger →</Link>
          </>
        )}
      </p>

      <AssetModal
        open={modal === 'addAsset'}
        onClose={closeModal}
        snapshotDate={selectedDate ?? ''}
        categories={categories}
        latestRate={latestRate}
        editingHolding={editingHolding}
        defaultCategoryId={defaultCategoryId}
      />
      <NewSnapshotModal
        open={modal === 'newSnapshot'}
        onClose={() => setModal(null)}
        latestSnapshot={snapshots[0]}
        onCreated={(date) => setSelectedDate(date)}
      />
    </div>
  )
}

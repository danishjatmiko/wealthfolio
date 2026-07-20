import { useEffect, useMemo, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useCategories } from '../../hooks/useCategories'
import { useSnapshots, useLatestSnapshot, useSnapshotByDate } from '../../hooks/useSnapshots'
import { useLatestRate } from '../../hooks/useRates'
import { AssetModal } from './AssetModal'
import { NewSnapshotModal } from './NewSnapshotModal'
import type { Holding } from '../../types'
import './Assets.css'

type ModalKind = 'addAsset' | 'newSnapshot' | null

export function Assets() {
  const { fmt } = useMoney()
  const { data: categories = [] } = useCategories()
  const { data: snapshots = [] } = useSnapshots()
  const { data: latestSnapshot } = useLatestSnapshot()
  const { data: latestRate } = useLatestRate()

  const [selectedDate, setSelectedDate] = useState<string | undefined>(undefined)
  const [assetGroup, setAssetGroup] = useState<string>('all')
  const [modal, setModal] = useState<ModalKind>(null)
  const [editingHolding, setEditingHolding] = useState<Holding | null>(null)

  useEffect(() => {
    if (!selectedDate && latestSnapshot) setSelectedDate(latestSnapshot.snapshot_date)
  }, [selectedDate, latestSnapshot])

  const { data: snapshot } = useSnapshotByDate(selectedDate)

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
            ⧉ Create new with copy this data
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
                  title="Edit"
                  className="a-edit-btn"
                  onClick={() => openEditAsset(h)}
                  disabled={!isEditable}
                >
                  ✎
                </button>
              </span>
            </div>
          )
        })}

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

import { useEffect, useMemo, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useDebtSnapshots,
  useLatestDebtSnapshot,
  useDebtSnapshotByDate,
  useDeleteDebtSnapshot,
} from '../../hooks/useDebtSnapshots'
import { useDashboard } from '../../hooks/useDashboard'
import { useTargets } from '../../hooks/useTargets'
import { api } from '../../lib/api'
import { DebtModal } from './DebtModal'
import { DebtSnapshotModal } from './DebtSnapshotModal'
import type { DebtEntry, DebtDirection } from '../../types'
import './Debts.css'

export function Debts() {
  const { fmt } = useMoney()
  const { showError, showSuccess } = useToast()
  const { data: snapshots = [] } = useDebtSnapshots()
  const { data: latestSnapshot } = useLatestDebtSnapshot()
  const { data: dashboard } = useDashboard()
  const { data: targets = [] } = useTargets()
  const deleteSnapshot = useDeleteDebtSnapshot()

  const [selectedDate, setSelectedDate] = useState<string | undefined>(undefined)
  const [modalDirection, setModalDirection] = useState<DebtDirection | null>(null)
  const [editingEntry, setEditingEntry] = useState<DebtEntry | null>(null)
  const [showNewSnapshot, setShowNewSnapshot] = useState(false)

  const { data: snapshot, isError: snapshotMissing } = useDebtSnapshotByDate(selectedDate)

  useEffect(() => {
    if (!latestSnapshot) return
    if (!selectedDate || snapshotMissing) setSelectedDate(latestSnapshot.snapshot_date)
  }, [selectedDate, snapshotMissing, latestSnapshot])

  const isViewingLatest = !!snapshot && !!latestSnapshot && snapshot.snapshot_date === latestSnapshot.snapshot_date
  const isEditable = !!snapshot?.is_editable && isViewingLatest

  const entries = snapshot?.entries ?? []
  const myDebts = useMemo(() => entries.filter((d) => d.direction === 'i_owe'), [entries])
  const receivables = useMemo(() => entries.filter((d) => d.direction === 'owed_to_me'), [entries])
  const totalDebt = useMemo(() => myDebts.reduce((s, d) => s + d.value_idr, 0), [myDebts])
  const totalReceivable = useMemo(() => receivables.reduce((s, d) => s + d.value_idr, 0), [receivables])

  const ratioTarget = targets.find((t) => t.metric_type === 'debt_ratio')
  const ratioPct = isViewingLatest ? (dashboard?.debt.ratio_pct ?? 0) : undefined

  function openAdd(direction: DebtDirection) {
    if (!isEditable) return
    setEditingEntry(null)
    setModalDirection(direction)
  }
  function openEdit(d: DebtEntry) {
    if (!isEditable) return
    setEditingEntry(d)
    setModalDirection(d.direction)
  }
  function closeModal() {
    setModalDirection(null)
    setEditingEntry(null)
  }

  async function handleDeleteSnapshot() {
    if (!snapshot) return
    const label = new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'long', year: 'numeric' }).format(
      new Date(snapshot.snapshot_date),
    )
    if (!window.confirm(`Delete the ${label} debt snapshot? This removes it from your history.`)) return
    try {
      await deleteSnapshot.mutateAsync(snapshot.id)
      showSuccess('Debt snapshot deleted.')
      // Fetch the fresh list directly rather than waiting on background
      // query invalidation to settle — picking the new date from
      // possibly-still-stale cached data would just re-select what was
      // just deleted.
      const fresh = await api.debtSnapshots.list()
      setSelectedDate(fresh[0]?.snapshot_date)
    } catch (err) {
      showError(errorMessage(err))
    }
  }

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
          <button type="button" className="btn btn-secondary" onClick={() => setShowNewSnapshot(true)}>
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
        </div>
      </div>

      <div className="debts-grid">
        <div className="card">
          <div className="debts-card-head">
            <div className="card-title-inline">My debts</div>
            <div className="mono debts-card-total">{fmt(totalDebt)}</div>
          </div>
          {myDebts.length === 0 && <div className="empty-state">No debts logged.</div>}
          {myDebts.map((d) => (
            <div
              className={'debt-row' + (isEditable ? '' : ' debt-row-locked')}
              key={d.id}
              onClick={() => openEdit(d)}
            >
              <div className="debt-row-info">
                <div className="debt-row-name">{d.name}</div>
                <div className="debt-row-type">{d.type}</div>
              </div>
              <span className="mono debt-row-val debt-row-val-red">{fmt(d.value_idr)}</span>
            </div>
          ))}
          <button
            type="button"
            className="btn-dashed"
            onClick={() => openAdd('i_owe')}
            disabled={!isEditable}
          >
            + Add debt
          </button>
        </div>

        <div className="card">
          <div className="debts-card-head">
            <div className="card-title-inline">Who owes me</div>
            <div className="mono debts-card-total">{fmt(totalReceivable)}</div>
          </div>
          {receivables.length === 0 && <div className="empty-state">Nobody owes you (yet).</div>}
          {receivables.map((d) => (
            <div
              className={'debt-row' + (isEditable ? '' : ' debt-row-locked')}
              key={d.id}
              onClick={() => openEdit(d)}
            >
              <div className="debt-row-info">
                <div className="debt-row-name">{d.name}</div>
                <div className="debt-row-type">{d.type}</div>
              </div>
              <span className="mono debt-row-val debt-row-val-green">{fmt(d.value_idr)}</span>
            </div>
          ))}
          <button
            type="button"
            className="btn-dashed"
            onClick={() => openAdd('owed_to_me')}
            disabled={!isEditable}
          >
            + Add receivable
          </button>
        </div>

        <div className="debt-ratio-banner">
          <div>
            <div className="debt-ratio-label">Debt-to-equity ratio</div>
            <div className="mono debt-ratio-value">{ratioPct !== undefined ? `${ratioPct.toFixed(2)}%` : '—'}</div>
            {ratioTarget && ratioPct !== undefined && (
              <div className="debt-ratio-note">
                Goal: keep below {ratioTarget.target_value.toFixed(2)}
                {ratioTarget.unit} — you have {Math.max(0, ratioTarget.target_value - ratioPct).toFixed(2)}
                {ratioTarget.unit} of headroom.
              </div>
            )}
            {ratioPct === undefined && (
              <div className="debt-ratio-note">Only shown for the current snapshot.</div>
            )}
          </div>
        </div>
      </div>
      <p className="assets-footnote">
        Debts are stored per snapshot date, just like Assets. Editing here only changes the current snapshot —
        past ones stay locked so your history stays accurate.
      </p>

      <DebtModal
        open={modalDirection !== null}
        onClose={closeModal}
        direction={modalDirection ?? 'i_owe'}
        editingEntry={editingEntry}
        snapshotDate={selectedDate ?? ''}
      />
      <DebtSnapshotModal
        open={showNewSnapshot}
        onClose={() => setShowNewSnapshot(false)}
        latestSnapshot={snapshots[0]}
        onCreated={(date) => setSelectedDate(date)}
      />
    </div>
  )
}

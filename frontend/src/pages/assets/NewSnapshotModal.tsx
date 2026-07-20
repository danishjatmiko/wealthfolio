import { useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateSnapshot } from '../../hooks/useSnapshots'
import { useDebts } from '../../hooks/useDebts'
import type { SnapshotSummary } from '../../types'

interface NewSnapshotModalProps {
  open: boolean
  onClose: () => void
  latestSnapshot: SnapshotSummary | undefined
  onCreated: (date: string) => void
}

function todayIso(): string {
  return new Date().toISOString().slice(0, 10)
}

export function NewSnapshotModal({ open, onClose, latestSnapshot, onCreated }: NewSnapshotModalProps) {
  const { fmt } = useMoney()
  const { showError, showSuccess } = useToast()
  const { data: debts } = useDebts()
  const createSnapshot = useCreateSnapshot()
  const [date, setDate] = useState(todayIso())

  async function handleCreate() {
    if (!date) {
      showError('Pick a date for the new snapshot.')
      return
    }
    try {
      const snap = await createSnapshot.mutateAsync({ snapshot_date: date, copy_from_latest: true })
      showSuccess('Snapshot created.')
      onCreated(snap.snapshot_date)
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create new snapshot"
      subtitle="Start next month by copying this data, then edit values."
      footer={
        <>
          <ModalCancelButton onClick={onClose} />
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleCreate}
            disabled={createSnapshot.isPending}
          >
            Create snapshot
          </button>
        </>
      }
    >
      <label className="field">
        New snapshot date
        <input
          type="date"
          className="field-input"
          value={date}
          onChange={(e) => setDate(e.target.value)}
        />
      </label>
      <div className="computed-box" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 6 }}>
        <div className="snap-copy-row">
          <span>Assets copied from {latestSnapshot?.snapshot_date ?? '—'}</span>
          <b className="mono">{latestSnapshot?.holdings_count ?? 0} rows</b>
        </div>
        <div className="snap-copy-row">
          <span>Debts &amp; receivables</span>
          <b className="mono">{debts?.length ?? 0} rows</b>
        </div>
        <div className="snap-copy-row">
          <span>Starting net equity</span>
          <b className="mono">{latestSnapshot ? fmt(latestSnapshot.net_equity_idr) : '—'}</b>
        </div>
      </div>
      <div className="snap-copy-note">Values carry over so you only change what moved — no more copy-pasting tabs.</div>
    </Modal>
  )
}

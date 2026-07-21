import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateDebtSnapshot } from '../../hooks/useDebtSnapshots'
import type { DebtSnapshotSummary } from '../../types'

interface DebtSnapshotModalProps {
  open: boolean
  onClose: () => void
  latestSnapshot: DebtSnapshotSummary | undefined
  onCreated: (date: string) => void
}

function todayIso(): string {
  return new Date().toISOString().slice(0, 10)
}

export function DebtSnapshotModal({ open, onClose, latestSnapshot, onCreated }: DebtSnapshotModalProps) {
  const { fmt } = useMoney()
  const { showError, showSuccess } = useToast()
  const createSnapshot = useCreateDebtSnapshot()
  const [date, setDate] = useState(todayIso())
  const [copyData, setCopyData] = useState(true)

  useEffect(() => {
    if (open) {
      setDate(todayIso())
      setCopyData(true)
    }
  }, [open])

  async function handleCreate() {
    if (!date) {
      showError('Pick a date for the new snapshot.')
      return
    }
    try {
      const snap = await createSnapshot.mutateAsync({ snapshot_date: date, copy_from_latest: copyData })
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
      title="Create new debt snapshot"
      subtitle="Start today (or a future date) with a copy of your current data, or blank."
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
          min={todayIso()}
          onChange={(e) => setDate(e.target.value)}
        />
      </label>
      <label className="field-checkbox">
        <input type="checkbox" checked={copyData} onChange={(e) => setCopyData(e.target.checked)} />
        Copy current data into the new snapshot
      </label>
      {copyData && (
        <>
          <div className="computed-box" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 6 }}>
            <div className="snap-copy-row">
              <span>Entries copied from {latestSnapshot?.snapshot_date ?? '—'}</span>
              <b className="mono">{latestSnapshot?.entries_count ?? 0} rows</b>
            </div>
            <div className="snap-copy-row">
              <span>Starting debt / owed to me</span>
              <b className="mono">
                {latestSnapshot ? fmt(latestSnapshot.i_owe_idr) : '—'} /{' '}
                {latestSnapshot ? fmt(latestSnapshot.owed_to_me_idr) : '—'}
              </b>
            </div>
          </div>
          <div className="snap-copy-note">Values carry over so you only change what moved.</div>
        </>
      )}
    </Modal>
  )
}

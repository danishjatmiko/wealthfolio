import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateDebtEntry, useUpdateDebtEntry } from '../../hooks/useDebtSnapshots'
import { parseNumeric } from '../../lib/format'
import type { DebtEntry, DebtDirection } from '../../types'

const DEBT_TYPES = ['KPR', 'Credit Card', 'Personal loan', 'Vehicle loan', 'Other']
const RECEIVABLE_TYPES = ['Personal loan', 'Business', 'Other']

interface DebtModalProps {
  open: boolean
  onClose: () => void
  direction: DebtDirection
  editingEntry: DebtEntry | null
  snapshotDate: string
}

export function DebtModal({ open, onClose, direction, editingEntry, snapshotDate }: DebtModalProps) {
  const { showError, showSuccess } = useToast()
  const createEntry = useCreateDebtEntry()
  const updateEntry = useUpdateDebtEntry()
  const types = direction === 'i_owe' ? DEBT_TYPES : RECEIVABLE_TYPES

  const [name, setName] = useState('')
  const [type, setType] = useState(types[0])
  const [amount, setAmount] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingEntry) {
      setName(editingEntry.name)
      setType(editingEntry.type)
      setAmount(String(editingEntry.value_idr))
    } else {
      setName('')
      setType(types[0])
      setAmount('')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, editingEntry])

  const isDebt = direction === 'i_owe'
  const title = isDebt ? (editingEntry ? 'Edit debt' : 'Add debt') : editingEntry ? 'Edit receivable' : 'Add receivable'
  const subtitle = isDebt ? 'Something you owe.' : 'Someone who owes you.'
  const nameLabel = isDebt ? 'Debt name' : 'Person / name'
  const namePlaceholder = isDebt ? 'e.g. OCBC KPA' : 'e.g. Edo Tole'
  const cta = isDebt ? (editingEntry ? 'Save changes' : 'Add debt') : editingEntry ? 'Save changes' : 'Add receivable'

  async function handleSave() {
    if (!name.trim()) {
      showError('Give it a name.')
      return
    }
    const input = { name: name.trim(), type, value_idr: Math.round(parseNumeric(amount)), direction }
    try {
      if (editingEntry) {
        await updateEntry.mutateAsync({ id: editingEntry.id, input })
        showSuccess('Updated.')
      } else {
        await createEntry.mutateAsync({ date: snapshotDate, input })
        showSuccess(isDebt ? 'Debt added.' : 'Receivable added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createEntry.isPending || updateEntry.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      subtitle={subtitle}
      footer={
        <>
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {cta}
          </button>
        </>
      }
    >
      <label className="field">
        {nameLabel}
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={namePlaceholder}
        />
      </label>
      <label className="field">
        Type
        <select className="field-input" value={type} onChange={(e) => setType(e.target.value)}>
          {types.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </label>
      <label className="field">
        Amount (rb Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder={isDebt ? '8800' : '4800'}
        />
      </label>
    </Modal>
  )
}

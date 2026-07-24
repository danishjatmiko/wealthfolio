import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useCreateBudgetEnvelope,
  useDeleteBudgetEnvelope,
  useUpdateBudgetEnvelope,
} from '../../hooks/useBudgetEnvelopes'
import { parseNumeric } from '../../lib/format'
import type { BudgetEnvelopeDetail } from '../../types'

interface BudgetEnvelopeModalProps {
  open: boolean
  onClose: () => void
  periodId: string
  editingEnvelope: BudgetEnvelopeDetail | null
}

export function BudgetEnvelopeModal({ open, onClose, periodId, editingEnvelope }: BudgetEnvelopeModalProps) {
  const { showError, showSuccess } = useToast()
  const createEnvelope = useCreateBudgetEnvelope()
  const updateEnvelope = useUpdateBudgetEnvelope()
  const deleteEnvelope = useDeleteBudgetEnvelope()

  const [name, setName] = useState('')
  const [amount, setAmount] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingEnvelope) {
      setName(editingEnvelope.name)
      setAmount(String(editingEnvelope.committed_amount_idr))
    } else {
      setName('')
      setAmount('')
    }
  }, [open, editingEnvelope])

  async function handleSave() {
    if (!name.trim()) {
      showError('Give it a name.')
      return
    }
    const input = {
      name: name.trim(),
      committed_amount_idr: Math.round(parseNumeric(amount)),
    }
    try {
      if (editingEnvelope) {
        await updateEnvelope.mutateAsync({ id: editingEnvelope.id, input })
        showSuccess('Envelope updated.')
      } else {
        await createEnvelope.mutateAsync({ periodId, input })
        showSuccess('Envelope added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  async function handleDelete() {
    if (!editingEnvelope) return
    if (
      !window.confirm(
        `Delete "${editingEnvelope.name}"? This also deletes every expense inside it and cannot be undone.`,
      )
    )
      return
    try {
      await deleteEnvelope.mutateAsync(editingEnvelope.id)
      showSuccess('Envelope deleted.')
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createEnvelope.isPending || updateEnvelope.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editingEnvelope ? 'Edit budget envelope' : 'Add budget envelope'}
      subtitle="A monthly target that bundles several expenses."
      footer={
        <>
          {editingEnvelope && (
            <button
              type="button"
              className="btn btn-danger"
              style={{ marginRight: 'auto' }}
              onClick={handleDelete}
              disabled={deleteEnvelope.isPending}
            >
              🗑 Delete
            </button>
          )}
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {editingEnvelope ? 'Save changes' : 'Add envelope'}
          </button>
        </>
      }
    >
      <label className="field">
        Envelope name
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Kebutuhan Keluarga Inti"
        />
      </label>
      <label className="field">
        Committed target (rb Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="20000"
        />
      </label>
    </Modal>
  )
}

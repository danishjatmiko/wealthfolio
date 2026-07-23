import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateFixedExpense, useDeleteFixedExpense, useUpdateFixedExpense } from '../../hooks/useFixedExpenses'
import { parseNumeric } from '../../lib/format'
import type { BudgetEnvelopeDetail, FixedExpense } from '../../types'

interface FixedExpenseModalProps {
  open: boolean
  onClose: () => void
  periodId: string
  envelopes: BudgetEnvelopeDetail[]
  editingExpense: FixedExpense | null
  defaultEnvelopeId: string | null
}

export function FixedExpenseModal({
  open,
  onClose,
  periodId,
  envelopes,
  editingExpense,
  defaultEnvelopeId,
}: FixedExpenseModalProps) {
  const { showError, showSuccess } = useToast()
  const createExpense = useCreateFixedExpense()
  const updateExpense = useUpdateFixedExpense()
  const deleteExpense = useDeleteFixedExpense()

  const [name, setName] = useState('')
  const [amount, setAmount] = useState('')
  const [envelopeId, setEnvelopeId] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingExpense) {
      setName(editingExpense.name)
      setAmount(String(editingExpense.amount_idr))
      setEnvelopeId(editingExpense.envelope_id)
    } else {
      setName('')
      setAmount('')
      setEnvelopeId(defaultEnvelopeId ?? envelopes[0]?.id ?? '')
    }
  }, [open, editingExpense, defaultEnvelopeId, envelopes])

  async function handleSave() {
    if (!name.trim()) {
      showError('Give it a name.')
      return
    }
    if (!envelopeId) {
      showError('Pick a budget envelope.')
      return
    }
    const input = {
      name: name.trim(),
      amount_idr: Math.round(parseNumeric(amount)),
      envelope_id: envelopeId,
    }
    try {
      if (editingExpense) {
        await updateExpense.mutateAsync({ id: editingExpense.id, input })
        showSuccess('Updated.')
      } else {
        await createExpense.mutateAsync({ periodId, input })
        showSuccess('Expense added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  async function handleDelete() {
    if (!editingExpense) return
    if (!window.confirm(`Delete "${editingExpense.name}"? This cannot be undone.`)) return
    try {
      await deleteExpense.mutateAsync(editingExpense.id)
      showSuccess('Expense deleted.')
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createExpense.isPending || updateExpense.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editingExpense ? 'Edit expense' : 'Add expense'}
      subtitle="A real amount you spent."
      footer={
        <>
          {editingExpense && (
            <button
              type="button"
              className="btn btn-danger"
              style={{ marginRight: 'auto' }}
              onClick={handleDelete}
              disabled={deleteExpense.isPending}
            >
              🗑 Delete
            </button>
          )}
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {editingExpense ? 'Save changes' : 'Add expense'}
          </button>
        </>
      }
    >
      <label className="field">
        Expense name
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Kasi Mojokerto"
        />
      </label>
      <label className="field">
        Budget envelope
        <select className="field-input" value={envelopeId} onChange={(e) => setEnvelopeId(e.target.value)}>
          {envelopes.map((env) => (
            <option key={env.id} value={env.id}>
              {env.name}
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
          placeholder="3000"
        />
      </label>
    </Modal>
  )
}

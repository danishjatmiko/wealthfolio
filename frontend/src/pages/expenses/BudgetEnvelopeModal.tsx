import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useCreateBudgetEnvelope,
  useDeleteBudgetEnvelope,
  useUpdateBudgetEnvelope,
} from '../../hooks/useBudgetEnvelopes'
import { useCreateExpenseCategory, useExpenseCategories } from '../../hooks/useExpenseCategories'
import { parseNumeric } from '../../lib/format'
import type { BudgetEnvelopeDetail } from '../../types'

interface BudgetEnvelopeModalProps {
  open: boolean
  onClose: () => void
  periodId: string
  editingEnvelope: BudgetEnvelopeDetail | null
}

const NEW_CATEGORY_VALUE = '__new__'

export function BudgetEnvelopeModal({ open, onClose, periodId, editingEnvelope }: BudgetEnvelopeModalProps) {
  const { showError, showSuccess } = useToast()
  const createEnvelope = useCreateBudgetEnvelope()
  const updateEnvelope = useUpdateBudgetEnvelope()
  const deleteEnvelope = useDeleteBudgetEnvelope()
  const { data: categories = [] } = useExpenseCategories()
  const createCategory = useCreateExpenseCategory()

  const [name, setName] = useState('')
  const [amount, setAmount] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [newCategoryName, setNewCategoryName] = useState('')

  useEffect(() => {
    if (!open) return
    setNewCategoryName('')
    if (editingEnvelope) {
      setName(editingEnvelope.name)
      setAmount(String(editingEnvelope.committed_amount_idr))
      setCategoryId(editingEnvelope.category_id)
    } else {
      setName('')
      setAmount('')
      setCategoryId(categories[0]?.id ?? NEW_CATEGORY_VALUE)
    }
  }, [open, editingEnvelope, categories])

  async function handleSave() {
    if (!name.trim()) {
      showError('Give it a name.')
      return
    }
    let resolvedCategoryId = categoryId
    if (categoryId === NEW_CATEGORY_VALUE) {
      if (!newCategoryName.trim()) {
        showError('Give the new category a name.')
        return
      }
      try {
        const created = await createCategory.mutateAsync({ name: newCategoryName.trim() })
        resolvedCategoryId = created.id
      } catch (err) {
        showError(errorMessage(err))
        return
      }
    }
    const input = {
      category_id: resolvedCategoryId,
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

  const saving = createEnvelope.isPending || updateEnvelope.isPending || createCategory.isPending

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
        Category
        <select className="field-input" value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          {categories.map((cat) => (
            <option key={cat.id} value={cat.id}>
              {cat.name}
            </option>
          ))}
          <option value={NEW_CATEGORY_VALUE}>+ New category…</option>
        </select>
      </label>
      {categoryId === NEW_CATEGORY_VALUE && (
        <label className="field">
          New category name
          <input
            className="field-input"
            value={newCategoryName}
            onChange={(e) => setNewCategoryName(e.target.value)}
            placeholder="e.g. Kebutuhan Pokok"
          />
        </label>
      )}
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

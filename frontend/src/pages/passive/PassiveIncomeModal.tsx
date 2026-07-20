import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreatePassiveIncome, useUpdatePassiveIncome } from '../../hooks/usePassiveIncome'
import { parseNumeric } from '../../lib/format'
import type { Category, PassiveIncomeSource } from '../../types'

interface PassiveIncomeModalProps {
  open: boolean
  onClose: () => void
  categories: Category[]
  editingSource: PassiveIncomeSource | null
}

export function PassiveIncomeModal({ open, onClose, categories, editingSource }: PassiveIncomeModalProps) {
  const { showError, showSuccess } = useToast()
  const createSource = useCreatePassiveIncome()
  const updateSource = useUpdatePassiveIncome()

  const [categoryId, setCategoryId] = useState<number>(0)
  const [name, setName] = useState('')
  const [amount, setAmount] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingSource) {
      setCategoryId(editingSource.category_id)
      setName(editingSource.name)
      setAmount(String(editingSource.per_year_idr))
    } else {
      setCategoryId(categories[0]?.id ?? 0)
      setName('')
      setAmount('')
    }
  }, [open, editingSource, categories])

  async function handleSave() {
    if (!categoryId) {
      showError('Choose a category first.')
      return
    }
    const cat = categories.find((c) => c.id === categoryId)
    const input = {
      category_id: categoryId,
      name: name.trim() || cat?.label || 'Income source',
      per_year_idr: Math.round(parseNumeric(amount)),
    }
    try {
      if (editingSource) {
        await updateSource.mutateAsync({ id: editingSource.id, input })
        showSuccess('Updated.')
      } else {
        await createSource.mutateAsync(input)
        showSuccess('Income source added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createSource.isPending || updateSource.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editingSource ? 'Edit income source' : 'Add income source'}
      subtitle="Estimated passive income per year."
      footer={
        <>
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {editingSource ? 'Save changes' : 'Add source'}
          </button>
        </>
      }
    >
      <label className="field">
        Category
        <select
          className="field-input"
          value={categoryId}
          onChange={(e) => setCategoryId(Number(e.target.value))}
        >
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.label}
            </option>
          ))}
        </select>
      </label>
      <label className="field">
        Label (optional)
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Dividends"
        />
      </label>
      <label className="field">
        Estimated / year (rb Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="127100"
        />
      </label>
    </Modal>
  )
}

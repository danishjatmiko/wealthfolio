import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useCreateBigExpense,
  useDeleteBigExpense,
  useUpdateBigExpense,
} from '../../hooks/useBigExpenses'
import { BIG_EXPENSE_CATEGORIES } from '../../lib/bigExpenseCategories'
import { fmtIdrExact, parseNumeric } from '../../lib/format'
import type { BigExpense } from '../../types'

interface BigExpenseModalProps {
  open: boolean
  onClose: () => void
  editingExpense: BigExpense | null
}

const OTHER_CATEGORY = '__other__'

export function BigExpenseModal({ open, onClose, editingExpense }: BigExpenseModalProps) {
  const { showError, showSuccess } = useToast()
  const createExpense = useCreateBigExpense()
  const updateExpense = useUpdateBigExpense()
  const deleteExpense = useDeleteBigExpense()

  const [name, setName] = useState('')
  const [category, setCategory] = useState<string>(BIG_EXPENSE_CATEGORIES[0])
  const [otherCategory, setOtherCategory] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingExpense) {
      const known = (BIG_EXPENSE_CATEGORIES as readonly string[]).includes(editingExpense.category)
      setName(editingExpense.name)
      setCategory(known ? editingExpense.category : OTHER_CATEGORY)
      setOtherCategory(known ? '' : editingExpense.category)
      setAmount(String(editingExpense.amount_idr))
      setDate(editingExpense.expense_date)
    } else {
      setName('')
      setCategory(BIG_EXPENSE_CATEGORIES[0])
      setOtherCategory('')
      setAmount('')
      setDate(new Date().toISOString().slice(0, 10))
    }
  }, [open, editingExpense])

  const resolvedCategory = category === OTHER_CATEGORY ? otherCategory.trim() : category
  const amountIdr = Math.round(parseNumeric(amount))

  async function handleSave() {
    if (!name.trim()) {
      showError('Give the expense a name.')
      return
    }
    if (!date) {
      showError('The date is required.')
      return
    }
    if (amountIdr <= 0) {
      showError('Amount must be greater than 0.')
      return
    }
    const input = {
      name: name.trim(),
      amount_idr: amountIdr,
      expense_date: date,
      category: resolvedCategory,
    }
    try {
      if (editingExpense) {
        await updateExpense.mutateAsync({ id: editingExpense.id, input })
        showSuccess('Expense updated.')
      } else {
        await createExpense.mutateAsync(input)
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
      title={editingExpense ? 'Edit big expense' : 'Add big expense'}
      subtitle="A large one-off purchase — not part of the monthly budget."
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
        Name
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Hotel Bali 2 malam"
        />
      </label>

      <div className="field-row">
        <label className="field">
          Category
          <select
            className="field-input"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          >
            {BIG_EXPENSE_CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
            <option value={OTHER_CATEGORY}>Other…</option>
          </select>
        </label>
        <label className="field">
          Date
          <input
            type="date"
            className="field-input"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </label>
      </div>

      {category === OTHER_CATEGORY && (
        <label className="field">
          Category name
          <input
            className="field-input"
            value={otherCategory}
            onChange={(e) => setOtherCategory(e.target.value)}
            placeholder="e.g. Wedding"
          />
        </label>
      )}

      <label className="field">
        Amount (Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="1800000"
        />
      </label>

      {amountIdr > 0 && (
        <div className="computed-box">
          <div>
            <div className="computed-box-label">Amount</div>
            <div className="computed-box-value mono">{fmtIdrExact(amountIdr)}</div>
          </div>
        </div>
      )}
    </Modal>
  )
}

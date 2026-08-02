import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useCreatePassiveIncome,
  useDeletePassiveIncome,
  useUpdatePassiveIncome,
} from '../../hooks/usePassiveIncome'
import { PASSIVE_INCOME_TYPES } from '../../lib/passiveIncomeTypes'
import { fmtIdrExact, parseSignedNumeric } from '../../lib/format'
import type { Category, PassiveIncomeEntry } from '../../types'

interface PassiveIncomeModalProps {
  open: boolean
  onClose: () => void
  categories: Category[]
  editingEntry: PassiveIncomeEntry | null
}

const OTHER_TYPE = '__other__'

export function PassiveIncomeModal({
  open,
  onClose,
  categories,
  editingEntry,
}: PassiveIncomeModalProps) {
  const { showError, showSuccess } = useToast()
  const createEntry = useCreatePassiveIncome()
  const updateEntry = useUpdatePassiveIncome()
  const deleteEntry = useDeletePassiveIncome()

  const [categoryId, setCategoryId] = useState<number>(0)
  const [name, setName] = useState('')
  const [incomeType, setIncomeType] = useState<string>(PASSIVE_INCOME_TYPES[0])
  const [otherType, setOtherType] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState('')
  const [note, setNote] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingEntry) {
      const known = (PASSIVE_INCOME_TYPES as readonly string[]).includes(editingEntry.income_type)
      setCategoryId(editingEntry.category_id)
      setName(editingEntry.name)
      setIncomeType(known ? editingEntry.income_type : OTHER_TYPE)
      setOtherType(known ? '' : editingEntry.income_type)
      setAmount(String(editingEntry.amount_idr))
      setDate(editingEntry.received_date)
      setNote(editingEntry.note)
    } else {
      setCategoryId(categories[0]?.id ?? 0)
      setName('')
      setIncomeType(PASSIVE_INCOME_TYPES[0])
      setOtherType('')
      setAmount('')
      setDate(new Date().toISOString().slice(0, 10))
      setNote('')
    }
  }, [open, editingEntry, categories])

  const resolvedType = incomeType === OTHER_TYPE ? otherType.trim() : incomeType
  const amountIdr = Math.round(parseSignedNumeric(amount))

  async function handleSave() {
    if (!categoryId) {
      showError('Choose a category first.')
      return
    }
    if (!name.trim()) {
      showError('Give the entry a name.')
      return
    }
    if (!date) {
      showError('The date is required.')
      return
    }
    // Negatives are allowed on purpose — a realized capital loss belongs in
    // this ledger at its true sign. Only zero is meaningless.
    if (amountIdr === 0) {
      showError('Amount cannot be 0.')
      return
    }
    const input = {
      category_id: categoryId,
      name: name.trim(),
      amount_idr: amountIdr,
      received_date: date,
      income_type: resolvedType,
      note: note.trim(),
    }
    try {
      if (editingEntry) {
        await updateEntry.mutateAsync({ id: editingEntry.id, input })
        showSuccess('Updated.')
      } else {
        await createEntry.mutateAsync(input)
        showSuccess('Income added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  async function handleDelete() {
    if (!editingEntry) return
    if (!window.confirm(`Delete "${editingEntry.name}"? This cannot be undone.`)) return
    try {
      await deleteEntry.mutateAsync(editingEntry.id)
      showSuccess('Income deleted.')
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
      title={editingEntry ? 'Edit income' : 'Add income'}
      subtitle="One payment actually received, on the date it landed."
      footer={
        <>
          {editingEntry && (
            <button
              type="button"
              className="btn btn-danger"
              style={{ marginRight: 'auto' }}
              onClick={handleDelete}
              disabled={deleteEntry.isPending}
            >
              🗑 Delete
            </button>
          )}
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {editingEntry ? 'Save changes' : 'Add income'}
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
          placeholder="e.g. Dividen BBRI"
        />
      </label>

      <div className="field-row">
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
          Date received
          <input
            type="date"
            className="field-input"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </label>
      </div>

      <label className="field">
        Type
        <select
          className="field-input"
          value={incomeType}
          onChange={(e) => setIncomeType(e.target.value)}
        >
          {PASSIVE_INCOME_TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
          <option value={OTHER_TYPE}>Other…</option>
        </select>
      </label>

      {incomeType === OTHER_TYPE && (
        <label className="field">
          Type name
          <input
            className="field-input"
            value={otherType}
            onChange={(e) => setOtherType(e.target.value)}
            placeholder="e.g. Cashback"
          />
        </label>
      )}

      <label className="field">
        Amount (Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="11000000"
        />
      </label>

      <label className="field">
        Note (optional)
        <input
          className="field-input"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g. kena harga rata2, ngitungnya rugi"
        />
      </label>

      {amountIdr !== 0 && (
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

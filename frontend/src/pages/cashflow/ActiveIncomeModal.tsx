import { useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useDeleteActiveIncome, useSaveActiveIncome } from '../../hooks/useCashflow'
import type { CashflowMonth } from '../../types'

// Mounted afresh for each selected month so unsaved values never bleed into another month.
export function ActiveIncomeModal({ year, month, onClose }: {
  year: number
  month: CashflowMonth
  onClose: () => void
}) {
  const { fmtExact } = useMoney()
  const { showError, showSuccess } = useToast()
  const save = useSaveActiveIncome()
  const remove = useDeleteActiveIncome()
  const [amount, setAmount] = useState(month.active_income ? String(month.active_income.amount_idr) : '')
  const [note, setNote] = useState(month.active_income?.note ?? '')
  const normalizedAmount = amount.trim()
  const amountIdr = Number(normalizedAmount.replace(/,/g, ''))
  const valid = /^(?:\d+|\d{1,3}(?:,\d{3})+)$/.test(normalizedAmount)
    && Number.isSafeInteger(amountIdr) && amountIdr >= 0
  const monthKey = `${year}-${String(month.month).padStart(2, '0')}`
  const busy = save.isPending || remove.isPending
  const close = () => { if (!busy) onClose() }

  async function handleSave() {
    if (!valid) { showError('Enter a non-negative whole Rupiah amount.'); return }
    try {
      await save.mutateAsync({ month: monthKey, input: { amount_idr: amountIdr, note: note.trim() } })
      showSuccess('Active income saved.')
      onClose()
    } catch (err) { showError(errorMessage(err)) }
  }

  async function handleDelete() {
    if (!window.confirm(`Remove active income for ${month.label} ${year}?`)) return
    try {
      await remove.mutateAsync(monthKey)
      showSuccess('Active income removed.')
      onClose()
    } catch (err) { showError(errorMessage(err)) }
  }

  return (
    <Modal open onClose={close} title={`${month.active_income ? 'Edit' : 'Add'} active income`}
      subtitle={`${month.label} ${year} · Total salary, freelance work, and other active income.`}
      footer={<>
        {month.active_income && <button type="button" className="btn btn-danger" disabled={busy}
          style={{ marginRight: 'auto' }} onClick={handleDelete}>Remove</button>}
        <ModalCancelButton onClick={close} />
        <button type="button" className="btn btn-primary" disabled={busy || !valid} onClick={handleSave}>
          {save.isPending ? 'Saving…' : 'Save income'}
        </button>
      </>}>
      <label className="field">Total active income (Rp)
        <input autoFocus className="field-input mono" inputMode="numeric" value={amount}
          onChange={(e) => setAmount(e.target.value)} placeholder="e.g. 25000000" disabled={busy} />
      </label>
      <div className="field-hint">Enter 0 to confirm no active income this month. Saving replaces this month’s total.</div>
      <label className="field">Note (optional)
        <textarea className="field-input" rows={3} value={note} maxLength={500} disabled={busy}
          onChange={(e) => setNote(e.target.value)} placeholder="e.g. Salary + freelance project" />
      </label>
      {valid && <div className="computed-box">
        <div><div className="computed-box-label">Cashflow after saving</div>
          <div className="computed-box-value mono">{fmtExact(amountIdr + month.passive_income_idr - month.total_expenses_idr)}</div>
        </div>
      </div>}
    </Modal>
  )
}

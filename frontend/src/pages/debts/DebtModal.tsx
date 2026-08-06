import { useEffect, useMemo, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateDebtEntry, useDeleteDebtEntry, useUpdateDebtEntry } from '../../hooks/useDebtSnapshots'
import {
  useCreateReceivableLoan,
  useDeleteReceivableLoan,
  useReceivableLoans,
  useUpdateReceivableLoan,
} from '../../hooks/useReceivableLoans'
import { fmtIdrExact, parseNumeric } from '../../lib/format'
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
  const deleteEntry = useDeleteDebtEntry()
  const { data: loans = [] } = useReceivableLoans()
  const createLoan = useCreateReceivableLoan()
  const updateLoan = useUpdateReceivableLoan()
  const deleteLoan = useDeleteReceivableLoan()
  const types = direction === 'i_owe' ? DEBT_TYPES : RECEIVABLE_TYPES

  const [name, setName] = useState('')
  const [type, setType] = useState(types[0])
  const [amount, setAmount] = useState('')

  // Loan terms — shown only for receivables. Linked to the debt entry by
  // matching borrower name, not a foreign key, the same relationship a
  // bond's name has to its Assets holding.
  const [startDate, setStartDate] = useState('')
  const [termMonths, setTermMonths] = useState('')
  const [interest, setInterest] = useState('')

  const existingLoan = useMemo(
    () => loans.find((l) => l.borrower_name.toLowerCase() === name.trim().toLowerCase()) ?? null,
    [loans, name],
  )

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
    const matchingName = editingEntry?.name ?? ''
    const loan = loans.find((l) => l.borrower_name.toLowerCase() === matchingName.toLowerCase())
    if (loan) {
      setStartDate(loan.start_date)
      setTermMonths(String(loan.term_months))
      setInterest(String(loan.interest_idr))
    } else {
      setStartDate('')
      setTermMonths('')
      setInterest('')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, editingEntry])

  const isDebt = direction === 'i_owe'
  const title = isDebt ? (editingEntry ? 'Edit debt' : 'Add debt') : editingEntry ? 'Edit receivable' : 'Add receivable'
  const subtitle = isDebt ? 'Something you owe.' : 'Someone who owes you.'
  const nameLabel = isDebt ? 'Debt name' : 'Person / name'
  const namePlaceholder = isDebt ? 'e.g. OCBC KPA' : 'e.g. Edo Tole'
  const cta = isDebt ? (editingEntry ? 'Save changes' : 'Add debt') : editingEntry ? 'Save changes' : 'Add receivable'

  const interestNum = parseNumeric(interest)
  const termMonthsNum = Math.round(parseNumeric(termMonths))
  const termFilledIn = startDate !== '' && termMonthsNum > 0 && interestNum > 0
  const monthlyPreviewIdr = termMonthsNum > 0 ? Math.round(interestNum / termMonthsNum) : 0

  // Saves the loan-terms fields against whatever the receivable is named
  // once the debt entry itself is saved — creating a loan if terms were
  // filled in and none existed, updating the matching one if there is one,
  // or deleting it if the fields were cleared back out. A no-op for `i_owe`
  // debts and for a receivable with no terms and no existing loan.
  async function saveLoanTerms() {
    if (isDebt) return
    if (termFilledIn) {
      const loanInput = {
        borrower_name: name.trim(),
        start_date: startDate,
        term_months: termMonthsNum,
        interest_idr: Math.round(interestNum),
        note: '',
      }
      if (existingLoan) {
        await updateLoan.mutateAsync({ id: existingLoan.id, input: loanInput })
      } else {
        await createLoan.mutateAsync(loanInput)
      }
    } else if (existingLoan) {
      await deleteLoan.mutateAsync(existingLoan.id)
    }
  }

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
      await saveLoanTerms()
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
      showSuccess('Deleted.')
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving =
    createEntry.isPending || updateEntry.isPending || createLoan.isPending || updateLoan.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      subtitle={subtitle}
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
        Amount (Rp)
        <input
          className="field-input mono"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder={isDebt ? '8800000' : '4800000'}
        />
      </label>

      {!isDebt && (
        <>
          <div className="card-title">Loan terms (optional)</div>
          <div className="field-row">
            <label className="field">
              Start date
              <input
                type="date"
                className="field-input"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </label>
            <label className="field">
              Term (months)
              <input
                className="field-input mono"
                value={termMonths}
                onChange={(e) => setTermMonths(e.target.value)}
                placeholder="12"
              />
            </label>
          </div>
          <label className="field">
            Interest debt (Rp)
            <input
              className="field-input mono"
              value={interest}
              onChange={(e) => setInterest(e.target.value)}
              placeholder="200000"
            />
          </label>

          {termFilledIn && (
            <div className="computed-box">
              <div>
                <div className="computed-box-label">Monthly amount</div>
                <div className="computed-box-value mono">{fmtIdrExact(monthlyPreviewIdr)}</div>
              </div>
              <div className="computed-box-note">
                {fmtIdrExact(Math.round(interestNum))} total interest over {termMonthsNum} months, counted as passive
                income
              </div>
            </div>
          )}
        </>
      )}
    </Modal>
  )
}

import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateExpensePeriod } from '../../hooks/useExpensePeriods'

interface NewPeriodModalProps {
  open: boolean
  onClose: () => void
  hasExistingPeriod: boolean
  latestPeriodEndDate: string | undefined
  onCreated: (id: string) => void
}

const YEARS = [2026, 2027, 2028]
const MONTHS = [
  { value: 1, label: 'January' },
  { value: 2, label: 'February' },
  { value: 3, label: 'March' },
  { value: 4, label: 'April' },
  { value: 5, label: 'May' },
  { value: 6, label: 'June' },
  { value: 7, label: 'July' },
  { value: 8, label: 'August' },
  { value: 9, label: 'September' },
  { value: 10, label: 'October' },
  { value: 11, label: 'November' },
  { value: 12, label: 'December' },
]

/** Defaults the picker to the month after the current latest period (a
 * sensible starting point for "add the next one"), clamped into the
 * supported Jan 2026 - Dec 2028 range, or to today's month if there's no
 * latest period yet. */
function defaultYearMonth(latestEndDate: string | undefined): { year: number; month: number } {
  let year: number
  let month: number
  if (latestEndDate) {
    const [y, m] = latestEndDate.split('-').map(Number)
    month = m + 1
    year = y
    if (month > 12) {
      month = 1
      year += 1
    }
  } else {
    const now = new Date()
    year = now.getFullYear()
    month = now.getMonth() + 1
  }
  if (year < YEARS[0]) return { year: YEARS[0], month: 1 }
  if (year > YEARS[YEARS.length - 1]) return { year: YEARS[YEARS.length - 1], month: 12 }
  return { year, month }
}

export function NewPeriodModal({
  open,
  onClose,
  hasExistingPeriod,
  latestPeriodEndDate,
  onCreated,
}: NewPeriodModalProps) {
  const { showError, showSuccess } = useToast()
  const createPeriod = useCreateExpensePeriod()
  const [copyEnvelopes, setCopyEnvelopes] = useState(true)
  const [year, setYear] = useState(YEARS[0])
  const [month, setMonth] = useState(1)

  useEffect(() => {
    if (!open) return
    const d = defaultYearMonth(latestPeriodEndDate)
    setYear(d.year)
    setMonth(d.month)
  }, [open, latestPeriodEndDate])

  async function handleCreate() {
    try {
      const period = await createPeriod.mutateAsync({ year, month, copy_envelopes: copyEnvelopes })
      showSuccess('Period created.')
      onCreated(period.id)
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create expense period"
      subtitle="Pick which pay-cycle period to create (25th of the prior month through the 24th)."
      footer={
        <>
          <ModalCancelButton onClick={onClose} />
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleCreate}
            disabled={createPeriod.isPending}
          >
            Create period
          </button>
        </>
      }
    >
      <div className="field-row">
        <label className="field">
          Month
          <select className="field-input" value={month} onChange={(e) => setMonth(Number(e.target.value))}>
            {MONTHS.map((m) => (
              <option key={m.value} value={m.value}>
                {m.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Year
          <select className="field-input" value={year} onChange={(e) => setYear(Number(e.target.value))}>
            {YEARS.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </label>
      </div>

      {hasExistingPeriod && (
        <label className="field-checkbox">
          <input
            type="checkbox"
            checked={copyEnvelopes}
            onChange={(e) => setCopyEnvelopes(e.target.checked)}
          />
          Copy budget envelope names &amp; targets from the current latest period
        </label>
      )}
    </Modal>
  )
}

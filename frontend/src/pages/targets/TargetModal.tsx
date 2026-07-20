import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateTarget, useDeleteTarget, useUpdateTarget } from '../../hooks/useTargets'
import { parseNumeric } from '../../lib/format'
import type { Target, TargetMetricType } from '../../types'

interface TargetModalProps {
  open: boolean
  onClose: () => void
  editingTarget: Target | null
}

const METRIC_OPTIONS: { value: TargetMetricType; label: string; unit: string }[] = [
  { value: 'equity', label: 'Total equity', unit: '' },
  { value: 'gold_grams', label: 'Gold (grams)', unit: 'g' },
  { value: 'passive_income', label: 'Passive income / yr', unit: '' },
  { value: 'debt_ratio', label: 'Debt-to-equity ratio', unit: '%' },
  { value: 'custom', label: 'Custom', unit: '' },
]

const CURRENT_YEAR = new Date().getFullYear()

export function TargetModal({ open, onClose, editingTarget }: TargetModalProps) {
  const { showError, showSuccess } = useToast()
  const createTarget = useCreateTarget()
  const updateTarget = useUpdateTarget()
  const deleteTarget = useDeleteTarget()

  const [name, setName] = useState('')
  const [metricType, setMetricType] = useState<TargetMetricType>('custom')
  const [year, setYear] = useState(String(CURRENT_YEAR))
  const [targetValue, setTargetValue] = useState('')
  const [unit, setUnit] = useState('')
  const [manualCurrent, setManualCurrent] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingTarget) {
      setName(editingTarget.name)
      setMetricType(editingTarget.metric_type)
      setYear(String(editingTarget.year))
      setTargetValue(String(editingTarget.target_value))
      setUnit(editingTarget.unit)
      setManualCurrent(String(editingTarget.current_value))
    } else {
      setName('')
      setMetricType('custom')
      setYear(String(CURRENT_YEAR))
      setTargetValue('')
      setUnit('')
      setManualCurrent('')
    }
  }, [open, editingTarget])

  async function handleSave() {
    if (!name.trim()) {
      showError('Give the target a name.')
      return
    }
    const input = {
      name: name.trim(),
      year: Math.round(parseNumeric(year)) || CURRENT_YEAR,
      metric_type: metricType,
      target_value: parseNumeric(targetValue),
      unit,
      manual_current_value: metricType === 'custom' ? parseNumeric(manualCurrent) : undefined,
    }
    try {
      if (editingTarget) {
        await updateTarget.mutateAsync({ id: editingTarget.id, input })
        showSuccess('Target updated.')
      } else {
        await createTarget.mutateAsync(input)
        showSuccess('Target added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  async function handleDelete() {
    if (!editingTarget) return
    try {
      await deleteTarget.mutateAsync(editingTarget.id)
      showSuccess('Target removed.')
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createTarget.isPending || updateTarget.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Add / edit target"
      subtitle="Set a year-end goal to track progress toward."
      footer={
        <>
          {editingTarget && (
            <button
              type="button"
              className="btn btn-secondary"
              style={{ color: 'var(--red)', marginRight: 'auto' }}
              onClick={handleDelete}
              disabled={deleteTarget.isPending}
            >
              Delete
            </button>
          )}
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            Save target
          </button>
        </>
      }
    >
      <label className="field">
        Target name
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Emas 400 gram"
        />
      </label>
      <div className="field-row">
        <label className="field">
          Metric
          <select
            className="field-input"
            value={metricType}
            onChange={(e) => {
              const mt = e.target.value as TargetMetricType
              setMetricType(mt)
              const opt = METRIC_OPTIONS.find((o) => o.value === mt)
              if (opt) setUnit(opt.unit)
            }}
          >
            {METRIC_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Year
          <input className="field-input mono" value={year} onChange={(e) => setYear(e.target.value)} placeholder="2026" />
        </label>
      </div>
      <div className="field-row">
        <label className="field">
          Target value
          <input
            className="field-input mono"
            value={targetValue}
            onChange={(e) => setTargetValue(e.target.value)}
            placeholder="400"
          />
        </label>
        <label className="field">
          Unit (optional)
          <input className="field-input" value={unit} onChange={(e) => setUnit(e.target.value)} placeholder="g" />
        </label>
      </div>
      {metricType === 'custom' && (
        <label className="field">
          Current value
          <input
            className="field-input mono"
            value={manualCurrent}
            onChange={(e) => setManualCurrent(e.target.value)}
            placeholder="381.08"
          />
        </label>
      )}
    </Modal>
  )
}

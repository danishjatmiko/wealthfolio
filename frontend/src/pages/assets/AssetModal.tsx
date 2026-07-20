import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import { useCreateHolding, useUpdateHolding } from '../../hooks/useHoldings'
import {
  GOLD_TYPES,
  buildDetail,
  computeHoldingValue,
  isCashCategory,
  isGoldCategory,
  prefillFromHolding,
  showsComputedBox,
  showsIdrInput,
  showsUsdInput,
} from '../../lib/holdingCalc'
import { fmtIdr, goldFmt, usdFmt, parseNumeric } from '../../lib/format'
import type { AssetFormValues } from '../../lib/holdingCalc'
import type { Category, Holding, RateEntry } from '../../types'

interface AssetModalProps {
  open: boolean
  onClose: () => void
  snapshotDate: string
  categories: Category[]
  latestRate: RateEntry | undefined
  editingHolding: Holding | null
  defaultCategoryId: number
}

const EMPTY: AssetFormValues = {
  categoryLabel: '',
  val: '',
  gram: '',
  qty: '',
  usd: '',
  currency: 'IDR',
  brand: 'Antam',
}

export function AssetModal({
  open,
  onClose,
  snapshotDate,
  categories,
  latestRate,
  editingHolding,
  defaultCategoryId,
}: AssetModalProps) {
  const { showError, showSuccess } = useToast()
  const createHolding = useCreateHolding()
  const updateHolding = useUpdateHolding()

  const [name, setName] = useState('')
  const [categoryId, setCategoryId] = useState<number>(0)
  const [form, setForm] = useState<AssetFormValues>(EMPTY)

  useEffect(() => {
    if (!open) return
    if (editingHolding) {
      setName(editingHolding.name)
      setCategoryId(editingHolding.category_id)
      setForm(
        prefillFromHolding({
          category_label: editingHolding.category_label,
          detail: editingHolding.detail,
          gram: editingHolding.gram,
          qty: editingHolding.qty,
          usd_value: editingHolding.usd_value,
          currency: editingHolding.currency,
          brand: editingHolding.brand,
          name: editingHolding.name,
          value_idr: editingHolding.value_idr,
        }),
      )
    } else {
      const cat = categories.find((c) => c.id === defaultCategoryId) ?? categories[0]
      setName('')
      setCategoryId(cat?.id ?? 0)
      setForm({ ...EMPTY, categoryLabel: cat?.label ?? '' })
    }
  }, [open, editingHolding, defaultCategoryId, categories])

  const selectedCategory = categories.find((c) => c.id === categoryId)
  const categoryLabel = selectedCategory?.label ?? form.categoryLabel

  const activeForm: AssetFormValues = { ...form, categoryLabel }
  const computedValue = computeHoldingValue(activeForm, latestRate)
  const isEdit = !!editingHolding

  const gold = isGoldCategory(categoryLabel)
  const cash = isCashCategory(categoryLabel)
  const showUsd = showsUsdInput(categoryLabel, form.currency)
  const showIdr = showsIdrInput(categoryLabel, form.currency)
  const showComputed = showsComputedBox(categoryLabel, form.currency)

  const computedNote = gold
    ? `Auto from ${form.brand} · ${goldFmt(
        form.brand === 'King Halim' ? latestRate?.kinghalim ?? 0 : form.brand === 'UBS' ? latestRate?.ubs ?? 0 : latestRate?.antam ?? 0,
      )}`
    : `Auto from USD rate · ${usdFmt(latestRate?.usd_idr ?? 0)}`

  async function handleSave() {
    if (!categoryId) {
      showError('Choose a category first.')
      return
    }
    const detail = buildDetail(activeForm)
    const value = Math.round(computedValue)
    const input = {
      category_id: categoryId,
      name: name || 'New asset',
      detail: detail || null,
      value_idr: value,
      gram: gold ? parseNumeric(form.gram) : null,
      qty: gold ? parseNumeric(form.qty) || 1 : null,
      brand: gold ? form.brand : null,
      usd_value: showUsd ? parseNumeric(form.usd) : null,
      currency: cash ? form.currency : null,
    }
    try {
      if (isEdit && editingHolding) {
        await updateHolding.mutateAsync({ id: editingHolding.id, input })
        showSuccess('Asset updated.')
      } else {
        await createHolding.mutateAsync({ date: snapshotDate, input })
        showSuccess('Asset added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createHolding.isPending || updateHolding.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Edit asset' : 'Add asset'}
      subtitle={`to snapshot · ${snapshotDate}`}
      footer={
        <>
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {isEdit ? 'Save changes' : 'Add asset'}
          </button>
        </>
      }
    >
      <label className="field">
        Asset name
        <input
          className="field-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Antam 5g"
        />
      </label>

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

      {gold && (
        <div className="field-row">
          <label className="field">
            Type
            <select
              className="field-input"
              value={form.brand}
              onChange={(e) => setForm((f) => ({ ...f, brand: e.target.value }))}
            >
              {GOLD_TYPES.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            Gram / pc
            <input
              className="field-input mono"
              value={form.gram}
              onChange={(e) => setForm((f) => ({ ...f, gram: e.target.value }))}
              placeholder="5"
            />
          </label>
          <label className="field">
            Quantity
            <input
              className="field-input mono"
              value={form.qty}
              onChange={(e) => setForm((f) => ({ ...f, qty: e.target.value }))}
              placeholder="1"
            />
          </label>
        </div>
      )}

      {cash && (
        <label className="field">
          Currency
          <select
            className="field-input"
            value={form.currency}
            onChange={(e) => setForm((f) => ({ ...f, currency: e.target.value as 'IDR' | 'USD' }))}
          >
            <option value="IDR">IDR</option>
            <option value="USD">USD</option>
          </select>
        </label>
      )}

      {showUsd && (
        <label className="field">
          Value (USD)
          <input
            className="field-input mono"
            value={form.usd}
            onChange={(e) => setForm((f) => ({ ...f, usd: e.target.value }))}
            placeholder="5640"
          />
        </label>
      )}

      {showIdr && (
        <label className="field">
          Value (rb Rp)
          <input
            className="field-input mono"
            value={form.val}
            onChange={(e) => setForm((f) => ({ ...f, val: e.target.value }))}
            placeholder="21200"
          />
        </label>
      )}

      {showComputed && (
        <div className="computed-box">
          <div>
            <div className="computed-box-label">IDR value (auto)</div>
            <div className="computed-box-value mono">{fmtIdr(computedValue)}</div>
          </div>
          <div className="computed-box-note">{computedNote}</div>
        </div>
      )}
    </Modal>
  )
}

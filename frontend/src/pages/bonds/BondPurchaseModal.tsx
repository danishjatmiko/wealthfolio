import { useEffect, useState } from 'react'
import { Modal, ModalCancelButton } from '../../components/Modal'
import { errorMessage, useToast } from '../../context/ToastContext'
import {
  useCreateBondPurchase,
  useDeleteBondPurchase,
  useUpdateBondPurchase,
} from '../../hooks/useBondPurchases'
import { useLatestRate } from '../../hooks/useRates'
import {
  BOND_PLATFORMS,
  bondTotalIdr,
  bondTotalUsd,
  couponPerCycleUsd,
  couponScheduleLabel,
} from '../../lib/bondCalc'
import { fmtIdrExact, fmtUsd, fmtUsdIdrRate, parseNumeric } from '../../lib/format'
import type { BondPurchase } from '../../types'

interface BondPurchaseModalProps {
  open: boolean
  onClose: () => void
  editingPurchase: BondPurchase | null
  /** Pre-fills the name when adding another lot of a bond already held. */
  defaultBondName?: string
}

const OTHER_PLATFORM = '__other__'

export function BondPurchaseModal({
  open,
  onClose,
  editingPurchase,
  defaultBondName,
}: BondPurchaseModalProps) {
  const { showError, showSuccess } = useToast()
  const createPurchase = useCreateBondPurchase()
  const updatePurchase = useUpdateBondPurchase()
  const deletePurchase = useDeleteBondPurchase()
  const { data: latestRate } = useLatestRate()

  const [bondName, setBondName] = useState('')
  const [platform, setPlatform] = useState<string>(BOND_PLATFORMS[0])
  const [otherPlatform, setOtherPlatform] = useState('')
  const [interestRate, setInterestRate] = useState('')
  const [buyDate, setBuyDate] = useState('')
  const [maturityDate, setMaturityDate] = useState('')
  const [quantity, setQuantity] = useState('1')
  const [faceValue, setFaceValue] = useState('1000')
  const [price, setPrice] = useState('')
  const [accrued, setAccrued] = useState('')
  const [usdIdr, setUsdIdr] = useState('')

  useEffect(() => {
    if (!open) return
    if (editingPurchase) {
      const known = (BOND_PLATFORMS as readonly string[]).includes(editingPurchase.platform)
      setBondName(editingPurchase.bond_name)
      setPlatform(known ? editingPurchase.platform : OTHER_PLATFORM)
      setOtherPlatform(known ? '' : editingPurchase.platform)
      setInterestRate(String(editingPurchase.interest_rate))
      setBuyDate(editingPurchase.buy_date)
      setMaturityDate(editingPurchase.maturity_date)
      setQuantity(String(editingPurchase.quantity))
      setFaceValue(String(editingPurchase.face_value_usd))
      setPrice(String(editingPurchase.price_usd))
      setAccrued(String(editingPurchase.accrued_interest_usd))
      setUsdIdr(String(editingPurchase.usd_idr_at_purchase))
    } else {
      setBondName(defaultBondName ?? '')
      setPlatform(BOND_PLATFORMS[0])
      setOtherPlatform('')
      setInterestRate('')
      setBuyDate(new Date().toISOString().slice(0, 10))
      setMaturityDate('')
      setQuantity('1')
      setFaceValue('1000')
      setPrice('')
      setAccrued('')
      // A purchase logged today almost always settled at today's rate —
      // prefilling saves retyping it. Only on add: an edit must keep the
      // historical rate it was actually bought at.
      setUsdIdr(latestRate ? String(latestRate.usd_idr) : '')
    }
  }, [open, editingPurchase, defaultBondName, latestRate])

  const resolvedPlatform = platform === OTHER_PLATFORM ? otherPlatform.trim() : platform
  const totalUsd = bondTotalUsd(parseNumeric(price), parseNumeric(accrued))
  const totalIdr = bondTotalIdr(totalUsd, parseNumeric(usdIdr))
  const perCycle = couponPerCycleUsd(
    parseNumeric(faceValue),
    parseNumeric(quantity),
    parseNumeric(interestRate),
  )
  const schedule = couponScheduleLabel(maturityDate)

  async function handleSave() {
    if (!bondName.trim()) {
      showError('Give the bond a name.')
      return
    }
    if (!buyDate || !maturityDate) {
      showError('Both the buy date and maturity date are required.')
      return
    }
    const input = {
      bond_name: bondName.trim(),
      platform: resolvedPlatform,
      interest_rate: parseNumeric(interestRate),
      buy_date: buyDate,
      maturity_date: maturityDate,
      quantity: parseNumeric(quantity),
      face_value_usd: parseNumeric(faceValue),
      price_usd: parseNumeric(price),
      accrued_interest_usd: parseNumeric(accrued),
      usd_idr_at_purchase: parseNumeric(usdIdr),
    }
    try {
      if (editingPurchase) {
        await updatePurchase.mutateAsync({ id: editingPurchase.id, input })
        showSuccess('Purchase updated.')
      } else {
        await createPurchase.mutateAsync(input)
        showSuccess('Purchase added.')
      }
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  async function handleDelete() {
    if (!editingPurchase) return
    if (
      !window.confirm(
        `Delete this ${editingPurchase.bond_name} purchase from ${editingPurchase.buy_date}? This cannot be undone.`,
      )
    )
      return
    try {
      await deletePurchase.mutateAsync(editingPurchase.id)
      showSuccess('Purchase deleted.')
      onClose()
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  const saving = createPurchase.isPending || updatePurchase.isPending

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editingPurchase ? 'Edit purchase' : 'Add bond purchase'}
      subtitle="One buy, exactly as the broker reported it."
      footer={
        <>
          {editingPurchase && (
            <button
              type="button"
              className="btn btn-danger"
              style={{ marginRight: 'auto' }}
              onClick={handleDelete}
              disabled={deletePurchase.isPending}
            >
              🗑 Delete
            </button>
          )}
          <ModalCancelButton onClick={onClose} />
          <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>
            {editingPurchase ? 'Save changes' : 'Add purchase'}
          </button>
        </>
      }
    >
      <label className="field">
        Bond name
        <input
          className="field-input"
          value={bondName}
          onChange={(e) => setBondName(e.target.value)}
          placeholder="e.g. INDON36NEWNEW"
        />
      </label>

      <div className="field-row">
        <label className="field">
          Platform
          <select
            className="field-input"
            value={platform}
            onChange={(e) => setPlatform(e.target.value)}
          >
            {BOND_PLATFORMS.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
            <option value={OTHER_PLATFORM}>Other…</option>
          </select>
        </label>
        <label className="field">
          Interest rate (%)
          <input
            className="field-input mono"
            value={interestRate}
            onChange={(e) => setInterestRate(e.target.value)}
            placeholder="5.69"
          />
        </label>
      </div>

      {platform === OTHER_PLATFORM && (
        <label className="field">
          Platform name
          <input
            className="field-input"
            value={otherPlatform}
            onChange={(e) => setOtherPlatform(e.target.value)}
            placeholder="e.g. Bibit"
          />
        </label>
      )}

      <div className="field-row">
        <label className="field">
          Buy date
          <input
            type="date"
            className="field-input"
            value={buyDate}
            onChange={(e) => setBuyDate(e.target.value)}
          />
        </label>
        <label className="field">
          Maturity date
          <input
            type="date"
            className="field-input"
            value={maturityDate}
            onChange={(e) => setMaturityDate(e.target.value)}
          />
        </label>
      </div>

      <div className="field-row">
        <label className="field">
          Quantity
          <input
            className="field-input mono"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="1"
          />
        </label>
        <label className="field">
          Face value ($ / lot)
          <input
            className="field-input mono"
            value={faceValue}
            onChange={(e) => setFaceValue(e.target.value)}
            placeholder="1000"
          />
        </label>
      </div>

      <div className="field-row">
        <label className="field">
          Price ($, whole lot)
          <input
            className="field-input mono"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="1014.50"
          />
        </label>
        <label className="field">
          Accrued interest ($)
          <input
            className="field-input mono"
            value={accrued}
            onChange={(e) => setAccrued(e.target.value)}
            placeholder="0.47"
          />
        </label>
      </div>

      <label className="field">
        IDR rate when bought
        <input
          className="field-input mono"
          value={usdIdr}
          onChange={(e) => setUsdIdr(e.target.value)}
          placeholder="17690"
        />
      </label>

      <div className="computed-box">
        <div>
          <div className="computed-box-label">Total paid</div>
          <div className="computed-box-value mono">{fmtUsd(totalUsd)}</div>
        </div>
        <div className="computed-box-note">
          {fmtIdrExact(totalIdr)}
          {parseNumeric(usdIdr) > 0 && <> @ {fmtUsdIdrRate(parseNumeric(usdIdr))}</>}
        </div>
      </div>

      <div className="computed-box">
        <div>
          <div className="computed-box-label">Coupon per cycle</div>
          <div className="computed-box-value mono">{fmtUsd(perCycle)}</div>
        </div>
        <div className="computed-box-note">2 × / year{schedule && <> · {schedule}</>}</div>
      </div>
    </Modal>
  )
}

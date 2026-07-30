import { useState } from 'react'
import { useCreateRate, useLatestRate, useRates } from '../../hooks/useRates'
import { goldFmt, goldFmtShort, parseNumeric, fmtUsdIdrRate } from '../../lib/format'
import { errorMessage, useToast } from '../../context/ToastContext'
import './Rates.css'

function todayIso(): string {
  return new Date().toISOString().slice(0, 10)
}

export function Rates() {
  const { data: latest } = useLatestRate()
  const { data: history = [] } = useRates()
  const createRate = useCreateRate()
  const { showError, showSuccess } = useToast()

  const [date, setDate] = useState(todayIso())
  const [antam, setAntam] = useState('')
  const [kinghalim, setKinghalim] = useState('')
  const [ubs, setUbs] = useState('')
  const [usdIdr, setUsdIdr] = useState('')

  async function handleSave() {
    if (!date) {
      showError('Pick a date.')
      return
    }
    try {
      await createRate.mutateAsync({
        entry_date: date,
        antam: parseNumeric(antam),
        kinghalim: parseNumeric(kinghalim),
        ubs: parseNumeric(ubs),
        usd_idr: parseNumeric(usdIdr),
      })
      showSuccess('Price logged.')
      setAntam('')
      setKinghalim('')
      setUbs('')
      setUsdIdr('')
    } catch (err) {
      showError(errorMessage(err))
    }
  }

  return (
    <div>
      <div className="rates-summary-grid">
        <div className="card rate-summary-card">
          <div className="rate-summary-label">Antam</div>
          <div className="mono rate-summary-value">{latest ? goldFmtShort(latest.antam) : '—'}</div>
        </div>
        <div className="card rate-summary-card">
          <div className="rate-summary-label">King Halim</div>
          <div className="mono rate-summary-value">{latest ? goldFmtShort(latest.kinghalim) : '—'}</div>
        </div>
        <div className="card rate-summary-card">
          <div className="rate-summary-label">UBS</div>
          <div className="mono rate-summary-value">{latest ? goldFmtShort(latest.ubs) : '—'}</div>
        </div>
        <div className="card rate-summary-card rate-summary-card-usd">
          <div className="rate-summary-label">USD → IDR</div>
          <div className="mono rate-summary-value">{latest ? fmtUsdIdrRate(latest.usd_idr) : '—'}</div>
        </div>
      </div>

      <div className="card rate-form-card">
        <div className="card-title-inline">Log a new price by date</div>
        <div className="rate-form-row">
          <label className="field">
            Date
            <input
              type="date"
              className="field-input"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </label>
          <label className="field">
            Antam (Rp/g)
            <input
              className="field-input mono"
              value={antam}
              onChange={(e) => setAntam(e.target.value)}
              placeholder="1650000"
            />
          </label>
          <label className="field">
            King Halim (Rp/g)
            <input
              className="field-input mono"
              value={kinghalim}
              onChange={(e) => setKinghalim(e.target.value)}
              placeholder="1610000"
            />
          </label>
          <label className="field">
            UBS (Rp/g)
            <input
              className="field-input mono"
              value={ubs}
              onChange={(e) => setUbs(e.target.value)}
              placeholder="1600000"
            />
          </label>
          <label className="field">
            USD/IDR
            <input
              className="field-input mono"
              value={usdIdr}
              onChange={(e) => setUsdIdr(e.target.value)}
              placeholder="15850"
            />
          </label>
          <button
            type="button"
            className="btn btn-primary rate-form-save"
            onClick={handleSave}
            disabled={createRate.isPending}
          >
            Save
          </button>
        </div>
      </div>

      <div className="card rate-history-card">
        <div className="card-title-inline">History</div>
        <div className="rate-history-table">
          <div className="rate-history-row rate-history-head">
            <span>Date</span>
            <span>Antam</span>
            <span>King Halim</span>
            <span>UBS</span>
            <span>USD/IDR</span>
          </div>
          {history.length === 0 && <div className="empty-state">No price history yet.</div>}
          {history.map((r) => (
            <div className="rate-history-row" key={r.id}>
              <span className="mono">{r.entry_date}</span>
              <span className="mono">{goldFmt(r.antam)}</span>
              <span className="mono">{goldFmt(r.kinghalim)}</span>
              <span className="mono">{goldFmt(r.ubs)}</span>
              <span className="mono">{fmtUsdIdrRate(r.usd_idr)}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

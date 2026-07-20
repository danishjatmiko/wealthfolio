import { useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useProgress } from '../../hooks/useProgress'
import { LineChart } from '../../components/charts/LineChart'
import { ACCENT } from '../../lib/colors'
import type { ProgressGranularity } from '../../types'
import './Progress.css'

const VIEWS: { id: ProgressGranularity; label: string }[] = [
  { id: 'monthly', label: 'Monthly' },
  { id: 'quarterly', label: 'Quarterly' },
  { id: 'yearly', label: 'Yearly' },
]

function plusPrefix(n: number): string {
  return n > 0 ? '+' : ''
}

export function Progress() {
  const { fmt } = useMoney()
  const [granularity, setGranularity] = useState<ProgressGranularity>('monthly')
  const { data, isLoading, isError } = useProgress(granularity)

  return (
    <div>
      <div className="progress-header">
        <div>
          <div className="mono progress-latest">{data ? fmt(data.latest_value_idr) : '—'}</div>
          {data && (
            <div className="progress-delta">
              {plusPrefix(data.delta_idr)}
              {fmt(data.delta_idr)} ({plusPrefix(data.delta_pct)}
              {data.delta_pct.toFixed(2)}%) vs previous period
            </div>
          )}
        </div>
        <div className="segmented">
          {VIEWS.map((v) => (
            <button
              key={v.id}
              type="button"
              className={'segmented-btn' + (granularity === v.id ? ' segmented-btn-active' : '')}
              onClick={() => setGranularity(v.id)}
            >
              {v.label}
            </button>
          ))}
        </div>
      </div>

      <div className="card progress-chart-card">
        <div className="card-title">Asset value trend</div>
        {isLoading && <div className="empty-state">Loading…</div>}
        {isError && <div className="empty-state">Couldn't load progress data.</div>}
        {data && (
          <>
            <LineChart
              series={data.series.map((p) => ({ label: p.label, value: p.net_equity_idr }))}
              color={ACCENT}
              formatValue={fmt}
            />
            <div className="progress-labels">
              {data.series.map((p) => (
                <span className="progress-label" key={p.date}>
                  {p.label}
                </span>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

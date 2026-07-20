import { useId, useState } from 'react'

// Ported from the prototype's buildLine() (Portfolio App.dc.html ~line 690):
// polyline + gradient area fill, last point marked, ~18% vertical padding.
// Extended with a left Y-axis and a hover tooltip showing the point's value.
export interface LinePoint {
  label: string
  value: number
}

interface LineChartProps {
  series: LinePoint[]
  color: string
  height?: number
  formatValue: (v: number) => string
}

const W = 600
const AXIS_W = 56
const TICK_COUNT = 4

export function LineChart({ series, color, height = 200, formatValue }: LineChartProps) {
  const rawId = useId()
  const gradientId = `grad-${rawId.replace(/[^a-zA-Z0-9]/g, '')}`
  const H = height
  const [hoverIndex, setHoverIndex] = useState<number | null>(null)

  if (series.length === 0) {
    return <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} style={{ display: 'block' }} />
  }

  const vals = series.map((p) => p.value)
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const pad = (max - min) * 0.18 || 1
  const lo = min - pad
  const hi = max + pad

  const plotX0 = AXIS_W
  const plotW = W - AXIS_W
  const X = (i: number) => plotX0 + (series.length > 1 ? (i / (series.length - 1)) * plotW : 0)
  const Y = (v: number) => H - ((v - lo) / (hi - lo)) * H

  const pts = series.map((p, i): [number, number] => [X(i), Y(p.value)])
  const linePoints = pts.map((p) => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' ')
  const areaPoints = `${plotX0},${H} ${linePoints} ${W},${H}`

  const ticks = Array.from({ length: TICK_COUNT }, (_, i) => lo + (hi - lo) * (i / (TICK_COUNT - 1)))

  const hovered = hoverIndex !== null ? { point: series[hoverIndex], xy: pts[hoverIndex] } : null
  const tooltipLeftPct = hovered ? (hovered.xy[0] / W) * 100 : 0
  const tooltipAbove = hovered ? hovered.xy[1] > 40 : false

  return (
    <div style={{ position: 'relative', width: '100%', height: H }}>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        width="100%"
        height={H}
        preserveAspectRatio="none"
        style={{ display: 'block', overflow: 'visible' }}
      >
        <defs>
          <linearGradient id={gradientId} x1={0} y1={0} x2={0} y2={1}>
            <stop offset="0%" stopColor={color} stopOpacity={0.2} />
            <stop offset="100%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>

        {ticks.map((t, i) => (
          <text
            key={i}
            x={AXIS_W - 8}
            y={Y(t)}
            textAnchor="end"
            dominantBaseline={i === 0 ? 'auto' : i === TICK_COUNT - 1 ? 'hanging' : 'middle'}
            fontSize={10}
            fill="var(--text-faint)"
          >
            {formatValue(t)}
          </text>
        ))}

        <polygon points={areaPoints} fill={`url(#${gradientId})`} />
        <polyline
          points={linePoints}
          fill="none"
          stroke={color}
          strokeWidth={2.5}
          vectorEffect="non-scaling-stroke"
          strokeLinejoin="round"
          strokeLinecap="round"
        />

        {hovered && (
          <line
            x1={hovered.xy[0]}
            y1={0}
            x2={hovered.xy[0]}
            y2={H}
            stroke={color}
            strokeOpacity={0.25}
            strokeWidth={1}
            strokeDasharray="3 3"
          />
        )}

        {pts.map((p, i) => (
          <circle
            key={i}
            cx={p[0]}
            cy={p[1]}
            r={i === hoverIndex ? 5.5 : i === pts.length - 1 ? 4.5 : 0}
            fill={color}
            stroke="#fff"
            strokeWidth={2}
          />
        ))}

        {/* Invisible wider hit-targets for hover, drawn last so they're on top. */}
        {pts.map((p, i) => (
          <circle
            key={`hit-${i}`}
            cx={p[0]}
            cy={p[1]}
            r={14}
            fill="transparent"
            onMouseEnter={() => setHoverIndex(i)}
            onMouseLeave={() => setHoverIndex((cur) => (cur === i ? null : cur))}
          />
        ))}
      </svg>

      {hovered && (
        <div
          className="line-chart-tooltip"
          style={{
            left: `${tooltipLeftPct}%`,
            top: tooltipAbove ? hovered.xy[1] - 10 : hovered.xy[1] + 10,
            transform: `translate(-50%, ${tooltipAbove ? '-100%' : '0'})`,
          }}
        >
          <div className="line-chart-tooltip-label">{hovered.point.label}</div>
          <div className="line-chart-tooltip-value mono">{formatValue(hovered.point.value)}</div>
        </div>
      )}
    </div>
  )
}

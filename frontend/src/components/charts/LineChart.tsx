import { useId, useLayoutEffect, useRef, useState } from 'react'

// Ported from the prototype's buildLine() (Portfolio App.dc.html ~line 690):
// polyline + gradient area fill, last point marked, ~18% vertical padding.
// Extended with a left Y-axis, a hover tooltip showing the point's value,
// and an optional secondary series (dashed, own right-side axis) for
// comparing two differently-scaled metrics — e.g. a Rupiah value against a
// percentage — on the same timeline.
export interface LinePoint {
  label: string
  value: number
}

interface LineChartProps {
  series: LinePoint[]
  color: string
  height?: number
  formatValue: (v: number) => string
  secondarySeries?: LinePoint[]
  secondaryColor?: string
  secondaryFormatValue?: (v: number) => string
}

const DEFAULT_W = 600
const AXIS_W = 56
const AXIS_W_RIGHT = 56
const TICK_COUNT = 4

function scaleFor(series: LinePoint[]) {
  const vals = series.map((p) => p.value)
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const pad = (max - min) * 0.18 || 1
  // Values that never go negative shouldn't grow a negative axis tick just from padding.
  const lo = min >= 0 ? Math.max(0, min - pad) : min - pad
  return { lo, hi: max + pad }
}

export function LineChart({
  series,
  color,
  height = 200,
  formatValue,
  secondarySeries,
  secondaryColor = 'var(--text-muted)',
  secondaryFormatValue,
}: LineChartProps) {
  const rawId = useId()
  const gradientId = `grad-${rawId.replace(/[^a-zA-Z0-9]/g, '')}`
  const H = height
  const [hoverIndex, setHoverIndex] = useState<number | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [W, setW] = useState(DEFAULT_W)

  // The SVG's viewBox is kept 1:1 with the actual rendered pixel width so
  // preserveAspectRatio="none" never has to stretch content (and glyphs)
  // horizontally to fill the container.
  useLayoutEffect(() => {
    const el = containerRef.current
    if (!el) return
    const update = () => {
      if (el.clientWidth) setW(el.clientWidth)
    }
    update()
    const obs = new ResizeObserver(update)
    obs.observe(el)
    return () => obs.disconnect()
  }, [])

  if (series.length === 0) {
    return (
      <div ref={containerRef} style={{ width: '100%', height: H }}>
        <svg viewBox={`0 0 ${W} ${H}`} width={W} height={H} style={{ display: 'block' }} />
      </div>
    )
  }

  const hasSecondary = !!secondarySeries && secondarySeries.length === series.length
  const { lo, hi } = scaleFor(series)

  const plotX0 = AXIS_W
  const plotW = W - AXIS_W - (hasSecondary ? AXIS_W_RIGHT : 0)
  const X = (i: number) => plotX0 + (series.length > 1 ? (i / (series.length - 1)) * plotW : 0)
  const Y = (v: number) => H - ((v - lo) / (hi - lo)) * H

  const pts = series.map((p, i): [number, number] => [X(i), Y(p.value)])
  const linePoints = pts.map((p) => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' ')
  const areaPoints = `${plotX0},${H} ${linePoints} ${plotX0 + plotW},${H}`

  const ticks = Array.from({ length: TICK_COUNT }, (_, i) => lo + (hi - lo) * (i / (TICK_COUNT - 1)))

  let sPts: [number, number][] = []
  let sLinePoints = ''
  let sTicks: number[] = []
  let sLo = 0
  let sHi = 0
  if (hasSecondary && secondarySeries) {
    const scale = scaleFor(secondarySeries)
    sLo = scale.lo
    sHi = scale.hi
    const Y2 = (v: number) => H - ((v - sLo) / (sHi - sLo)) * H
    sPts = secondarySeries.map((p, i): [number, number] => [X(i), Y2(p.value)])
    sLinePoints = sPts.map((p) => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' ')
    sTicks = Array.from({ length: TICK_COUNT }, (_, i) => sLo + (sHi - sLo) * (i / (TICK_COUNT - 1)))
  }

  const hovered = hoverIndex !== null ? { point: series[hoverIndex], xy: pts[hoverIndex] } : null
  const hoveredSecondary = hoverIndex !== null && hasSecondary && secondarySeries ? secondarySeries[hoverIndex] : null
  const tooltipLeftPct = hovered ? (hovered.xy[0] / W) * 100 : 0
  const tooltipAbove = hovered ? hovered.xy[1] > 40 : false

  return (
    <div ref={containerRef} style={{ position: 'relative', width: '100%', height: H }}>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        width={W}
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

        {hasSecondary &&
          secondaryFormatValue &&
          sTicks.map((t, i) => (
            <text
              key={`s-${i}`}
              x={plotX0 + plotW + 8}
              y={H - ((t - sLo) / (sHi - sLo)) * H}
              textAnchor="start"
              dominantBaseline={i === 0 ? 'auto' : i === TICK_COUNT - 1 ? 'hanging' : 'middle'}
              fontSize={10}
              fill="var(--text-faint)"
            >
              {secondaryFormatValue(t)}
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

        {hasSecondary && (
          <polyline
            points={sLinePoints}
            fill="none"
            stroke={secondaryColor}
            strokeWidth={2}
            strokeDasharray="5 4"
            vectorEffect="non-scaling-stroke"
            strokeLinejoin="round"
            strokeLinecap="round"
          />
        )}

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

        {hasSecondary &&
          sPts.map((p, i) => (
            <circle
              key={`s-${i}`}
              cx={p[0]}
              cy={p[1]}
              r={i === hoverIndex ? 5 : i === sPts.length - 1 ? 4 : 0}
              fill={secondaryColor}
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
          {hoveredSecondary && secondaryFormatValue && (
            <div className="line-chart-tooltip-value mono line-chart-tooltip-secondary">
              {secondaryFormatValue(hoveredSecondary.value)}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

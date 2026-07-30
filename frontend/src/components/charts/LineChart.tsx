import { useId, useLayoutEffect, useRef, useState, type PointerEvent } from 'react'

// Ported from the prototype's buildLine() (Portfolio App.dc.html ~line 690):
// polyline + gradient area fill, last point marked, ~18% vertical padding.
// Extended with a left Y-axis, a hover tooltip showing the point's value,
// and an optional secondary series (dashed) for comparing two metrics on
// the same timeline — either on its own right-side axis (two differently-
// scaled metrics, e.g. Rupiah vs a percentage) or sharing the primary's
// left axis (same unit, e.g. nominal vs inflation-adjusted Rupiah — see
// `sharedScale`).
export interface LinePoint {
  label: string
  value: number
}

interface LineChartProps {
  series: LinePoint[]
  color: string
  height?: number
  formatValue: (v: number) => string
  /** Axis-tick formatter, e.g. an abbreviated "rb/mn/B" style — falls back
   * to `formatValue` (used for the hover tooltip) when omitted. */
  axisFormatValue?: (v: number) => string
  secondarySeries?: LinePoint[]
  secondaryColor?: string
  secondaryFormatValue?: (v: number) => string
  secondaryAxisFormatValue?: (v: number) => string
  /** When true, the secondary series is plotted against the SAME left
   * axis as the primary (scaled from both series' combined range) instead
   * of getting its own right-side axis — for when both series are the
   * same unit and a second axis would just be visual noise. Ignored if
   * there's no secondary series. */
  sharedScale?: boolean
  /** An extra value shown in the tooltip only — never drawn as a line,
   * dot, or axis. For a related figure on a totally different scale (or
   * that isn't really "continuous" in spirit, e.g. a per-year cost that's
   * constant within a year and steps between years) where plotting it
   * alongside the main series would just be visual noise or misleading.
   * Snapped to the same point the tooltip's label snaps to — not
   * interpolated like the plotted series are, since a stepped/bucketed
   * value has nothing meaningful "between" two points. */
  tooltipExtra?: {
    label: string
    series: LinePoint[] // must be the same length as `series`
    formatValue: (v: number) => string
  }
}

const DEFAULT_W = 600
const AXIS_W = 56
const AXIS_W_RIGHT = 56
const TICK_COUNT = 4

function scaleFor(values: number[]) {
  const min = Math.min(...values)
  const max = Math.max(...values)
  const pad = (max - min) * 0.18 || 1
  // The axis never dips below zero — these charts track net worth/debt,
  // which don't read meaningfully as negative figures.
  const lo = Math.max(0, min - pad)
  return { lo, hi: max + pad }
}

/** Linear interpolation between series[i] and series[i+1] at fractional
 * index `frac` — lets hovering "between" two points estimate a value
 * rather than only ever snapping to an actual data point. */
function interpolate(series: LinePoint[], frac: number): number {
  const i0 = Math.floor(frac)
  const i1 = Math.min(series.length - 1, i0 + 1)
  const t = frac - i0
  return series[i0].value + (series[i1].value - series[i0].value) * t
}

export function LineChart({
  series,
  color,
  height = 200,
  formatValue,
  axisFormatValue = formatValue,
  secondarySeries,
  secondaryColor = 'var(--text-muted)',
  secondaryFormatValue,
  secondaryAxisFormatValue = secondaryFormatValue,
  sharedScale = false,
  tooltipExtra,
}: LineChartProps) {
  const rawId = useId()
  const gradientId = `grad-${rawId.replace(/[^a-zA-Z0-9]/g, '')}`
  const H = height
  const [hoverFrac, setHoverFrac] = useState<number | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)
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
  const showRightAxis = hasSecondary && !sharedScale
  const { lo, hi } =
    hasSecondary && sharedScale && secondarySeries
      ? scaleFor([...series, ...secondarySeries].map((p) => p.value))
      : scaleFor(series.map((p) => p.value))

  const plotX0 = AXIS_W
  const plotW = W - AXIS_W - (showRightAxis ? AXIS_W_RIGHT : 0)
  const X = (i: number) => plotX0 + (series.length > 1 ? (i / (series.length - 1)) * plotW : 0)
  const Y = (v: number) => H - ((v - lo) / (hi - lo)) * H

  const pts = series.map((p, i): [number, number] => [X(i), Y(p.value)])
  const linePoints = pts.map((p) => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' ')
  const areaPoints = `${plotX0},${H} ${linePoints} ${plotX0 + plotW},${H}`

  const ticks = Array.from({ length: TICK_COUNT }, (_, i) => lo + (hi - lo) * (i / (TICK_COUNT - 1)))

  let sPts: [number, number][] = []
  let sLinePoints = ''
  let sTicks: number[] = []
  let sLo = lo
  let sHi = hi
  if (hasSecondary && secondarySeries) {
    if (showRightAxis) {
      const scale = scaleFor(secondarySeries.map((p) => p.value))
      sLo = scale.lo
      sHi = scale.hi
    }
    const Y2 = showRightAxis ? (v: number) => H - ((v - sLo) / (sHi - sLo)) * H : Y
    sPts = secondarySeries.map((p, i): [number, number] => [X(i), Y2(p.value)])
    sLinePoints = sPts.map((p) => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' ')
    if (showRightAxis) {
      sTicks = Array.from({ length: TICK_COUNT }, (_, i) => sLo + (sHi - sLo) * (i / (TICK_COUNT - 1)))
    }
  }

  // Continuous hover: track the mouse across the ENTIRE plot area rather
  // than only within a small hit-circle right on top of each data point —
  // sparse data (e.g. yearly points across a wide chart) otherwise leaves
  // most of the chart with no hover response at all. The fractional index
  // this resolves to is also used to *interpolate* a value between the two
  // bracketing points, so the tooltip reads as a smooth estimate while
  // sweeping across, not just snapping from dot to dot.
  function handlePointerMove(e: PointerEvent<SVGRectElement>) {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const svgX = ((e.clientX - rect.left) / rect.width) * W
    const clampedX = Math.max(plotX0, Math.min(plotX0 + plotW, svgX))
    const frac = series.length > 1 ? ((clampedX - plotX0) / plotW) * (series.length - 1) : 0
    setHoverFrac(frac)
  }

  const hovered =
    hoverFrac !== null
      ? (() => {
          // The tooltip's label is categorical ("Year 7"), so it snaps to
          // whichever point the cursor is currently closer to rather than
          // trying to express a fractional label — tooltipExtra snaps the
          // same way, for the same reason.
          const snappedIndex = Math.min(series.length - 1, Math.round(hoverFrac))
          return {
            x: X(hoverFrac),
            value: interpolate(series, hoverFrac),
            secondaryValue: hasSecondary && secondarySeries ? interpolate(secondarySeries, hoverFrac) : null,
            label: series[snappedIndex].label,
            extraValue: tooltipExtra ? tooltipExtra.series[snappedIndex].value : null,
          }
        })()
      : null
  const hoveredY = hovered ? Y(hovered.value) : 0
  const hoveredSecondaryY =
    hovered && hovered.secondaryValue !== null
      ? showRightAxis
        ? H - ((hovered.secondaryValue - sLo) / (sHi - sLo)) * H
        : Y(hovered.secondaryValue)
      : 0
  const tooltipLeftPct = hovered ? (hovered.x / W) * 100 : 0
  const tooltipAbove = hovered ? hoveredY > 40 : false

  return (
    <div ref={containerRef} style={{ position: 'relative', width: '100%', height: H }}>
      <svg
        ref={svgRef}
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
            {axisFormatValue(t)}
          </text>
        ))}

        {showRightAxis &&
          secondaryAxisFormatValue &&
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
              {secondaryAxisFormatValue(t)}
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
            x1={hovered.x}
            y1={0}
            x2={hovered.x}
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
            r={i === pts.length - 1 ? 4.5 : 0}
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
              r={i === sPts.length - 1 ? 4 : 0}
              fill={secondaryColor}
              stroke="#fff"
              strokeWidth={2}
            />
          ))}

        {/* The moving "current estimate" marker(s), at the exact
            (possibly interpolated) hover position rather than snapped to
            a data point. */}
        {hovered && <circle cx={hovered.x} cy={hoveredY} r={5.5} fill={color} stroke="#fff" strokeWidth={2} />}
        {hovered && hovered.secondaryValue !== null && (
          <circle cx={hovered.x} cy={hoveredSecondaryY} r={5} fill={secondaryColor} stroke="#fff" strokeWidth={2} />
        )}

        {/* One large transparent hit-area covering the whole plot, drawn
            last so it's on top — this is what makes hovering work
            anywhere across the chart, not just precisely on a dot. */}
        <rect
          x={plotX0}
          y={0}
          width={plotW}
          height={H}
          fill="transparent"
          onPointerMove={handlePointerMove}
          onPointerLeave={() => setHoverFrac(null)}
        />
      </svg>

      {hovered && (
        <div
          className="line-chart-tooltip"
          style={{
            left: `${tooltipLeftPct}%`,
            top: tooltipAbove ? hoveredY - 10 : hoveredY + 10,
            transform: `translate(-50%, ${tooltipAbove ? '-100%' : '0'})`,
          }}
        >
          <div className="line-chart-tooltip-label">{hovered.label}</div>
          <div className="line-chart-tooltip-value mono">{formatValue(hovered.value)}</div>
          {hovered.secondaryValue !== null && secondaryFormatValue && (
            <div className="line-chart-tooltip-value mono line-chart-tooltip-secondary">
              {secondaryFormatValue(hovered.secondaryValue)}
            </div>
          )}
          {hovered.extraValue !== null && tooltipExtra && (
            <div className="line-chart-tooltip-extra-row">
              {tooltipExtra.label}: <span className="mono">{tooltipExtra.formatValue(hovered.extraValue)}</span>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

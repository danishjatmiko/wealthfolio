import { useState } from 'react'

// Ported from the prototype's buildDonut() (Portfolio App.dc.html ~line 682):
// SVG stroke-dasharray segments, one per category, colored by palette.
export interface DonutDatum {
  value: number
  color: string
  label?: string
}

interface DonutChartProps {
  data: DonutDatum[]
  size?: number
  thickness?: number
  onHover?: (datum: DonutDatum | null) => void
  /** Called when a segment is clicked. Optional — pages that only read the
   *  ring pass nothing, and the segments stay hover-only. */
  onSelect?: (datum: DonutDatum) => void
  /** Emphasises the slice with this label, for pages that drive selection
   *  from a table rather than the ring — matched by label rather than index
   *  because zero-value entries are filtered out below, so a caller's index
   *  wouldn't line up. Live hover still wins over selection. */
  selectedLabel?: string | null
}

export function DonutChart({
  data,
  size = 190,
  thickness = 26,
  onHover,
  onSelect,
  selectedLabel = null,
}: DonutChartProps) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)
  const total = data.reduce((s, d) => s + Math.max(0, d.value), 0) || 1
  const r = (size - thickness) / 2
  const C = 2 * Math.PI * r
  const cx = size / 2
  const cy = size / 2
  let acc = 0

  const visible = data.filter((d) => d.value > 0)

  function handleEnter(i: number, d: DonutDatum) {
    setHoverIdx(i)
    onHover?.(d)
  }
  function handleLeave() {
    setHoverIdx(null)
    onHover?.(null)
  }

  // Hovering anything overrides the selection, so the ring always tracks
  // the pointer when there is one.
  const selectedIdx =
    hoverIdx === null && selectedLabel != null
      ? visible.findIndex((d) => d.label === selectedLabel)
      : -1
  const activeIdx = hoverIdx ?? (selectedIdx === -1 ? null : selectedIdx)

  const segments = visible.map((d, i) => {
    const frac = d.value / total
    // Overshoot each segment's far edge by a hair — the next segment (or,
    // for the last one, the first) is drawn on top of it in DOM order and
    // covers the overlap, so this can't be seen. Without it, floating-point
    // rounding can leave a hairline gap at a seam, most visible where a
    // very thin slice sits next to a much larger one.
    const segLen = Math.min(C, frac * C + 0.75)
    const dasharray = `${segLen.toFixed(2)} ${Math.max(0, C - segLen).toFixed(2)}`
    const dashoffset = (-acc * C).toFixed(2)
    acc += frac
    const dimmed = activeIdx !== null && activeIdx !== i
    return (
      <circle
        key={i}
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke={d.color}
        strokeWidth={activeIdx === i ? thickness + 4 : thickness}
        strokeDasharray={dasharray}
        strokeDashoffset={dashoffset}
        transform={`rotate(-90 ${cx} ${cy})`}
        opacity={dimmed ? 0.35 : 1}
        style={{ cursor: 'pointer', transition: 'opacity 0.15s, stroke-width 0.15s' }}
        onMouseEnter={() => handleEnter(i, d)}
        onMouseLeave={handleLeave}
        onClick={onSelect ? () => onSelect(d) : undefined}
      />
    )
  })

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} style={{ display: 'block' }}>
      {segments}
    </svg>
  )
}

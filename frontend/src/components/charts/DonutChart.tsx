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
}

export function DonutChart({ data, size = 190, thickness = 26, onHover }: DonutChartProps) {
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

  const segments = visible.map((d, i) => {
    const frac = d.value / total
    const dasharray = `${(frac * C).toFixed(2)} ${(C - frac * C).toFixed(2)}`
    const dashoffset = (-acc * C).toFixed(2)
    acc += frac
    const dimmed = hoverIdx !== null && hoverIdx !== i
    return (
      <circle
        key={i}
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke={d.color}
        strokeWidth={hoverIdx === i ? thickness + 4 : thickness}
        strokeDasharray={dasharray}
        strokeDashoffset={dashoffset}
        transform={`rotate(-90 ${cx} ${cy})`}
        opacity={dimmed ? 0.35 : 1}
        style={{ cursor: 'pointer', transition: 'opacity 0.15s, stroke-width 0.15s' }}
        onMouseEnter={() => handleEnter(i, d)}
        onMouseLeave={handleLeave}
      />
    )
  })

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} style={{ display: 'block' }}>
      {segments}
    </svg>
  )
}

// Ported from the prototype's buildDonut() (Portfolio App.dc.html ~line 682):
// SVG stroke-dasharray segments, one per category, colored by palette.
export interface DonutDatum {
  value: number
  color: string
}

interface DonutChartProps {
  data: DonutDatum[]
  size?: number
  thickness?: number
}

export function DonutChart({ data, size = 190, thickness = 26 }: DonutChartProps) {
  const total = data.reduce((s, d) => s + Math.max(0, d.value), 0) || 1
  const r = (size - thickness) / 2
  const C = 2 * Math.PI * r
  const cx = size / 2
  const cy = size / 2
  let acc = 0

  const segments = data
    .filter((d) => d.value > 0)
    .map((d, i) => {
      const frac = d.value / total
      const dasharray = `${(frac * C).toFixed(2)} ${(C - frac * C).toFixed(2)}`
      const dashoffset = (-acc * C).toFixed(2)
      acc += frac
      return (
        <circle
          key={i}
          cx={cx}
          cy={cy}
          r={r}
          fill="none"
          stroke={d.color}
          strokeWidth={thickness}
          strokeDasharray={dasharray}
          strokeDashoffset={dashoffset}
          transform={`rotate(-90 ${cx} ${cy})`}
        />
      )
    })

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} style={{ display: 'block' }}>
      {segments}
    </svg>
  )
}

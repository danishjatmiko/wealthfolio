// Pure, side-effect-free asset growth/decline simulation. Monthly
// compounding: the annual growth rate is converted to its monthly
// equivalent (geometric, not divided by 12) so a 12%/year input actually
// compounds to 12%/year, not more. Monthly cost inflates once per
// completed year (a step function, not continuous compounding within a
// year) — a simple, defensible model of "cost of living rises every
// year," which is the whole reason inflation is an input that should
// actually move the balance, not just a cosmetic label.

export interface SimulationInput {
  /** Today's starting balance (e.g. current net worth), full/raw IDR. */
  startingBalanceIdr: number
  /** Annual nominal growth rate, as a percent (5..50). */
  annualRatePct: number
  /** Simulation horizon in whole years (3..25). */
  years: number
  /** Today's monthly cost, full/raw IDR — inflates each year. */
  monthlyCostIdr: number
  /** Annual inflation rate, as a percent (3..20). */
  annualInflationPct: number
  /** Extra monthly income added on top of growth, full/raw IDR. */
  monthlyIncomeIdr: number
}

export interface SimulationYearPoint {
  year: number // 0 = starting point ("Now"), 1..years = end of that year
  label: string
  nominalBalanceIdr: number
  /** Nominal balance deflated back to today's purchasing power. */
  realBalanceIdr: number
  /** Total cost paid during this specific year (the inflated monthly cost
   * × 12) — 0 for the "Now" point, since no time has elapsed yet. */
  annualCostIdr: number
}

export interface SimulationResult {
  points: SimulationYearPoint[]
  finalNominalBalanceIdr: number
  finalRealBalanceIdr: number
  totalIncomeAddedIdr: number
  totalCostPaidIdr: number
  /** 1-based month index the balance first went negative, or null if it never did. */
  depletedAtMonth: number | null
}

export function runSimulation(input: SimulationInput): SimulationResult {
  const monthlyGrowthRate = Math.pow(1 + input.annualRatePct / 100, 1 / 12) - 1
  const totalMonths = Math.round(input.years * 12)

  let balance = input.startingBalanceIdr
  let totalIncome = 0
  let totalCost = 0
  let costThisYear = 0
  let depletedAtMonth: number | null = null

  const points: SimulationYearPoint[] = [
    { year: 0, label: 'Now', nominalBalanceIdr: balance, realBalanceIdr: balance, annualCostIdr: 0 },
  ]

  for (let month = 1; month <= totalMonths; month++) {
    const yearIndex = Math.floor((month - 1) / 12) // 0-based year this month falls within
    const inflatedCost = input.monthlyCostIdr * Math.pow(1 + input.annualInflationPct / 100, yearIndex)

    balance = balance * (1 + monthlyGrowthRate) + input.monthlyIncomeIdr - inflatedCost
    totalIncome += input.monthlyIncomeIdr
    totalCost += inflatedCost
    costThisYear += inflatedCost

    if (depletedAtMonth === null && balance < 0) {
      depletedAtMonth = month
    }

    if (month % 12 === 0) {
      const yearNum = month / 12
      const realBalanceIdr = balance / Math.pow(1 + input.annualInflationPct / 100, yearNum)
      points.push({ year: yearNum, label: `Year ${yearNum}`, nominalBalanceIdr: balance, realBalanceIdr, annualCostIdr: costThisYear })
      costThisYear = 0
    }
  }

  const last = points[points.length - 1]
  return {
    points,
    finalNominalBalanceIdr: last.nominalBalanceIdr,
    finalRealBalanceIdr: last.realBalanceIdr,
    totalIncomeAddedIdr: totalIncome,
    totalCostPaidIdr: totalCost,
    depletedAtMonth,
  }
}

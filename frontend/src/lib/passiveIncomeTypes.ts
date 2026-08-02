// Curated income-type list for the Passive Income modal's dropdown —
// seeded from the type column of the user's original "Realized P/L"
// spreadsheet. Not a closed set: "Other" reveals a free-text field, same as
// BIG_EXPENSE_CATEGORIES and BOND_PLATFORMS.
//
// Orthogonal to the entry's category, which stays the asset class the
// income came from: a dividend and a realized capital gain are both Saham.
export const PASSIVE_INCOME_TYPES = [
  'Dividen',
  'Capital',
  'Bunga',
  'Cashback',
  'IPO',
  'Reksa Dana',
  'Sewa',
] as const

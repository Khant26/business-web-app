const bahtFormatter = new Intl.NumberFormat('th-TH', {
  style: 'currency',
  currency: 'THB',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

const bahtCompact = new Intl.NumberFormat('th-TH', {
  style: 'currency',
  currency: 'THB',
  minimumFractionDigits: 0,
  maximumFractionDigits: 2,
})

/** 129000 satang → "฿1,290.00" */
export function formatBaht(satang: number | null | undefined): string {
  return bahtFormatter.format((satang ?? 0) / 100)
}

/** 129000 satang → "฿1,290" (drops .00) */
export function formatBahtShort(satang: number | null | undefined): string {
  return bahtCompact.format((satang ?? 0) / 100)
}

/** Converts a baht amount (number or user-typed string like "1,290.50") to integer satang. */
export function bahtToSatang(baht: number | string): number {
  const n = typeof baht === 'number' ? baht : Number(String(baht).replace(/[,\s฿]/g, ''))
  if (!Number.isFinite(n)) return NaN
  return Math.round(n * 100)
}

/** 129050 → 1290.5 */
export function satangToBaht(satang: number | null | undefined): number {
  return (satang ?? 0) / 100
}

/** Percent off between price and compare-at price, or 0 when not discounted. */
export function percentOff(price: number, compareAt: number | null | undefined): number {
  if (!compareAt || compareAt <= price || compareAt <= 0) return 0
  return Math.round(((compareAt - price) / compareAt) * 100)
}

/** 700 basis points → "7%" */
export function formatBp(bp: number): string {
  const pct = bp / 100
  return `${Number.isInteger(pct) ? pct : pct.toFixed(2)}%`
}

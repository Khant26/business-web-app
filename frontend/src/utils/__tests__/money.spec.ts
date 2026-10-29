import { describe, expect, it } from 'vitest'
import {
  bahtToSatang,
  formatBaht,
  formatBahtShort,
  formatBp,
  percentOff,
  satangToBaht,
} from '../money'

describe('money', () => {
  it('formats satang as Thai baht', () => {
    expect(formatBaht(129000)).toBe('฿1,290.00')
    expect(formatBaht(0)).toBe('฿0.00')
    expect(formatBaht(5)).toBe('฿0.05')
    expect(formatBaht(123456789)).toBe('฿1,234,567.89')
    expect(formatBaht(null)).toBe('฿0.00')
  })

  it('formats short baht without trailing zeros', () => {
    expect(formatBahtShort(129000)).toBe('฿1,290')
    expect(formatBahtShort(129050)).toBe('฿1,290.5')
  })

  it('converts between baht and satang', () => {
    expect(bahtToSatang(1290.5)).toBe(129050)
    expect(bahtToSatang('1,290.50')).toBe(129050)
    expect(bahtToSatang('0.1')).toBe(10)
    expect(bahtToSatang('19.99')).toBe(1999)
    expect(Number.isNaN(bahtToSatang('abc'))).toBe(true)
    expect(satangToBaht(129050)).toBe(1290.5)
  })

  it('computes percent off', () => {
    expect(percentOff(8000, 10000)).toBe(20)
    expect(percentOff(10000, null)).toBe(0)
    expect(percentOff(10000, 9000)).toBe(0)
  })

  it('formats basis points', () => {
    expect(formatBp(700)).toBe('7%')
    expect(formatBp(725)).toBe('7.25%')
  })
})

import { describe, expect, it } from 'vitest'
import { isThaiPhone, normalizePhone, slugify } from '../misc'

describe('misc utils', () => {
  it('slugifies English names', () => {
    expect(slugify('Summer Sale 2026!')).toBe('summer-sale-2026')
    expect(slugify('  Café — Latte  ')).toBe('cafe-latte')
  })

  it('validates Thai phone numbers', () => {
    expect(isThaiPhone('081 234 5678')).toBe(true)
    expect(isThaiPhone('+66812345678')).toBe(true)
    expect(isThaiPhone('02-123-4567')).toBe(true)
    expect(isThaiPhone('12345')).toBe(false)
    expect(normalizePhone('+66 81-234-5678')).toBe('0812345678')
  })
})

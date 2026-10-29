import { describe, expect, it } from 'vitest'
import { pickLocalized } from '../localized'

describe('pickLocalized', () => {
  const p = { name_en: 'Green tea', name_my: 'လက်ဖက်ရည်စိမ်း' }

  it('returns English for en', () => {
    expect(pickLocalized(p, 'name', 'en')).toBe('Green tea')
  })

  it('returns Myanmar for my', () => {
    expect(pickLocalized(p, 'name', 'my')).toBe('လက်ဖက်ရည်စိမ်း')
  })

  it('falls back to English when Myanmar is empty or blank', () => {
    expect(pickLocalized({ name_en: 'Soap', name_my: '' }, 'name', 'my')).toBe('Soap')
    expect(pickLocalized({ name_en: 'Soap', name_my: '   ' }, 'name', 'my')).toBe('Soap')
    expect(pickLocalized({ title_en: 'Sale' }, 'title', 'my')).toBe('Sale')
  })

  it('handles missing objects', () => {
    expect(pickLocalized(null, 'name', 'en')).toBe('')
    expect(pickLocalized(undefined, 'title', 'my')).toBe('')
  })
})

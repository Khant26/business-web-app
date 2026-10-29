import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { CART_STORAGE_KEY, MAX_LINE_QTY, useCartStore } from '../cart'

vi.mock('@/api/catalog', () => ({
  catalogApi: {
    validateCart: vi.fn(async () => ({
      lines: [
        {
          variant_id: 1,
          product_slug: 'tea',
          product_name_en: 'Tea',
          product_name_my: '',
          variant_name_en: '500g',
          variant_name_my: '',
          image_url: '',
          unit_price_satang: 15000,
          quantity: 2,
          line_total_satang: 30000,
          available: 3,
        },
      ],
      problems: [{ variant_id: 2, code: 'out_of_stock', available: 0 }],
    })),
  },
}))

const flush = () => new Promise((r) => setTimeout(r, 0))

describe('cart store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('adds and merges lines for the same variant', () => {
    const cart = useCartStore()
    cart.add(1, 1)
    cart.add(1, 2)
    cart.add(2, 1)
    expect(cart.items).toHaveLength(2)
    expect(cart.find(1)?.quantity).toBe(3)
    expect(cart.count).toBe(4)
  })

  it('caps quantity and ignores non-positive adds', () => {
    const cart = useCartStore()
    cart.add(1, 500)
    expect(cart.find(1)?.quantity).toBe(MAX_LINE_QTY)
    cart.add(3, 0)
    expect(cart.find(3)).toBeUndefined()
  })

  it('sets quantity and removes when set to zero', () => {
    const cart = useCartStore()
    cart.add(1, 1)
    cart.setQuantity(1, 5)
    expect(cart.find(1)?.quantity).toBe(5)
    cart.setQuantity(1, 0)
    expect(cart.items).toHaveLength(0)
  })

  it('removes and clears', () => {
    const cart = useCartStore()
    cart.add(1, 1)
    cart.add(2, 1)
    cart.remove(1)
    expect(cart.items.map((i) => i.variant_id)).toEqual([2])
    cart.clear()
    expect(cart.count).toBe(0)
  })

  it('persists to localStorage and restores', async () => {
    const cart = useCartStore()
    cart.add(7, 2, {
      product_slug: 'x',
      product_name_en: 'X',
      product_name_my: '',
      variant_name_en: '',
      variant_name_my: '',
      image_url: '',
      unit_price_satang: 1000,
    })
    await flush()
    const saved = JSON.parse(localStorage.getItem(CART_STORAGE_KEY) ?? '[]')
    expect(saved[0]).toMatchObject({ variant_id: 7, quantity: 2 })

    setActivePinia(createPinia())
    const restored = useCartStore()
    expect(restored.find(7)?.quantity).toBe(2)
    expect(restored.subtotal).toBe(2000)
  })

  it('ignores corrupt storage', () => {
    localStorage.setItem(CART_STORAGE_KEY, '{not json')
    expect(useCartStore().items).toEqual([])
    setActivePinia(createPinia())
    localStorage.setItem(
      CART_STORAGE_KEY,
      JSON.stringify([
        { variant_id: 'a', quantity: 1 },
        { variant_id: 2, quantity: -1 },
      ]),
    )
    expect(useCartStore().items).toEqual([])
  })

  it('validate() refreshes line info and records problems', async () => {
    const cart = useCartStore()
    cart.add(1, 2)
    cart.add(2, 1)
    await cart.validate()
    expect(cart.find(1)?.info?.unit_price_satang).toBe(15000)
    expect(cart.subtotal).toBe(30000)
    expect(cart.problemFor(2)?.code).toBe('out_of_stock')
    expect(cart.hasProblems).toBe(true)
    expect(cart.payload).toEqual([
      { variant_id: 1, quantity: 2 },
      { variant_id: 2, quantity: 1 },
    ])
  })
})

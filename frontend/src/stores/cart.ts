import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { catalogApi } from '@/api/catalog'
import type { CartItemInput, CartProblem, QuoteLine } from '@/types/api'

export const CART_STORAGE_KEY = 'sawyun-cart'
export const MAX_LINE_QTY = 99

/** Display info cached alongside each line so the cart renders instantly before validation. */
export interface CartLineInfo {
  product_slug: string
  product_name_en: string
  product_name_my: string
  variant_name_en: string
  variant_name_my: string
  image_url: string
  unit_price_satang: number
  available?: number
}

export interface CartItem extends CartItemInput {
  info?: CartLineInfo
}

function load(): CartItem[] {
  try {
    const raw = localStorage.getItem(CART_STORAGE_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter(
        (x): x is CartItem =>
          typeof x === 'object' &&
          x !== null &&
          Number.isInteger((x as CartItem).variant_id) &&
          Number.isInteger((x as CartItem).quantity) &&
          (x as CartItem).quantity > 0,
      )
      .map((x) => ({
        variant_id: x.variant_id,
        quantity: Math.min(x.quantity, MAX_LINE_QTY),
        info: x.info,
      }))
  } catch {
    return []
  }
}

function infoFromLine(l: QuoteLine): CartLineInfo {
  return {
    product_slug: l.product_slug,
    product_name_en: l.product_name_en,
    product_name_my: l.product_name_my,
    variant_name_en: l.variant_name_en,
    variant_name_my: l.variant_name_my,
    image_url: l.image_url,
    unit_price_satang: l.unit_price_satang,
    available: l.available,
  }
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>(load())
  const problems = ref<CartProblem[]>([])
  const validating = ref(false)
  const validatedAt = ref<number | null>(null)

  const count = computed(() => items.value.reduce((n, i) => n + i.quantity, 0))
  const lineCount = computed(() => items.value.length)
  const subtotal = computed(() =>
    items.value.reduce((s, i) => s + (i.info?.unit_price_satang ?? 0) * i.quantity, 0),
  )
  const payload = computed<CartItemInput[]>(() =>
    items.value.map((i) => ({ variant_id: i.variant_id, quantity: i.quantity })),
  )
  const hasProblems = computed(() => problems.value.length > 0)

  function persist() {
    try {
      localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(items.value))
    } catch {
      /* quota / private mode */
    }
  }
  watch(items, persist, { deep: true })

  function find(variantId: number) {
    return items.value.find((i) => i.variant_id === variantId)
  }

  function problemFor(variantId: number): CartProblem | undefined {
    return problems.value.find((p) => p.variant_id === variantId)
  }

  /** Adds quantity to a variant, merging with an existing line. */
  function add(variantId: number, quantity = 1, info?: CartLineInfo) {
    if (quantity <= 0) return
    const existing = find(variantId)
    if (existing) {
      existing.quantity = Math.min(existing.quantity + quantity, MAX_LINE_QTY)
      if (info) existing.info = { ...existing.info, ...info }
    } else {
      items.value.push({ variant_id: variantId, quantity: Math.min(quantity, MAX_LINE_QTY), info })
    }
  }

  function setQuantity(variantId: number, quantity: number) {
    if (quantity <= 0) return remove(variantId)
    const existing = find(variantId)
    if (existing) existing.quantity = Math.min(Math.floor(quantity), MAX_LINE_QTY)
  }

  function remove(variantId: number) {
    items.value = items.value.filter((i) => i.variant_id !== variantId)
    problems.value = problems.value.filter((p) => p.variant_id !== variantId)
  }

  function clear() {
    items.value = []
    problems.value = []
  }

  /** Refreshes cached prices/names via POST /cart/validate and records problems. */
  async function validate(): Promise<void> {
    if (items.value.length === 0) {
      problems.value = []
      return
    }
    validating.value = true
    try {
      const res = await catalogApi.validateCart(payload.value)
      applyLines(res.lines)
      problems.value = res.problems
      validatedAt.value = Date.now()
    } finally {
      validating.value = false
    }
  }

  function applyLines(lines: QuoteLine[]) {
    for (const line of lines) {
      const item = find(line.variant_id)
      if (item) item.info = infoFromLine(line)
    }
  }

  // Keep tabs in sync.
  if (typeof window !== 'undefined') {
    window.addEventListener('storage', (e) => {
      if (e.key === CART_STORAGE_KEY) items.value = load()
    })
  }

  return {
    items,
    problems,
    validating,
    validatedAt,
    count,
    lineCount,
    subtotal,
    payload,
    hasProblems,
    find,
    problemFor,
    add,
    setQuantity,
    remove,
    clear,
    validate,
    applyLines,
  }
})

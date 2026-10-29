import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import ProductCard from '../store/ProductCard.vue'
import { i18n, setLocale } from '@/i18n'
import type { ProductCard as Card } from '@/types/api'

const MY_NAME = 'ဆန်မွှေး'
const MY_CAT = 'အစားအစာ'

const product: Card = {
  id: 1,
  slug: 'jasmine-rice',
  name_en: 'Jasmine Rice 5kg',
  name_my: MY_NAME,
  image_url: '',
  price_satang: 19900,
  compare_at_satang: 24900,
  rating_avg: 4.5,
  rating_count: 12,
  sold_count: 30,
  in_stock: true,
  featured: false,
  category: { slug: 'food', name_en: 'Food', name_my: MY_CAT },
}

function mountCard(p: Card = product) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/products/:slug', name: 'product', component: { template: '<div />' } },
      { path: '/login', name: 'login', component: { template: '<div />' } },
    ],
  })
  return mount(ProductCard, { props: { product: p }, global: { plugins: [i18n, router] } })
}

describe('ProductCard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('en')
  })

  it('renders name, price, discount badge and link', () => {
    const w = mountCard()
    expect(w.text()).toContain('Jasmine Rice 5kg')
    expect(w.text()).toContain('฿199.00')
    expect(w.text()).toContain('฿249.00')
    expect(w.text()).toContain('-20%')
    expect(w.find('a.link').attributes('href')).toBe('/products/jasmine-rice')
  })

  it('shows a placeholder initial when there is no image', () => {
    const w = mountCard()
    expect(w.find('img').exists()).toBe(false)
    expect(w.find('.letter').text()).toBe('J')
  })

  it('shows the out-of-stock badge', () => {
    const w = mountCard({ ...product, in_stock: false })
    expect(w.text()).toContain('Out of stock')
  })

  it('uses Myanmar names when the locale is my', () => {
    setLocale('my')
    const w = mountCard()
    expect(w.text()).toContain(MY_NAME)
    expect(w.text()).toContain(MY_CAT)
    expect(document.documentElement.lang).toBe('my')
    setLocale('en')
  })
})
